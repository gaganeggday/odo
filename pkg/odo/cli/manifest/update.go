package manifest

import (
	"fmt"

	"github.com/openshift/odo/pkg/manifest"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubernetes/pkg/kubectl/util/templates"
)

const (
	// UpdateRecommendedCommandName the recommended command name
	UpdateRecommendedCommandName = "update"
)

var (
	updateExample = ktemplates.Examples(`
	# Update OpenShift GitOps manifest
	%[1]s 
	`)

	updateLongDesc  = ktemplates.LongDesc(`Update GitOps manifest`)
	updateShortDesc = `Update manifest`
)

// UpdateParameters encapsulates the parameters for the odo manifest update command.
type UpdateParameters struct {
	output string // path to add Gitops manifest file
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewUpdateParameters bootstraps a UpdateParameters instance.
func NewUpdateParameters() *UpdateParameters {
	return &UpdateParameters{}
}

// Complete completes InitParameters after they've been created.
//
// If the prefix provided doesn't have a "-" then one is added, this makes the
// generated environment names nicer to read.
func (io *UpdateParameters) Complete(name string, cmd *cobra.Command, args []string) error {
	return nil
}

// Validate validates the parameters of the InitParameters.
func (io *UpdateParameters) Validate() error {
	return nil
}

// Run runs the project update command.
func (io *UpdateParameters) Run() error {
	options := manifest.UpdateParameters{
		Output: io.output,
	}
	return manifest.Update(&options)
}

// NewCmdUpdate creates the project update command.
func NewCmdUpdate(name, fullName string) *cobra.Command {
	o := NewUpdateParameters()

	updateCmd := &cobra.Command{
		Use:     name,
		Short:   updateShortDesc,
		Long:    updateLongDesc,
		Example: fmt.Sprintf(updateExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	updateCmd.Flags().StringVar(&o.output, "output", ".", "folder path to add GitOps resources")
	updateCmd.MarkFlagRequired("output")
	return updateCmd
}
