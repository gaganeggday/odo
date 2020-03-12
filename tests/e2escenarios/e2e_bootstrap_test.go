package e2escenarios

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var operatorYAML = `
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: pipelines
  namespace: openshift-operators
spec:
  channel: canary
  name: openshift-pipelines-operator
  source: community-operators
  sourceNamespace: openshift-marketplace
`

var _ = Describe("odo pipelines e2e tests", func() {
	var context string
	var project string
	var oc helper.OcRunner
	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		oc = helper.NewOcRunner("oc")
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		project = helper.CreateRandProject()
		oc.RunOcWithInput(strings.NewReader(operatorYAML), "apply", "-f", "-")
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	// Test Bootstrapping.
	Context("when bootstrapping pipelines", func() {
		It("creates the resources", func() {
			output := helper.CmdShouldPass("odo", "pipelines", "bootstrap", "--git-repo", "my-org/my-repo", "--prefix", "test",
				"--github-token", "abc123", "--image-repo", fmt.Sprintf("%s/taxi", project), "--skip-checks")

			setupOutput := oc.RunOcWithInput(strings.NewReader(output), "apply", "-f", "-")
			Expect(setupOutput).To(ContainSubstring("rolebinding.rbac.authorization.k8s.io/pipeline-edit-stage created"))
			oc.RunOcWithInput(strings.NewReader(output), "delete", "-f", "-")
		})

	})
})
