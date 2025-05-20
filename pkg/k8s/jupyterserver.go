package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var jupyterServerGroupVersionResource = schema.GroupVersionResource{
	Group:    "amalthea.dev",
	Version:  "v1alpha1",
	Resource: "jupyterservers",
}

func ListJupyterServers(ctx context.Context, client *dynamic.DynamicClient, namespace string, gvr *schema.GroupVersionResource) (servers []string, err error) {
	if gvr == nil {
		gvr = &jupyterServerGroupVersionResource
	}

	jss, err := client.Resource(*gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, js := range jss.Items {
		servers = append(servers, js.GetName())
	}
	return servers, nil
}

func DeleteJupyterServer(ctx context.Context, client *dynamic.DynamicClient, namespace string, name string, gvr *schema.GroupVersionResource) error {
	if gvr == nil {
		gvr = &jupyterServerGroupVersionResource
	}

	propagation := metav1.DeletePropagationBackground
	err := client.Resource(*gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &propagation})
	if err != nil {
		return err
	}
	return nil
}

func DeleteJupyterServers(ctx context.Context, client *dynamic.DynamicClient, namespace string, gvr *schema.GroupVersionResource) error {
	if gvr == nil {
		gvr = &jupyterServerGroupVersionResource
	}

	propagation := metav1.DeletePropagationBackground
	err := client.Resource(*gvr).Namespace(namespace).DeleteCollection(ctx, metav1.DeleteOptions{PropagationPolicy: &propagation}, metav1.ListOptions{})
	if err != nil {
		return err
	}
	return nil
}

func ForciblyDeleteJupyterServer(ctx context.Context, client *dynamic.DynamicClient, namespace string, name string, gvr *schema.GroupVersionResource) error {
	if gvr == nil {
		gvr = &jupyterServerGroupVersionResource
	}

	js, err := client.Resource(*gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	js.SetFinalizers(nil)
	_, err = client.Resource(*gvr).Namespace(namespace).Update(ctx, js, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	propagation := metav1.DeletePropagationForeground
	err = client.Resource(*gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &propagation})
	if err != nil {
		return err
	}
	return nil
}

func ForciblyDeleteJupyterServers(ctx context.Context, client *dynamic.DynamicClient, namespace string, gvr *schema.GroupVersionResource) error {
	servers, err := ListJupyterServers(ctx, client, namespace, gvr)
	if err != nil {
		return err
	}

	for _, server := range servers {
		err = ForciblyDeleteJupyterServer(ctx, client, namespace, server, gvr)
		if err != nil {
			fmt.Printf("Ignoring error: %s\n", err)
		}
	}

	return nil
}
