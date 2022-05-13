package command

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type CommandContext interface {
	Env() []string
	Dir() string
	Stdin() io.Reader
	Run(cmd *exec.Cmd, path ...string) ([]byte, error)
}

// TODO: Add a default here
type GenericCommandContext struct {
	env   []string
	dir   string
	stdin io.Reader
}

type GenericCommandContextOption func(gcc *GenericCommandContext)

func WithEnv(env ...string) GenericCommandContextOption {
	return func(gcc *GenericCommandContext) {
		gcc.env = make([]string, len(env))
		copy(gcc.env, env)
	}
}

func WithDir(dir string) GenericCommandContextOption {
	return func(gcc *GenericCommandContext) {
		gcc.dir = dir
	}
}

func WithStdin(stdin io.Reader) GenericCommandContextOption {
	return func(gcc *GenericCommandContext) {
		gcc.stdin = stdin
	}
}

func NewGenericCommandContext(opts ...GenericCommandContextOption) *GenericCommandContext {
	gcc := &GenericCommandContext{
		dir:   "",
		stdin: os.Stdin,
	}

	for _, opt := range opts {
		opt(gcc)
	}

	return gcc
}

func (gcc *GenericCommandContext) Run(cmd *exec.Cmd, path ...string) ([]byte, error) {

	dir := strings.Join(append([]string{gcc.dir}, path...), "/")
	// make the directory if it does not already exist
	if dir != "" {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.MkdirAll(dir, 0755)
		}
	}

	cmd.Dir = dir
	cmd.Env = append(os.Environ(), gcc.env...)
	cmd.Stdin = gcc.stdin
	fmt.Println("Running command:", strings.Join(cmd.Args, " "))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, err
	}

	return output, nil
}

func (gcc *GenericCommandContext) Env() []string {
	return gcc.env
}

func (gcc *GenericCommandContext) Dir() string {
	return gcc.dir
}

func (gcc *GenericCommandContext) Stdin() io.Reader {
	return gcc.stdin
}
