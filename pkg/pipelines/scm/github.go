package scm

import (
	"net/url"

	"github.com/openshift/odo/pkg/pipelines/meta"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
)

var (
	triggerBindingTypeMeta = meta.TypeMeta("TriggerBinding", "tekton.dev/v1alpha1")
)

// Filters for interceptors
const (
	githubCIDryRunFilters = "(header.match('X-GitHub-Event', 'pull_request') && body.action == 'opened' || body.action == 'synchronize') && body.pull_request.head.repo.full_name == '%s'"

	githubCDDeployFilters = "(header.match('X-GitHub-Event', 'push') && body.repository.full_name == '%s') && body.ref.startsWith('refs/heads/master')"

	webhookSecretKey = "webhook-secret-key"

	githubPRBindingName   = "github-pr-binding"
	githubPushBindingName = "github-push-binding"
)

// GithubRepository represents a service on a Github repo
type GithubRepository struct {
	url *url.URL
}

// NewGithubRepository returns an instance of GithubRepository
func NewGithubRepository(rawURL string) (*GithubRepository, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &GithubRepository{url: parsedURL}, nil
}

// CreatePRBinding returns a TriggerBinding for Github PullRequest hooks.
func (repo *GithubRepository) CreatePRBinding(ns string) (triggersv1.TriggerBinding, string) {
	return triggersv1.TriggerBinding{
		TypeMeta:   triggerBindingTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, githubPRBindingName)),
		Spec: triggersv1.TriggerBindingSpec{
			Params: []pipelinev1.Param{
				createBindingParam("gitref", "$(body.pull_request.head.ref)"),
				createBindingParam("gitsha", "$(body.pull_request.head.sha)"),
				createBindingParam("gitrepositoryurl", "$(body.repository.clone_url)"),
				createBindingParam("fullname", "$(body.repository.full_name)"),
			},
		},
	}, githubPRBindingName
}

// CreatePushBinding returns a TriggerBinding for Github Push hooks.
func (repo *GithubRepository) CreatePushBinding(ns string) (triggersv1.TriggerBinding, string) {
	return triggersv1.TriggerBinding{
		TypeMeta:   triggerBindingTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, githubPushBindingName)),
		Spec: triggersv1.TriggerBindingSpec{
			Params: []pipelinev1.Param{
				createBindingParam("gitref", "$(body.ref)"),
				createBindingParam("gitsha", "$(body.head_commit.id)"),
				createBindingParam("gitrepositoryurl", "$(body.repository.clone_url)"),
			},
		},
	}, githubPushBindingName
}

// URL returns the URL of the Github repository
func (repo *GithubRepository) URL() string {
	return repo.url.String()
}

// CreateCITrigger creates a CI eventlistener trigger for Github
func (repo *GithubRepository) CreateCITrigger(name, secretName, secretNs, template string, bindings []string) (v1alpha1.EventListenerTrigger, error) {
	repoName, err := getRepoName(repo.URL())
	if err != nil {
		return v1alpha1.EventListenerTrigger{}, err
	}
	return createListenerTrigger(repo, name, githubCIDryRunFilters, repoName, secretName, secretNs, template, bindings), nil
}

// CreateCDTrigger creates a CD eventlistener trigger for Github
func (repo *GithubRepository) CreateCDTrigger(name, secretName, secretNs, template string, bindings []string) (v1alpha1.EventListenerTrigger, error) {
	repoName, err := getRepoName(repo.URL())
	if err != nil {
		return v1alpha1.EventListenerTrigger{}, err
	}
	return createListenerTrigger(repo, name, githubCDDeployFilters, repoName, secretName, secretNs, template, bindings), nil
}

// CreateInterceptor returns a Github event interceptor
func (repo *GithubRepository) CreateInterceptor(secretName, secretNs string) *triggersv1.EventInterceptor {
	return &triggersv1.EventInterceptor{
		GitHub: &triggersv1.GitHubInterceptor{
			SecretRef: &triggersv1.SecretRef{
				SecretName: secretName,
				SecretKey:  webhookSecretKey,
				Namespace:  secretNs,
			},
		},
	}
}
