package k8s

import (
	"fmt"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	config, err = kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
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
