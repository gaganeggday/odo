package pipelines

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/tests/helper"
)

func TestMarshalOutputs(t *testing.T) {
	marshalTests := []struct {
		name   string
		values []interface{}
		want   []byte
	}{
		{
			name:   "single document",
			values: []interface{}{map[string]string{"testing": "value"}},
			want:   []byte("testing: value\n"),
		},
		{
			name:   "multiple documents",
			values: []interface{}{yamlDoc(1), yamlDoc(2)},
			want:   []byte("item1: value1\n---\nitem2: value2\n"),
		},
	}

	for _, tt := range marshalTests {
		var out bytes.Buffer
		err := marshalOutputs(&out, tt.values)
		if err != nil {
			t.Errorf("marshalOutputs() %s got error: %s", tt.name, err)
			continue
		}
		if diff := cmp.Diff(tt.want, out.Bytes()); diff != "" {
			t.Errorf("marshalOutputs() %s failed: %s", tt.name, diff)
		}
	}
}

func TestMarshalOutputsWithError(t *testing.T) {
	marshalTests := []struct {
		name    string
		values  []interface{}
		wantErr string
	}{
		{
			name:    "marshal something unmarshalable",
			values:  []interface{}{func() int { return 2 }},
			wantErr: "failed to marshal data.*unsupported type",
		},
	}

	for _, tt := range marshalTests {
		var out bytes.Buffer
		err := marshalOutputs(&out, tt.values)
		if !helper.MatchErrorString(t, tt.wantErr, err) {
			t.Errorf("marshalOutputs() %s got %s, want %s", tt.name, err, tt.wantErr)
		}
	}
}

func TestMarshalFailingToWrite(t *testing.T) {
	values := []interface{}{yamlDoc(1)}
	out := diskFullWriter{}

	err := marshalOutputs(&out, values)
	wantErr := "failed to write data: disk full"
	if !helper.MatchErrorString(t, wantErr, err) {
		t.Errorf("marshalOutputs() %s, want %s", err, wantErr)
	}
}

type diskFullWriter struct {
}

func (bw *diskFullWriter) Write(p []byte) (int, error) {
	return 0, errors.New("disk full")
}

func yamlDoc(n int) map[string]string {
	return map[string]string{fmt.Sprintf("item%d", n): fmt.Sprintf("value%d", n)}
}
