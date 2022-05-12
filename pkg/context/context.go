package context

import (
	"github.com/everettraven/plugin-testing-poc/pkg/command"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Context interface {
	// TODO: Narrow this scope down when
	// known exactly what info is needed from the test context
	// CommandContext returns the CommandContext that the Context uses for execution
	CommandContext() command.CommandContext
	// Domain returns the domain that is being used during subcommand execution
	Domain() string
	// GVK returns the GroupVersionKind that is being used during execution
	GVK() schema.GroupVersionKind
	// ImageName returns the name of the image that should be used when testing on cluster
	ImageName() string

	// Actual functions
	// Prepare will prepare the test environment
	Prepare() error
	// Init runs an `init` subcommand
	Init(initOptions ...string) error
	// CreateApi runs a `create api` subcommand
	CreateApi(apiOptions ...string) error
	// CreateWebhook runs a `create webhook` subcommand
	CreateWebhook(webhookOptions ...string) error
	// Make runs a `make` command
	Make(makeOptions ...string) error
	// Command runs the specified command
	Command(command ...string) error
}

// TestContext implements Context and can be used for a simple testing context
type TestContext struct {
	testSuffix     string
	domain         string
	gvk            schema.GroupVersionKind
	imageName      string
	binary         string
	commandContext command.CommandContext
}

type TestContextOption func(t *TestContext)

// TODO: Add Option functions

// WithTestSuffix sets the suffix to be used when creating a temporary testing directory
func WithTestSuffix(suffix string) TestContextOption {
	return func(t *TestContext) {
		t.testSuffix = suffix
	}
}

// WithDomain sets the domain to be used when executing subcommands that require a domain
func WithDomain(domain string) TestContextOption {
	return func(t *TestContext) {
		t.domain = domain
	}
}

// WithGvk sets the GroupVersionKind to be used when executing subcommands that require a Group, Version, and Kind
func WithGvk(gvk schema.GroupVersionKind) TestContextOption {
	return func(t *TestContext) {
		t.gvk = gvk
	}
}

// WithImageName sets the image name to be used
func WithImageName(imageName string) TestContextOption {
	return func(t *TestContext) {
		t.imageName = imageName
	}
}

// WithBinary sets the binary that is used to execute scaffold subcommands
func WithBinary(binary string) TestContextOption {
	return func(t *TestContext) {
		t.binary = binary
	}
}

// WithCommandContext sets the command context that is used to execute commands

// func NewTestContext(opts ...TestContextOption) {
// 	// defaults
// 	tc := &TestContext{
// 		// TODO: consider making this a random string
// 		testSuffix: "temp",
// 	}
// }

// // TODO: Implement the interface
// func (tc *TestContext) Prepare() error {
// 	// Clean up tools so the correct version is installed for each test
// 	gingko.GinkgoWriter.Println("cleaning up tools")
// 	tools := []string{"controller-gen", "kustomize"}
// 	for _, tool := range tools {

// 	}

// }
