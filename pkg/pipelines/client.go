package pipelines

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
)

func getClientSet() (*kubernetes.Clientset, error) {
	clientConfig, err := getClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get client config due to %w", err)
	}
	clientSet, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get APIs client due to %w", err)
	}
	return clientSet, nil
}
