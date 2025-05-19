package k8s

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type DeleteAllSessionsOptions struct {
	AmaltheaSessionGvr *schema.GroupVersionResource
	JupyterServerGvr   *schema.GroupVersionResource
}

func DeleteAllSessions(ctx context.Context, client *dynamic.DynamicClient, namespace string, opts DeleteAllSessionsOptions) error {
	err := DeleteAmaltheaSessions(ctx, client, namespace, opts.AmaltheaSessionGvr)
	if err != nil {
		return err
	}

	err = DeleteJupyterServers(ctx, client, namespace, opts.JupyterServerGvr)
	if err != nil {
		return err
	}

	return nil
}

func ForciblyDeleteAllSessions(ctx context.Context, client *dynamic.DynamicClient, namespace string, opts DeleteAllSessionsOptions) error {
	err := ForciblyDeleteAmaltheaSessions(ctx, client, namespace, opts.AmaltheaSessionGvr)
	if err != nil {
		return err
	}

	err = ForciblyDeleteJupyterServers(ctx, client, namespace, opts.JupyterServerGvr)
	if err != nil {
		return err
	}

	return nil
}
