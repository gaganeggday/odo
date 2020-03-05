package secrets

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/pkg/pipelines/meta"
	"github.com/openshift/odo/tests/helper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateOpaque(t *testing.T) {
	data := "abcdefghijklmnop"
	secret, err := CreateOpaque(meta.NamespacedName("cicd", "github-auth"), data)
	if err != nil {
		t.Fatal(err)
	}

	want := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "github-auth",
			Namespace: "cicd",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"token": []byte(data),
		},
	}

	if diff := cmp.Diff(want, secret); diff != "" {
		t.Fatalf("CreateOpaque() failed got\n%s", diff)
	}
}

func TestCreateDockerConfigWithErrorReading(t *testing.T) {
	testErr := errors.New("test failure")
	_, err := CreateDockerConfig(meta.NamespacedName("cici", "github-auth"), errorReader{testErr})
	if !helper.MatchErrorString(t, "failed to read .* test failure", err) {
		t.Fatalf("got an unexpected error: %#v", err)
	}
}

func TestCreateDockerConfig(t *testing.T) {
	data := []byte(`abcdefghijklmnop`)
	secret, err := CreateDockerConfig(meta.NamespacedName("cicd", "regcred"), bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	want := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "regcred",
			Namespace: "cicd",
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			".dockerconfigjson": data,
		},
	}

	if diff := cmp.Diff(want, secret); diff != "" {
		t.Fatalf("createDockerConfigSecret() failed got\n%s", diff)
	}
}

type errorReader struct {
	err error
}

func (e errorReader) Read(p []byte) (int, error) {
	return 0, e.err
}
