package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var amaltheaSessionGroupVersionResource = schema.GroupVersionResource{
	Group:    "amalthea.dev",
	Version:  "v1alpha1",
	Resource: "amaltheasessions",
}

func ListAmaltheaSessions(ctx context.Context, client *dynamic.DynamicClient, namespace string, gvr *schema.GroupVersionResource) (sessions []string, err error) {
	if gvr == nil {
		gvr = &amaltheaSessionGroupVersionResource
	}

	ams, err := client.Resource(*gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, session := range ams.Items {
		sessions = append(sessions, session.GetName())
	}
	return sessions, nil
}

func DeleteAmaltheaSession(ctx context.Context, client *dynamic.DynamicClient, namespace string, name string, gvr *schema.GroupVersionResource) error {
	if gvr == nil {
		gvr = &amaltheaSessionGroupVersionResource
	}

	propagation := metav1.DeletePropagationBackground
	err := client.Resource(*gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{PropagationPolicy: &propagation})
	if err != nil {
		return err
	}
	return nil
}

func DeleteAmaltheaSessions(ctx context.Context, client *dynamic.DynamicClient, namespace string, gvr *schema.GroupVersionResource) error {
	if gvr == nil {
		gvr = &amaltheaSessionGroupVersionResource
	}

	propagation := metav1.DeletePropagationBackground
	err := client.Resource(*gvr).Namespace(namespace).DeleteCollection(ctx, metav1.DeleteOptions{PropagationPolicy: &propagation}, metav1.ListOptions{})
	if err != nil {
		return err
	}
	return nil
}

func ForciblyDeleteAmaltheaSession(ctx context.Context, client *dynamic.DynamicClient, namespace string, name string, gvr *schema.GroupVersionResource) error {
	if gvr == nil {
		gvr = &amaltheaSessionGroupVersionResource
	}

	session, err := client.Resource(*gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	session.SetFinalizers(nil)
	_, err = client.Resource(*gvr).Namespace(namespace).Update(ctx, session, metav1.UpdateOptions{})
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

func ForciblyDeleteAmaltheaSessions(ctx context.Context, client *dynamic.DynamicClient, namespace string, gvr *schema.GroupVersionResource) error {
	servers, err := ListAmaltheaSessions(ctx, client, namespace, gvr)
	if err != nil {
		return err
	}

	for _, server := range servers {
		err = ForciblyDeleteAmaltheaSession(ctx, client, namespace, server, gvr)
		if err != nil {
			fmt.Printf("Ignoring error: %w\n", err)
		}
	}

	return nil
}
