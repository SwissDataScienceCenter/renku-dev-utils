package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetSecret(ctx context.Context, clients *kubernetes.Clientset, namespace string, secretName string) (secret *corev1.Secret, err error) {
	return clients.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
}
