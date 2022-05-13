package samples

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/everettraven/plugin-testing-poc/pkg/command"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Sample interface {
	CommandContext() command.CommandContext
	Name() string
	GVK() schema.GroupVersionKind
	GenerateInit() error
	GenerateApi() error
	GenerateWebhook() error
}

// TODO: Add a default here
type GenericSample struct {
	domain         string
	repo           string
	gvk            schema.GroupVersionKind
	commandContext command.CommandContext
	name           string
	binary         string
	plugins        []string

	initOptions    []string
	apiOptions     []string
	webhookOptions []string

	// TODO: Add values that allow for implementing hooks in the logic
}

type GenericSampleOption func(gs *GenericSample)

// WithDomain sets the domain to be used during scaffold execution
func WithDomain(domain string) GenericSampleOption {
	return func(gs *GenericSample) {
		gs.domain = domain
	}
}

// WithRepository sets the repository to be used during scaffold execution
func WithRepository(repo string) GenericSampleOption {
	return func(gs *GenericSample) {
		gs.repo = repo
	}
}

// WithGvk sets the GroupVersionKind to be used during scaffold execution
func WithGvk(gvk schema.GroupVersionKind) GenericSampleOption {
	return func(gs *GenericSample) {
		gs.gvk = gvk
	}
}

// WithName sets the name of the sample that is scaffolded
func WithName(name string) GenericSampleOption {
	return func(gs *GenericSample) {
		gs.name = name
	}
}

// WithCommandContext sets the CommandContext that is used to execute scaffold commands
func WithCommandContext(commandContext command.CommandContext) GenericSampleOption {
	return func(gs *GenericSample) {
		gs.commandContext = commandContext
	}
}

// WithBinary sets the binary that should be used to run scaffold commands
func WithBinary(binary string) GenericSampleOption {
	return func(gs *GenericSample) {
		gs.binary = binary
	}
}

// WithPlugins sets the plugins that should be used during scaffolding
func WithPlugins(plugins ...string) GenericSampleOption {
	return func(gs *GenericSample) {
		gs.plugins = make([]string, len(plugins))
		copy(gs.plugins, plugins)
	}
}

func WithExtraInitOptions(options ...string) GenericSampleOption {
	return func(gs *GenericSample) {
		gs.initOptions = make([]string, len(options))
		copy(gs.initOptions, options)
	}
}

func WithExtraApiOptions(options ...string) GenericSampleOption {
	return func(gs *GenericSample) {
		gs.apiOptions = make([]string, len(options))
		copy(gs.apiOptions, options)
	}
}

func WithExtraWebhookOptions(options ...string) GenericSampleOption {
	return func(gs *GenericSample) {
		gs.webhookOptions = make([]string, len(options))
		copy(gs.webhookOptions, options)
	}
}

func NewGenericSample(opts ...GenericSampleOption) *GenericSample {
	gs := &GenericSample{
		domain: "example.com",
		name:   "generic-sample",
		gvk: schema.GroupVersionKind{
			Group:   "sample",
			Version: "v1",
			Kind:    "Generic",
		},
		// by default use kubebuilder unless otherwise specified
		binary:         "kubebuilder",
		repo:           "",
		commandContext: command.NewGenericCommandContext(),
		plugins:        []string{"go/v3"},
	}

	for _, opt := range opts {
		opt(gs)
	}

	return gs
}

func (gs *GenericSample) CommandContext() command.CommandContext {
	return gs.commandContext
}

func (gs *GenericSample) Name() string {
	return gs.name
}

func (gs *GenericSample) GVK() schema.GroupVersionKind {
	return gs.gvk
}

func (gs *GenericSample) GenerateInit() error {
	options := []string{
		"init",
		"--plugins",
		strings.TrimRight(strings.Join(gs.plugins, ","), ","),
		"--domain",
		gs.domain,
	}

	if gs.repo != "" {
		options = append(options, "--repo", gs.repo)
	}

	options = append(options, gs.initOptions...)

	ex := exec.Command(gs.binary, options...)

	output, err := gs.commandContext.Run(ex, gs.name)
	if err != nil {
		return fmt.Errorf("error running command: ", err, "output:", string(output))
	}

	return nil
}

func (gs *GenericSample) GenerateApi() error {
	options := []string{
		"create",
		"api",
		"--plugins",
		strings.TrimRight(strings.Join(gs.plugins, ","), ","),
		"--group",
		gs.gvk.Group,
		"--version",
		gs.gvk.Version,
		"--kind",
		gs.gvk.Kind,
	}

	options = append(options, gs.apiOptions...)

	ex := exec.Command(gs.binary, options...)

	output, err := gs.commandContext.Run(ex, gs.name)
	if err != nil {
		return fmt.Errorf("error running command: ", err, "output:", string(output))
	}

	return nil
}

func (gs *GenericSample) GenerateWebhook() error {
	options := []string{
		"create",
		"webhook",
		"--plugins",
		strings.TrimRight(strings.Join(gs.plugins, ","), ","),
		"--group",
		gs.gvk.Group,
		"--version",
		gs.gvk.Version,
		"--kind",
		gs.gvk.Kind,
	}

	options = append(options, gs.webhookOptions...)

	ex := exec.Command(gs.binary, options...)

	output, err := gs.commandContext.Run(ex, gs.name)
	if err != nil {
		return fmt.Errorf("error running command: ", err, "output:", string(output))
	}

	return nil
}
