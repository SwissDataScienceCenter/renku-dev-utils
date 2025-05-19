package k8s

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type DeleteAllSessionsOptions struct {
	JupyterServerGvr *schema.GroupVersionResource
}

func DeleteAllSessions(ctx context.Context, client *dynamic.DynamicClient, namespace string, opts DeleteAllSessionsOptions) error {
	err := DeleteJupyterServers(ctx, client, namespace, opts.JupyterServerGvr)
	if err != nil {
		return err
	}

	err = DeleteAmaltheaSessions(ctx, client, namespace, opts.JupyterServerGvr)
	if err != nil {
		return err
	}

	return nil
}

func ForciblyDeleteAllSessions(ctx context.Context, client *dynamic.DynamicClient, namespace string, opts DeleteAllSessionsOptions) error {
	err := ForciblyDeleteJupyterServers(ctx, client, namespace, opts.JupyterServerGvr)
	if err != nil {
		return err
	}

	err = ForciblyDeleteAmaltheaSessions(ctx, client, namespace, opts.JupyterServerGvr)
	if err != nil {
		return err
	}

	return nil
}
