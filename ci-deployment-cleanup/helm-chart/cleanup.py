import subprocess, shlex
import shutil
import json
import re
import os
from dateutil import parser
from datetime import datetime, timedelta
from github import Github, Auth
from kubernetes import client, config
import logging

namespace_patterns_str = os.environ.get("NAMESPACE_PATTERNS", "")
NAMESPACE_REGEXES = namespace_patterns_str.split() if namespace_patterns_str else []

GITHUB_TOKEN = os.environ.get("GITHUB_TOKEN", "")
MAX_AGE_HOURS = int(os.environ.get("MAX_AGE_HOURS", "720"))
DRY_RUN = os.environ.get("DRY_RUN", "false").lower() == "true"

exemption_label_str = os.environ.get("EXEMPTION_LABEL", "")
if exemption_label_str and "=" in exemption_label_str:
    EXEMPTION_ANNOTATION = exemption_label_str.split("=", 1)[0]
else:
    EXEMPTION_ANNOTATION = "renku.io/cleanup-exempt"


logger = logging.getLogger(__name__)
logger.setLevel(logging.INFO)
console_handler = logging.StreamHandler()
console_handler.setLevel(logging.INFO)
formatter = logging.Formatter("%(asctime)s - %(name)s - %(levelname)s - %(message)s")
console_handler.setFormatter(formatter)
logger.addHandler(console_handler)

pr_repositories_str = os.environ.get("PR_REPOSITORIES", "")
NAMESPACE_PATTERN_TO_REPO_MAP = {}
if pr_repositories_str:
    for mapping in pr_repositories_str.split():
        if ":" in mapping:
            pattern, repo = mapping.split(":", 1)
            NAMESPACE_PATTERN_TO_REPO_MAP[pattern] = repo


class CIDeployment:
    def __init__(self, name, namespace, revision, updated, status, chart, app_version):
        self.name = name
        self.namespace = namespace
        self.revision = revision
        self.updated = updated
        self.status = status
        self.chart = chart
        self.app_version = app_version
        self.repo = None
        self.pr_number = None
        self.pr_is_open = None


class NamespaceChecker:
    def __init__(self):
        try:
            config.load_incluster_config()
        except config.ConfigException:
            config.load_kube_config()
        self.v1 = client.CoreV1Api()

    def is_namespace_exempt(self, namespace_name):
        try:
            namespace = self.v1.read_namespace(namespace_name)
            if namespace.metadata.annotations:
                exempt_value = namespace.metadata.annotations.get(EXEMPTION_ANNOTATION)
                return exempt_value == "true"
            return False
        except Exception as e:
            logger.error(
                f"Error checking namespace annotations for {namespace_name}: {e}"
            )
            return True


class GithubPRChecker:
    def __init__(self, github_token):
        self.g = Github(auth=Auth.Token(github_token))

    def is_pr_open(self, repo_name, pr_number):
        try:
            repo = self.g.get_repo(repo_name)
            pr = repo.get_pull(pr_number)
            return pr.state == "open"
        except Exception as e:
            logger.error(f"Error checking PR status for {repo_name}#{pr_number}: {e}")
            return True


class ShellExecution:
    def __init__(self, command):
        self.command = command

    def execute(self, dry_run=True):
        try:
            args = shlex.split(self.command)
            path = shutil.which(args[0])
            if path is None:
                raise FileNotFoundError(f"Command not found: {self.command.split()[0]}")
            else:
                args[0] = path

            logger.debug(f"Executing with resolved path: {args}")

            if dry_run:
                return "Dry run enabled. No action taken.", "", 0

            result = subprocess.run(
                args,
                timeout=900,
                encoding="utf-8",
                capture_output=True,
                check=False,
            )

            return result.stdout, result.stderr, result.returncode
        except subprocess.TimeoutExpired:
            return "", "Command timed out", -1
        except FileNotFoundError as e:
            return "", str(e), -1
        except Exception as e:
            return "", str(e), -1


