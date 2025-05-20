package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"runtime"

	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/executils"
	"github.com/SwissDataScienceCenter/renku-dev-utils/pkg/github"
	ns "github.com/SwissDataScienceCenter/renku-dev-utils/pkg/namespace"
	"github.com/spf13/cobra"
)

var openDeploymentCmd = &cobra.Command{
	Use:   "open",
	Short: "Open a renku deployment in the browser",
	Run:   openDeployment,
}

func openDeployment(cmd *cobra.Command, args []string) {
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

	// TODO: Can we derive the URL by inspecting ingresses in the k8s namespace?
	openURL, err := url.Parse(fmt.Sprintf("https://%s.dev.renku.ch", namespace))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	openURLStr := openURL.String()
	fmt.Printf("Open URL: %s\n", openURLStr)

	if runtime.GOOS == "darwin" {
		cmd := exec.CommandContext(ctx, "open", openURLStr)
		_, err = executils.FormatOutput(cmd.Output())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	if runtime.GOOS == "linux" {
		cmd := exec.CommandContext(ctx, "xdg-open", openURLStr)
		_, err = executils.FormatOutput(cmd.Output())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	fmt.Printf("Sorry, I do not know how to \"open\" on '%s'\n", runtime.GOOS)
	os.Exit(1)
}

func init() {
	openDeploymentCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "k8s namespace")
}
