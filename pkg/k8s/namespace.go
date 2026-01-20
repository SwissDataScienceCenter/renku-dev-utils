package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListNamespaces(ctx context.Context, clients *kubernetes.Clientset) (namespaceList *corev1.NamespaceList, err error) {
	return clients.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
}
