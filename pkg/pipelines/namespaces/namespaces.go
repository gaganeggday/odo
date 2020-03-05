package namespaces

import (
	"fmt"

	"github.com/openshift/odo/pkg/pipelines/meta"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var namespaceBaseNames = map[string]string{
	"dev":   "dev-environment",
	"stage": "stage-environment",
	"cicd":  "cicd-environment",
}

// Prefixed creates and returns a map of named environments to their prefixed
// namespaces.
func Prefixed(prefix string) map[string]string {
	prefixedNames := make(map[string]string)
	for k, v := range namespaceBaseNames {
		prefixedNames[k] = fmt.Sprintf("%s%s", prefix, v)
	}
	return prefixedNames
}

// Create creates and returns a corev1.Namespace.
func Create(name string) *corev1.Namespace {
	ns := &corev1.Namespace{
		TypeMeta: meta.TypeMeta("Namespace", "v1"),
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return ns
}

// Exists returns true if the named namespace already exists in Kubernetes.
func Exists(clientSet kubernetes.Interface, name string) (bool, error) {
	_, err := clientSet.CoreV1().Namespaces().Get(name, metav1.GetOptions{})
	if err != nil {
		return false, nil
	}
	return true, nil
}
