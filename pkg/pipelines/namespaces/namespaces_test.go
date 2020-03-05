package namespaces

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func TestCreateNamespace(t *testing.T) {
	ns := Create("test-environment")
	want := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-environment",
		},
	}

	if diff := cmp.Diff(want, ns); diff != "" {
		t.Fatalf("Create() failed got\n%s", diff)
	}
}

func TestPrefixed(t *testing.T) {
	ns := Prefixed("test-")
	want := map[string]string{
		"dev":   "test-dev-environment",
		"stage": "test-stage-environment",
		"cicd":  "test-cicd-environment",
	}
	if diff := cmp.Diff(want, ns); diff != "" {
		t.Fatalf("namespaceNames() failed got\n%s", diff)
	}
}

func TestExists(t *testing.T) {
	tests := []struct {
		desc      string
		namespace string
		valid     bool
	}{
		{
			"namespace already exists",
			"sample",
			true,
		},
		{
			"namespace doesn't exist",
			"test",
			false,
		},
	}
	validNamespace := Create("sample")
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			cs := testclient.NewSimpleClientset(validNamespace)
			namespaceExists, _ := Exists(cs, test.namespace)
			if diff := cmp.Diff(namespaceExists, test.valid); diff != "" {
				t.Fatalf("checkNamespace() failed:\n%v", diff)
			}
		})
	}
}
