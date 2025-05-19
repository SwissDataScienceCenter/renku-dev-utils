package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/github"
	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/helm"
	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/k8s"
	ns "github.com/SwissDataScienceCenter/renku-dev-utils/pkg/namespace"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var cleanupDeploymentCmd = &cobra.Command{
	Use:   "cleanup-deployment",
	Short: "Cleanup a renku deployment",
	Run:   cleanupDeployment,
}

func cleanupDeployment(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	if namespace == "" {
		cli, err := github.NewGitHubCLI("")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		namespace, err = ns.FindCurrentNamespace(ctx, cli)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	clients, err := k8s.GetClientset()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	client, err := k8s.GetDynamicClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Ask for confirmation
	fmt.Printf("This command will perform the following actions in the namespace '%s':\n", namespace)
	fmt.Println("  1. Delete all sessions")
	fmt.Println("  2. Uninstall all helm releases")
	fmt.Println("  3. Delete all jobs")
	fmt.Println("  4. Delete all PVCs")
	fmt.Println("  5. Forcibly delete all sessions")
	if deleteNamespace {
		fmt.Printf("  6. Delete the namespace '%s'\n", namespace)
	}
	proceed, err := askForConfirmation("Proceed?")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if !proceed {
		os.Exit(0)
	}

	// 1. Delete all sessions
	fmt.Println("1. Delete all sessions")
	err = k8s.DeleteAllSessions(ctx, client, namespace, k8s.DeleteAllSessionsOptions{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// 2. Uninstall all helm releases
	fmt.Println("2. Uninstall all helm releases")
	helm, err := helm.NewHelmCLI("")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = helm.UninstallAllReleases(ctx, namespace)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// 3. Delete all jobs
	fmt.Println("3. Delete all jobs")
	propagation := metav1.DeletePropagationForeground
	err = clients.BatchV1().Jobs(namespace).DeleteCollection(ctx, metav1.DeleteOptions{PropagationPolicy: &propagation}, metav1.ListOptions{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// 4. Delete all PVCs
	fmt.Println("4. Delete all PVCs")
	err = clients.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(ctx, metav1.DeleteOptions{PropagationPolicy: &propagation}, metav1.ListOptions{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// 5. Forcibly delete all sessions
	fmt.Println("5. Forcibly delete all sessions")
	err = k8s.DeleteAllSessions(ctx, client, namespace, k8s.DeleteAllSessionsOptions{})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// 6. Delete the namespace
	if deleteNamespace {
		fmt.Printf("6. Delete the namespace '%s'\n", namespace)
		err = clients.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{PropagationPolicy: &propagation})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func init() {
	cleanupDeploymentCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "k8s namespace")
	cleanupDeploymentCmd.Flags().BoolVar(&deleteNamespace, "delete-namespace", false, "if set, the namespace will be deleted")
}

func askForConfirmation(question string) (response bool, err error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s (yes/no): ", question)
	res, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	res = strings.ToLower(strings.TrimSpace(res))
	if res == "yes" {
		return true, nil
	} else if res == "no" {
		return false, nil
	}
	return false, fmt.Errorf("Invalid answer, aborting.")
}
