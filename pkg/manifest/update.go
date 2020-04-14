package manifest

import (
	"path/filepath"

	"github.com/openshift/odo/pkg/manifest/ioutils"
)

type UpdateParameters struct {
	Output string
}

func Update(o *UpdateParameters) error {

	manifestPath := filepath.Join(o.Output, "manifest.yaml")

	// check if the manifest file exists
	exists, err := ioutils.IsExisting(manifestPath)
	if !exists {
		return err
	}

	return nil
}