class CIDeploymentsManager:
    def __init__(self):
        self.deployments = []

    def get_deployments(self):
        command = "helm list --all-namespaces -o json"
        shell_exec = ShellExecution(command)
        stdout, stderr, returncode = shell_exec.execute(dry_run=False)

        if returncode != 0:
            raise RuntimeError(
                f"helm command failed with return code {returncode}: {stderr}"
            )

        if not stdout:
            raise RuntimeError(f"helm command returned empty output. stderr: {stderr}")

        input_dict = json.loads(stdout)
        output_set = set()
        for ns_regex in NAMESPACE_REGEXES:
            output_dict = filter(
                lambda ns: re.match(ns_regex, ns["namespace"]), input_dict
            )
            for item in output_dict:
                last_activity = parser.parse(item["updated"][:19])
                item = CIDeployment(
                    name=item["name"],
                    namespace=item["namespace"],
                    revision=item["revision"],
                    updated=last_activity,
                    status=item["status"],
                    chart=item["chart"],
                    app_version=item["app_version"],
                )
                output_set.add(item)
        self.deployments = list(output_set)

    def filter_by_age(self, deployments, hours):
        threshold_time = datetime.now() - timedelta(hours=hours)
        return [dep for dep in deployments if dep.updated < threshold_time]

    def filter_by_closed_prs(self, deployments):
        pr_checker = GithubPRChecker(GITHUB_TOKEN)
        filtered = []
        for dep in deployments:
            if dep.repo and dep.pr_number:
                if not pr_checker.is_pr_open(dep.repo, int(dep.pr_number)):
                    dep.pr_is_open = False
                    filtered.append(dep)
                else:
                    dep.pr_is_open = True
            else:
                filtered.append(dep)
        return filtered

    def filter_exempt_namespaces(self, deployments):
        ns_checker = NamespaceChecker()
        filtered = []
        for dep in deployments:
            if ns_checker.is_namespace_exempt(dep.namespace):
                logger.info(f"Skipping exempt namespace: {dep.namespace}")
            else:
                filtered.append(dep)
        return filtered

    def get_deletable_deployments(self, max_age_hours):
        old = self.filter_by_age(self.deployments, max_age_hours)
        closed_pr = self.filter_by_closed_prs(self.deployments)
        candidates = list(set(old).union(set(closed_pr)))
        return self.filter_exempt_namespaces(candidates)

    def print_deployments(self, deployments):
        for dep in deployments:
            logger.debug(f"\nName: {dep.name}")
            logger.debug(f"  Namespace: {dep.namespace}")
            logger.debug(f"  Updated: {dep.updated}")
            logger.debug(f"  Repo: {dep.repo}")
            logger.debug(f"  PR: {dep.pr_number}")
            logger.debug(f"  PR Open: {dep.pr_is_open}")

    def exclude_deployments(self, names_to_exclude):
        self.deployments = [
            dep for dep in self.deployments if dep.name not in names_to_exclude
        ]

    def match_namespaces_to_repos(self):
        for dep in self.deployments:
            for pattern, repo in NAMESPACE_PATTERN_TO_REPO_MAP.items():
                if re.match(pattern, dep.namespace):
                    dep.repo = repo
                    break

    def assign_pr_numbers(self):
        for dep in self.deployments:
            potential_pr = dep.namespace.split("-")[-1]
            try:
                pr_num = int(potential_pr)
                dep.pr_number = pr_num
            except ValueError:
                logger.info(
                    f"Warning: Could not parse PR number from namespace {dep.namespace}, skipping PR assignment"
                )
                dep.pr_number = None

    def run_cleanup(self, max_age_hours=None, dry_run=None):
        if max_age_hours is None:
            max_age_hours = MAX_AGE_HOURS
        if dry_run is None:
            dry_run = DRY_RUN

        logger.debug(
            f"Starting cleanup with max_age_hours={max_age_hours}, dry_run={dry_run}"
        )
        if dry_run:
            logger.info("DRY RUN MODE: No actual deletions will be performed")

        logger.debug("Getting CI deployments")
        self.get_deployments()
        logger.debug(f"Found {len(self.deployments)} CI deployments")
        self.match_namespaces_to_repos()
        self.assign_pr_numbers()

        logger.debug("Determining deletable CI deployments")
        deployments_to_delete = self.get_deletable_deployments(max_age_hours)

        logger.info(f"Total CI deployments to delete: {len(deployments_to_delete)}")
        self.print_deployments(deployments=deployments_to_delete)

        successful_deletions = []
        failed_deletions = []

        for deployment in deployments_to_delete:
            remover = CIDeploymentRemover(deployment, dry_run=dry_run)
            stdout, stderr, returncode = remover.remove_with_rdu()

            if returncode == 0:
                successful_deletions.append(deployment.namespace)
            else:
                failed_deletions.append((deployment.namespace, returncode, stderr))

        self.print_summary(
            deployments_to_delete, successful_deletions, failed_deletions
        )

        return successful_deletions, failed_deletions

    def print_summary(self, all_deployments, successful, failed):
        logger.info("=" * 80)
        logger.info("CLEANUP SUMMARY")
        logger.info("=" * 80)
        logger.info(f"Total CI deployments processed: {len(all_deployments)}")
        logger.info(f"Successful deletions: {len(successful)}")
        logger.info(f"Failed deletions: {len(failed)}")

        if failed:
            logger.error("Failed namespaces:")
            for namespace, returncode, stderr in failed:
                logger.error(f"  - {namespace} (exit code: {returncode})")
                if stderr:
                    logger.error(f"    Error: {stderr[:200]}")


class CIDeploymentRemover:
    def __init__(self, deployment, dry_run=True):
        self.deployment = deployment
        self.dry_run = dry_run

    def remove(self):
        self.remove_with_rdu()

    def remove_with_rdu(self):
        command = f"rdu cleanup-deployment --namespace {self.deployment.namespace} --delete-namespace --yes"
        logger.info(
            f"\n{'[DRY RUN] ' if self.dry_run else ''}Deleting namespace: {self.deployment.namespace}"
        )
        logger.debug(f"  Updated: {self.deployment.updated}")
        logger.debug(f"  Repo: {self.deployment.repo}")
        logger.debug(
            f"  PR: {self.deployment.pr_number} (Open: {self.deployment.pr_is_open})"
        )

        if self.dry_run:
            logger.info(f"  Command: {command}")
            return "Dry run enabled. No action taken.", "", 0
        else:
            logger.debug(f"  Executing: {command}")
            shell_exec = ShellExecution(command)
            stdout, stderr, returncode = shell_exec.execute(dry_run=False)

            if returncode == 0:
                logger.info(
                    f"  ✓ Successfully deleted namespace: {self.deployment.namespace}"
                )
            else:
                logger.error(
                    f"  ✗ Failed to delete namespace: {self.deployment.namespace}"
                )
                logger.debug(f"    Return code: {returncode}")
                if stderr:
                    logger.error(f"    Error output: {stderr}")
                if stdout:
                    logger.debug(f"    Standard output: {stdout}")

            return stdout, stderr, returncode


if __name__ == "__main__":
    if not GITHUB_TOKEN:
        logger.error("ERROR: GITHUB_TOKEN environment variable is required but not set")
        exit(1)

    logger.info(f"Environment: MAX_AGE_HOURS={MAX_AGE_HOURS}, DRY_RUN={DRY_RUN}")

    manager = CIDeploymentsManager()
    manager.run_cleanup()
