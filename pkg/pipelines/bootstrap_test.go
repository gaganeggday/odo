package pipelines

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestValidateImageRepo(t *testing.T) {
	tests := []struct {
		description string
		URL         string
		validURL    string
		isValid     bool
	}{
		{
			"Valid image regsitry URL",
			"quay.io/sample-user/sample-repo",
			"quay.io/sample-user/sample-repo",
			true,
		},
		{
			"Invalid image registry URL with missing repo name",
			"quay.io/sample-user",
			"quay.io/sample-user/sample-repo",
			false,
		},
		{
			"Invalid image registry URL with whitespaces",
			"quay.io/sample-user/ ",
			"quay.io/sample-user/sample-repo",
			false,
		},
	}
	for _, test := range tests {
		isValid, _ := validateImageRepo(test.URL)
		if diff := cmp.Diff(isValid, test.isValid); diff != "" {
			t.Fatalf("validateImageRepo() failed:\n%s", diff)
		}
	}
}
