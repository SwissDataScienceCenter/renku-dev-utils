# Renku CI Deployment Cleanup

A Kubernetes-based CI deployment cleanup system that uses a Helm chart to deploy automated cleanup of old Renku CI deployments. This system runs as a CronJob that leverages the `rdu` tool for comprehensive cleanup.

## Installation

Install the Helm chart:
```bash
helm install renku-ci-cleanup ./helm-chart
```

## Exemption

Namespaces can be exempted from cleanup by adding the label `renku.io/cleanup-exempt: "true"` to the namespace.

## How It Works

1. The CronJob runs on the specified schedule (default: every 6 hours)
2. It queries Kubernetes for ALL namespaces in the cluster
3. For each namespace found:
   - Checks if the namespace has the exemption label (if so, skips it)
   - Checks if the namespace name matches any of the configured patterns (if enforcement is enabled)
   - Calculates the age based on the namespace creation timestamp
   - Checks GitHub PR status for PR-based cleanup (if enabled)
   - If the namespace is older than the configured threshold AND matches the naming patterns AND is not exempt, it uses `rdu cleanup-deployment` to:
     - Delete all sessions
     - Uninstall all Helm releases
     - Delete all jobs and PVCs
     - Delete the entire namespace
4. Logging shows what actions were taken, including exemption and pattern matching results

## Key Configuration

The main configuration options in `values.yaml`:

- `cleanup.maxAge`: Maximum age in hours before cleanup (default: 720 hours / 30 days)
- `cleanup.dryRun`: Enable dry-run mode (default: false)
- `cleanup.namespacePatterns`: List of regex patterns for namespace names
- `cleanup.enforceNamePatterns`: Enable strict pattern matching (default: true)
- `cleanup.prCleanup.enabled`: Enable GitHub PR-based cleanup (default: false)
- `cronJob.schedule`: Cron schedule (default: "0 */6 * * *" - every 6 hours)

## PR-Based Cleanup

The system supports GitHub PR-based cleanup that can automatically clean up namespaces when their associated pull requests are closed or merged. This feature requires:

- `cleanup.prCleanup.enabled: true`
- GitHub API token configured
- Repository mappings in `cleanup.prCleanup.repositories`

Example configuration maps namespace patterns to GitHub repositories and PR numbers.
