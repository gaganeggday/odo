package project

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo pipelines command tests", func() {
	// This is run after every Spec (It)
	var context string
	var project string
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultConsistentlyDuration(30 * time.Second)
		context = helper.CreateNewContext()
		os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
		project = helper.CreateRandProject()
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		helper.DeleteProject(project)
		helper.DeleteDir(context)
		os.Unsetenv("GLOBALODOCONFIG")
	})

	Context("when running help for pipelines bootstrap command", func() {
		It("should display the help", func() {
			output := helper.CmdShouldPass("odo", "pipelines", "bootstrap", "--help")
			Expect(output).To(ContainSubstring("Bootstrap OpenShift pipelines in a cluster"))
		})
	})

	Context("when bootstrapping pipelines", func() {
		It("outputs valid yaml", func() {
			output := helper.CmdShouldPass("odo", "pipelines", "bootstrap", "--git-repo", "my-org/my-repo", "--prefix", "test",
				"--github-token", "abc123", "--image-repo", fmt.Sprintf("%s/taxi", project), "--skip-checks")

			var f interface{}
			err := yaml.Unmarshal([]byte(output), &f)
			Expect(err).To(BeNil())
		})
	})
})
