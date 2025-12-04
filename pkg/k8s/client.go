package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func getConfig() (*rest.Config, error) {
	// Try in-cluster config first (for running inside K8s pods)
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Check KUBECONFIG environment variable
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, err
		}
		return config, nil
	}

	// Fall back to default kubeconfig location
	home := homedir.HomeDir()
	if home == "" {
		return nil, fmt.Errorf("could not determine home directory")
	}

	kubeconfigPath = filepath.Join(home, ".kube", "config")
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func GetClientset() (*kubernetes.Clientset, error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func GetDynamicClient() (client *dynamic.DynamicClient, err error) {
	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(config)
}
