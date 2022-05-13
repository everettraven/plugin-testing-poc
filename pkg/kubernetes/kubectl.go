package kubernetes

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/everettraven/plugin-testing-poc/pkg/command"
)

type Kubectl interface {
	CommandContext() command.CommandContext
	Namespace() string
	ServiceAccount() string

	// Actual functions
	Command(options ...string) (string, error)
	CommandInNamespace(options ...string) (string, error)
	Apply(inNamespace bool, options ...string) (string, error)
	Get(inNamespace bool, options ...string) (string, error)
	Delete(inNamespace bool, options ...string) (string, error)
	Logs(inNamespace bool, options ...string) (string, error)
	Wait(inNamespace bool, options ...string) (string, error)
	Version() (KubernetesVersion, error)
}

// TODO: Add a default here
type KubectlUtil struct {
	commandContext command.CommandContext
	namespace      string
	serviceAccount string
}

type KubectlUtilOptions func(ku *KubectlUtil)

func WithCommandContext(cc command.CommandContext) KubectlUtilOptions {
	return func(ku *KubectlUtil) {
		ku.commandContext = cc
	}
}

func WithNamespace(ns string) KubectlUtilOptions {
	return func(ku *KubectlUtil) {
		ku.namespace = ns
	}
}

func WithServiceAccount(sa string) KubectlUtilOptions {
	return func(ku *KubectlUtil) {
		ku.serviceAccount = sa
	}
}

// TODO: Implement interface

func NewKubectlUtil(opts ...KubectlUtilOptions) *KubectlUtil {
	ku := &KubectlUtil{
		commandContext: command.NewGenericCommandContext(),
		namespace:      "test-ns",
		serviceAccount: "test-sa",
	}

	for _, opt := range opts {
		opt(ku)
	}

	return ku
}

func (ku *KubectlUtil) CommandContext() command.CommandContext {
	return ku.commandContext
}

func (ku *KubectlUtil) Namespace() string {
	return ku.namespace
}

func (ku *KubectlUtil) ServiceAccount() string {
	return ku.serviceAccount
}

func (ku *KubectlUtil) Command(options ...string) (string, error) {
	cmd := exec.Command("kubectl", options...)
	output, err := ku.commandContext.Run(cmd)
	return string(output), err
}

func (ku *KubectlUtil) CommandInNamespace(options ...string) (string, error) {
	opts := append([]string{"-n", ku.namespace}, options...)
	return ku.Command(opts...)
}

func (ku *KubectlUtil) Apply(inNamespace bool, options ...string) (string, error) {
	return ku.prefixCommand("apply", inNamespace, options...)
}

func (ku *KubectlUtil) Get(inNamespace bool, options ...string) (string, error) {
	return ku.prefixCommand("get", inNamespace, options...)
}

func (ku *KubectlUtil) Delete(inNamespace bool, options ...string) (string, error) {
	return ku.prefixCommand("delete", inNamespace, options...)
}

func (ku *KubectlUtil) Logs(inNamespace bool, options ...string) (string, error) {
	return ku.prefixCommand("logs", inNamespace, options...)
}

func (ku *KubectlUtil) Wait(inNamespace bool, options ...string) (string, error) {
	return ku.prefixCommand("wait", inNamespace, options...)
}

func (ku *KubectlUtil) Version() (KubernetesVersion, error) {
	out, err := ku.Command("version", "-o", "json")
	if err != nil {
		return nil, err
	}

	var versions map[string]json.RawMessage

	err = json.Unmarshal([]byte(out), &versions)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling json: %w", err)
	}

	clientVersion, err := NewKubeVersionInfo(string(versions["clientVersion"]))
	if err != nil {
		return nil, fmt.Errorf("error getting client version: %w", err)
	}

	serverVersion, err := NewKubeVersionInfo(string(versions["serverVersion"]))
	if err != nil {
		return nil, fmt.Errorf("error getting server version: %w", err)
	}

	return NewKubeVersion(WithClientVersion(clientVersion), WithServerVersion(serverVersion)), nil
}

func (ku *KubectlUtil) prefixCommand(prefix string, inNamespace bool, options ...string) (string, error) {
	opts := append([]string{prefix}, options...)

	if inNamespace {
		return ku.CommandInNamespace(opts...)
	}

	return ku.Command(opts...)
}
