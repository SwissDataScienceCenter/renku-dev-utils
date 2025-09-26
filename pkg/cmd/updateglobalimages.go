package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/github"
	ns "github.com/SwissDataScienceCenter/renku-dev-utils/pkg/namespace"
	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/renkuapi"
	"github.com/spf13/cobra"
)

var updateGlobalImagesCmd = &cobra.Command{
	Use:   "update-global-images",
	Short: "Updates the global images",
	Run:   updateGlobalImages,
}

func updateGlobalImages(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	release, err := cmd.Flags().GetString("release")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if release == "" {
		cli, err := github.NewGitHubCLI("")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		release, err = cli.GetLatestRenkuRelease(ctx)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	fmt.Printf("Renku release: %s\n", release)

	url, err := cmd.Flags().GetString("url")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if url == "" {
		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
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

		deploymentURL, err := ns.GetDeploymentURL(namespace)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		url = deploymentURL.String()
	}

	fmt.Printf("Renku URL: %s\n", url)

	rac, err := renkuapi.NewRenkuApiClient(url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !rac.IsLoggedIn(ctx) {
		fmt.Println("Error: not logged in")
		showCmd := "rdu login"
		if url != "" {
			showCmd = showCmd + fmt.Sprintf(" --url %s", url)
		}
		if namespace != "" {
			showCmd = showCmd + fmt.Sprintf(" --namespace %s", namespace)
		}
		fmt.Printf("Please run \"%s\" before running this command\n", showCmd)
		os.Exit(1)
	}

	if !rac.IsAdmin(ctx) {
		fmt.Println("Error: not an admin")
		fmt.Println("Please make sure you are a Renku admin before running this command")
		fmt.Println("See: rdu make-me-admin --help")
		os.Exit(1)
	}

	rsc, err := rac.Session()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	envs, err := rsc.GetGlobalEnvironments(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = rsc.UpdateGlobalImages(ctx, github.GetGlobalImages(), release, envs, dryRun)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	updateGlobalImagesCmd.Flags().String("url", "", "instance URL")
	updateGlobalImagesCmd.Flags().StringP("namespace", "n", "", "k8s namespace")
	updateGlobalImagesCmd.Flags().String("release", "", "renku release")
	updateGlobalImagesCmd.Flags().Bool("dry-run", false, "dry run")
}
