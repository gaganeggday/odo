package config

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/pkg/pipelines/ioutils"
)

func TestParse(t *testing.T) {
	parseTests := []struct {
		filename string
		want     *Manifest
	}{
		{"testdata/example1.yaml", &Manifest{
			GitOpsURL: "https://github.com/...",
			Config: &Config{
				Pipelines: &PipelinesConfig{
					Name: "test-pipelines",
				},
				ArgoCD: &ArgoCDConfig{
					Namespace: "test-argocd",
				},
			},
			Environments: []*Environment{
				{
					Name: "development",
					Pipelines: &Pipelines{
						Integration: &TemplateBinding{
							Template: "dev-ci-template",
							Bindings: []string{"dev-ci-binding"},
						},
					},
					Services: []*Service{
						{
							Name:      "service-http",
							SourceURL: "https://github.com/myproject/myservice.git",
						},
						{Name: "service-redis"},
					},
				},
				{
					Name: "staging",
				},
				{
					Name: "production",
					Services: []*Service{
						{Name: "service-http"},
						{Name: "service-metrics"},
					},
				},
			},
			Apps: []*Application{
				{
					Name: "my-app-1",
					Environments: []*EnvironmentRefs{
						{
							Refs: "development",
							ServiceRefs: []string{
								"service-http",
								"service-redis",
							},
						},
						{
							Refs: "production",
							ServiceRefs: []string{
								"service-http",
								"service-metrics",
							},
						},
					},
				},
			},
		},
		},

		{"testdata/example2.yaml", &Manifest{
			Environments: []*Environment{
				{
					Name: "development",
					Services: []*Service{
						{
							Name:      "app-1-service-http",
							SourceURL: "https://github.com/myproject/myservice.git",
						},
						{Name: "app-1-service-metrics"},
					},
				},
				{
					Name: "tst-cicd",
				},
			},
			Apps: []*Application{
				{
					Name: "my-app-1",
					Environments: []*EnvironmentRefs{
						{
							Refs: "development",
							ServiceRefs: []string{
								"app-1-service-http",
								"app-1-service-metrics",
							},
						},
					},
				},
			},
		},
		},
	}

	for _, tt := range parseTests {
		t.Run(fmt.Sprintf("parsing %s", tt.filename), func(rt *testing.T) {
			fs := ioutils.NewFilesystem()
			f, err := fs.Open(tt.filename)
			if err != nil {
				rt.Fatalf("failed to open %v: %s", tt.filename, err)
			}
			defer f.Close()

			got, err := Parse(f)
			if err != nil {
				rt.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				rt.Errorf("Parse(%s) failed diff\n%s", tt.filename, diff)
			}
		})
	}
}
