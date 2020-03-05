package secrets

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/openshift/odo/pkg/pipelines/meta"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// CreateOpaque creates a Kubernetes v1/Secret with the provided name and
// body, and type Opaque.
func CreateOpaque(name types.NamespacedName, data string) (*corev1.Secret, error) {
	r := strings.NewReader(data)

	return createSecret(name, "token", corev1.SecretTypeOpaque, r)
}

// CreateDockerConfig creates a Kubernetes v1/Secret with the provided name and
// body, and type DockerConfigJson.
func CreateDockerConfig(name types.NamespacedName, in io.Reader) (*corev1.Secret, error) {
	return createSecret(name, ".dockerconfigjson", corev1.SecretTypeDockerConfigJson, in)
}

func createSecret(name types.NamespacedName, key string, st corev1.SecretType, in io.Reader) (*corev1.Secret, error) {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret data: %w", err)
	}
	secret := &corev1.Secret{
		TypeMeta:   meta.TypeMeta("Secret", "v1"),
		ObjectMeta: meta.ObjectMeta(name),
		Type:       st,
		Data: map[string][]byte{
			key: data,
		},
	}
	return secret, nil
}
