package pipelines

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestCompleteAddParameters(t *testing.T) {
	completeTests := []struct {
		name       string
		prefix     string
		wantPrefix string
	}{
		{"no prefix", "", ""},
		{"prefix with hyphen", "test-", "test-"},
		{"prefix without hyphen", "test", "test-"},
	}

	for _, tt := range completeTests {
		o := AddParameters{prefix: tt.prefix}

		err := o.Complete("test", &cobra.Command{}, []string{"test", "test/repo"})

		if err != nil {
			t.Errorf("Complete() %#v failed: ", err)
		}

		if o.prefix != tt.wantPrefix {
			t.Errorf("Complete() %#v prefix: got %s, want %s", tt.name, o.prefix, tt.wantPrefix)
		}
	}
}

func TestValidateAddParameters(t *testing.T) {
	optionTests := []struct {
		name      string
		gitRepo   string
		errMsgGit string
		appRepo   string
		errMsgApp string
	}{
		{"invalid repo", "test", "repo must be org/repo", "test2", "repo must be org/repo"},
		{"valid repo", "test/repo", "", "test2/repo", ""},
	}

	for _, tt := range optionTests {
		o := AddParameters{gitopsRepo: tt.gitRepo, prefix: "test", appGitRepo: tt.appRepo}

		err := o.Validate()

		if err != nil && tt.errMsgGit == "" {
			t.Errorf("Validate() %#v got an unexpected error: %s", tt.name, err)
			continue
		}

		if !matchError(t, tt.errMsgGit, err) {
			t.Errorf("Validate() %#v failed to match error: got %s, want %s", tt.name, err, tt.errMsgGit)
		}
		if !matchError(t, tt.errMsgApp, err) {
			t.Errorf("Validate() %#v failed to match error: got %s, want %s", tt.name, err, tt.errMsgApp)
		}
	}
}

func TestAddCommandWithMissingParams(t *testing.T) {
	cmdTests := []struct {
		desc    string
		flags   []keyValuePair
		wantErr string
	}{
		{"Missing app-git-repo flag",
			[]keyValuePair{flag("app-image-repo", "image/repo"),
				flag("app-webhook-secret", "123"), flag("env-name", "default"), flag("skip-checks", "true"), flag("gitops-repo", "gaganeggday/cide-pipelines"), flag("gitops-webhook-secret", "123"), flag("output", "."), flag("services-git-repo", "gaganeggday/service-1")},
			`Required flag(s) "app-git-repo" have/has not been set`},
		{"Missing app-image-repo flag",
			[]keyValuePair{flag("app-git-repo", "gaganeggday/cide-pipelines"),
				flag("app-webhook-secret", "123"), flag("env-name", "default"), flag("skip-checks", "true"), flag("gitops-repo", "gaganeggday/cide-pipelines"), flag("gitops-webhook-secret", "123"), flag("output", "."), flag("services-git-repo", "gaganeggday/service-1")},
			`Required flag(s) "app-image-repo" have/has not been set`},
		{"Missing app-webhook-secret flag",
			[]keyValuePair{flag("app-git-repo", "gaganeggday/cide-pipelines"),
				flag("app-image-repo", "example/image"), flag("env-name", "default"), flag("skip-checks", "true"), flag("gitops-repo", "gaganeggday/cide-pipelines"), flag("gitops-webhook-secret", "123"), flag("output", "."), flag("services-git-repo", "gaganeggday/service-1")},
			`Required flag(s) "app-webhook-secret" have/has not been set`},
		{"Missing env-name",
			[]keyValuePair{flag("app-git-repo", "gaganeggday/cide-pipelines"),
				flag("app-image-repo", "example/image"), flag("app-webhook-secret", "default"), flag("skip-checks", "true"), flag("gitops-repo", "gaganeggday/cide-pipelines"), flag("gitops-webhook-secret", "123"), flag("output", "."), flag("services-git-repo", "gaganeggday/service-1")},
			`Required flag(s) "env-name" have/has not been set`},
		{"Missing gitops-repo",
			[]keyValuePair{flag("app-git-repo", "gaganeggday/cide-pipelines"),
				flag("app-image-repo", "example/image"), flag("app-webhook-secret", "default"), flag("env-name", "default"), flag("skip-checks", "true"), flag("gitops-webhook-secret", "123"), flag("output", "."), flag("services-git-repo", "gaganeggday/service-1")},
			`Required flag(s) "gitops-repo" have/has not been set`},
		{"Missing gitops-webhook-secret",
			[]keyValuePair{flag("app-git-repo", "123"),
				flag("app-image-repo", "example/image"), flag("app-webhook-secret", "default"), flag("env-name", "default"), flag("skip-checks", "true"), flag("gitops-repo", "gaganeggday/cide-pipelines"), flag("output", "."), flag("services-git-repo", "gaganeggday/service-1")},
			`Required flag(s) "gitops-webhook-secret" have/has not been set`},
	}
	for _, tt := range cmdTests {
		t.Run(tt.desc, func(t *testing.T) {
			_, _, err := executeCommand(NewCmdAdd("add-service", "odo pipelines add-service"), tt.flags...)
			if err.Error() != tt.wantErr {
				t.Errorf("got %s, want %s", err, tt.wantErr)
			}
		})
	}
}

func TestPipelinesChecks(t *testing.T) {
	tests := []struct {
		description        string
		skipChecks         bool
		wantedBypassChecks bool
	}{
		{"bypass tekton installation checks", true, true},
		{"don't bypass tekton installation checks", false, false},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			o := AddParameters{skipChecks: test.skipChecks}

			err := o.Complete("test", &cobra.Command{}, []string{"test", "test/repo"})

			if err != nil {
				t.Errorf("Complete() %#v failed: ", err)
			}

			if o.skipChecks != test.wantedBypassChecks {
				t.Errorf("Complete() %#v bypassChecks flag: got %v, want %v", test.description, o.skipChecks, test.wantedBypassChecks)
			}
		})
	}

}
