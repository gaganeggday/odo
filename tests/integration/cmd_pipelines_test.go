package integration

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/odo/tests/helper"
)

var _ = Describe("odo pipelines command tests", func() {

	// This is run after every Spec (It)
	var _ = BeforeEach(func() {
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultConsistentlyDuration(30 * time.Second)
		/*
			context = helper.CreateNewContext()
				os.Setenv("GLOBALODOCONFIG", filepath.Join(context, "config.yaml"))
				project = helper.CreateRandProject()
				originalDir = helper.Getwd()
		*/
	})

	// Clean up after the test
	// This is run after every Spec (It)
	var _ = AfterEach(func() {
		/*
			helper.DeleteProject(project)
			helper.DeleteDir(context)
			os.Unsetenv("GLOBALODOCONFIG")
		*/
	})
	/*
		Context("when running help for pipelines command", func() {
			It("should display the help", func() {
				help := helper.CmdShouldPass("odo", "pipelines", "-h")
				Expect(help).To(ContainSubstring("Performs application operations related to your OpenShift project."))
			})
		})
	*/

	Context("when running bootstrap ", func() {
		It("should display the help", func() {
			yaml := helper.CmdShouldPass("odo", "pipelines", "bootstrap", "--deployment-path", "deploy", "--dockerconfigjson",
				" ~/Downloads/wtam-robot-auth.json", "--github-token", "djfklajdflkas", "--quay-username", "--quay-username",
				"wtam2018", "--git-repo", "wtam2018/taxi", "--prefix", "wtam1")
			Expect(yaml).To(ContainSubstring("Performs application operations related to your OpenShift project."))
		})
	})
	/*
		Context("when running app delete, describe and list command on fresh cluster", func() {
			It("should error out display the help", func() {
				appList := helper.CmdShouldPass("odo", "pipelines", "list", "--project", project)
				Expect(appList).To(ContainSubstring("There are no applications deployed"))
				actual := helper.CmdShouldPass("odo", "app", "list", "-o", "json", "--project", project)
				desired := `{"kind":"List","apiVersion":"odo.openshift.io/v1alpha1","metadata":{},"items":[]}`
				Expect(desired).Should(MatchJSON(actual))

				appDelete := helper.CmdShouldFail("odo", "app", "delete", "test", "--project", project, "-f")
				Expect(appDelete).To(ContainSubstring("test app does not exists"))
				appDescribe := helper.CmdShouldPass("odo", "app", "describe", "test", "--project", project)
				Expect(appDescribe).To(ContainSubstring("Application test has no components or services deployed."))
			})
		})
	*/

})
