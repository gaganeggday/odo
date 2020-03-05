package pipelines

import (
	"bytes"
	"fmt"
	"io"

	"sigs.k8s.io/yaml"
)

// marshalOutputs marshal outputs to given writer
func marshalOutputs(out io.Writer, values []interface{}) error {
	outputs := make([][]byte, len(values))
	for i, r := range values {
		data, err := yaml.Marshal(r)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}
		outputs[i] = data
	}

	marshaled := bytes.Join(outputs, []byte("---\n"))

	_, err := out.Write(marshaled)
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}
	return nil
}
