package pipelines

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	"github.com/mitchellh/go-homedir"
	"github.com/openshift/odo/pkg/pipelines/eventlisteners"
	"github.com/openshift/odo/pkg/pipelines/meta"
	"github.com/openshift/odo/pkg/pipelines/roles"
	"github.com/openshift/odo/pkg/pipelines/secrets"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

// InitParameters is a struct that provides flags for initialise command
type AddParameters struct {
	GitopsRepo          string
	GitopsWebhookSecret string
	Output              string
	Prefix              string
	AppGitRepo          string
	AppWebhookSecret    string
	AppImageRepo        string
	EnvName             string
	DockerCfgJson       string
	ServicesGitRepo     string
	SkipChecks          bool
}

const (
	overlays     = "overlays"
	appDir       = "apps"
	configDir    = "config"
	servicesDir  = "services"
	secretName   = "quay"
	configSApath = "base/config/serviceaccount.yaml"
	secretPath   = "base/config/secret.yaml"
	PatchPath    = "overlays/Eventlistener.yaml"
	overlaysDir  = "overlays"
)

// Note: struct fields must be public in order for unmarshal to
// correctly populate the data.

type patchStringValue struct {
	Op    string                          `yaml:"op"`
	Path  string                          `yaml:"path"`
	Value triggersv1.EventListenerTrigger `yaml:"value"`
}

// Init function will initialise the gitops directory
func Add(o *AddParameters) error {

	if !o.SkipChecks {
		installed, err := checkTektonInstall()
		if err != nil {
			return fmt.Errorf("failed to run Tekton Pipelines installation check: %w", err)
		}
		if !installed {
			return errors.New("failed due to Tekton Pipelines or Triggers are not installed")
		}
	}

	gitopsName := getGitopsRepoName(o.GitopsRepo)

	AppRepo := getGitopsRepoName(o.AppGitRepo)

	gitopsPath := filepath.Join(o.Output, gitopsName)

	outputs := map[string]interface{}{}

	exists, _ := isExisting(gitopsPath)

	createKustomizeFiles(outputs)

	if !exists {
		return fmt.Errorf("%s does not exist at %s", gitopsName, gitopsPath)
	}
	configPath := filepath.Join(gitopsPath, servicesDir, o.ServicesGitRepo)

	files := createResourcesConfig(outputs, o.EnvName, o.DockerCfgJson)

	fileNames, err := writeResources(configPath, files)

	fmt.Println("These are the fileNames", fileNames)
	if err != nil {
		return err
	}

	if err := addKustomize("bases", []string{"../base"}, filepath.Join(gitopsPath, appDir, AppRepo, overlays, kustomize)); err != nil {
		return err
	}
	if err := addKustomize("bases", []string{"../../../services/service-1/overlays"}, filepath.Join(gitopsPath, appDir, AppRepo, baseDir, kustomize)); err != nil {
		return err
	}
	if err := addKustomize("bases", []string{"../config"}, filepath.Join(gitopsPath, servicesDir, o.ServicesGitRepo, baseDir, kustomize)); err != nil {
		return err
	}
	if err := addKustomize("bases", []string{"../base"}, filepath.Join(gitopsPath, servicesDir, o.ServicesGitRepo, baseDir, kustomize)); err != nil {
		return err
	}
	if err := addKustomize("bases", []string{"../base"}, filepath.Join(gitopsPath, servicesDir, o.ServicesGitRepo, overlaysDir, kustomize)); err != nil {
		return err
	}
	kustomize1 := map[string][]string{
		"bases":     []string{"../../../../envs/chetan-cicd-environment/overlays/"},
		"resources": []string{"secret.yaml", "serviceaccount.yaml"},
	}

	if err := addModKustomize(kustomize1, filepath.Join(gitopsPath, servicesDir, o.ServicesGitRepo, baseDir, configDir, kustomize)); err != nil {
		return err
	}
	return nil
}
func addModKustomize(values map[string][]string, path string) error {
	content := make([]interface{}, 0)
	for name, items := range values {
		content = append(content, map[string]interface{}{name: items})
	}
	return marshalItemsToFile(path, content)
}

func createResourcesConfig(outputs map[string]interface{}, namespace, dockerFilePath string) map[string]interface{} {
	sa := roles.CreateServiceAccount(meta.NamespacedName(namespace, saName))
	ServiceAcc := roles.AddSecretToSA(sa, secretName)
	outputs[configSApath] = ServiceAcc
	secrets, _ := createSealedDockerSecret(dockerFilePath, namespace)
	outputs[secretPath] = secrets
	return outputs
}

func createSealedDockerSecret(dockerConfigJSONFileName, ns string) (*ssv1alpha1.SealedSecret, error) {
	if dockerConfigJSONFileName == "" {
		return nil, errors.New("failed to generate path to file: --dockerconfigjson flag is not provided")
	}

	authJSONPath, err := homedir.Expand(dockerConfigJSONFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate path to file: %w", err)
	}

	f, err := os.Open(authJSONPath)

	if err != nil {
		return nil, fmt.Errorf("failed to read docker file '%s' : %w", authJSONPath, err)
	}
	defer f.Close()

	dockerSecret, err := secrets.CreateSealedDockerConfigSecret(meta.NamespacedName(ns, dockerSecretName), f)
	if err != nil {
		return nil, err
	}

	return dockerSecret, nil

}

func createKustomizeFiles(outputs map[string]interface{}) {
	// m := make(map[interface{}]interface{})
	t := []patchStringValue{
		patchStringValue{
			Op:    "Name",
			Path:  "SomePath",
			Value: eventlisteners.CreateListenerTrigger("sampleName", "sampleFilter", "sample", "sampleBindingName", "sampleTemplateName"),
		},
		patchStringValue{
			Op:    "Name",
			Path:  "SomePath",
			Value: eventlisteners.CreateListenerTrigger("sampleName", "sampleFilter", "sample", "sampleBindingName", "sampleTemplateName"),
		},
	}
	outputs[PatchPath] = &t

}
