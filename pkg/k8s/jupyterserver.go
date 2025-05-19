package k8s

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var jupyterServerGroupVersionResource = schema.GroupVersionResource{
	Group:    "amalthea.dev",
	Version:  "v1alpha1",
	Resource: "jupyterservers",
}

func ListJS(ctx context.Context, client *dynamic.DynamicClient, gvr *schema.GroupVersionResource) error {
	if gvr == nil {
		gvr = &jupyterServerGroupVersionResource
	}

	jss, err := client.Resource(*gvr).List(ctx, v1.ListOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Hey: %s\n", jss)

	return nil
}
