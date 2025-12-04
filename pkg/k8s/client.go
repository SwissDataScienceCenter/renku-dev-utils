package k8s

import (
	"fmt"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func GetClientset() (*kubernetes.Clientset, error) {
	home := homedir.HomeDir()
	if home == "" {
		return nil, fmt.Errorf("could not determine home directory")
	}

	kubeconfig := filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func GetDynamicClient() (client *dynamic.DynamicClient, err error) {
	home := homedir.HomeDir()
	if home == "" {
		return nil, fmt.Errorf("could not determine home directory")
	}

	kubeconfig := filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	return dynamic.NewForConfig(config)
}
