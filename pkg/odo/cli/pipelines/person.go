package pipelines

import (
	"fmt"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	// PersonRecommendedCommandName the recommended command name
	PersonRecommendedCommandName = "person"
)

var (
	personExample = ktemplates.Examples(`
	# Initialize OpenShift GitOps pipelines
	%[1]s 
	`)

	personLongDesc  = ktemplates.LongDesc(`Initialize GitOps pipelines`)
	personShortDesc = `Initialize pipelines`
)

// PersonParameters encapsulates the parameters for the odo pipelines init command.
type PersonParameters struct {
	Name    string
	Gender  string
	Married bool
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewPersonParameters bootstraps a InitParameters instance.
func NewPersonParameters() *PersonParameters {
	return &PersonParameters{}
}

// Complete completes InitParameters after they've been created.
//
// If the prefix provided doesn't have a "-" then one is added, this makes the
// generated environment names nicer to read.
func (io *PersonParameters) Complete(name string, cmd *cobra.Command, args []string) error {
	return nil
}

// Validate validates the parameters of the InitParameters.
func (io *PersonParameters) Validate() error {
	return nil
}

// Run runs the project bootstrap command.
func (io *PersonParameters) Run() error {
	// options := pipelines.PersonParameters{
	// 	Name: io.Name,
	// }
	// err := pipelines.Init(&options, ioutils.NewFilesystem())
	// if err != nil {
	// 	return err
	// }
	log.Successf("Intialized GitOps sucessfully.")
	return nil
}

// NewCmdPerson creates the project init command.
func NewCmdPerson(name, fullName string) *cobra.Command {
	o := NewPersonParameters()

	personCmd := &cobra.Command{
		Use:     name,
		Short:   initShortDesc,
		Long:    initLongDesc,
		Example: fmt.Sprintf(initExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	personCmd.Flags().StringVar(&o.Name, "Name", "", "Returns the name")
	personCmd.MarkFlagRequired("Name")

	return personCmd
}
