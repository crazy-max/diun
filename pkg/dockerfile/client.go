package dockerfile

import (
	"bytes"
	"os"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/shell"
	"github.com/pkg/errors"
)

// Client represents an active dockerfile object
type Client struct {
	ast      *parser.Node
	stages   []instructions.Stage
	metaArgs shell.EnvGetter
	shlex    *shell.Lex
}

// Options holds dockerfile client object options
type Options struct {
	Filename string
}

// New initializes a new dockerfile client
func New(opts Options) (*Client, error) {
	b, err := os.ReadFile(opts.Filename)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read Dockerfile %s", opts.Filename)
	}

	parsed, err := parser.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse Dockerfile %s", opts.Filename)
	}

	stages, metaArgs, err := instructions.Parse(parsed.AST, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot parse stages for Dockerfile %s", opts.Filename)
	}

	var kvpoArgs []string
	shlex := shell.NewLex(parsed.EscapeToken)
	for _, cmd := range metaArgs {
		for _, metaArg := range cmd.Args {
			if metaArg.Value != nil {
				if name, _, err := shlex.ProcessWord(*metaArg.Value, shell.EnvsFromSlice(kvpoArgs)); err == nil {
					metaArg.Value = &name
				}
			}
			kvpoArgs = append(kvpoArgs, metaArg.String())
		}
	}

	return &Client{
		ast:      parsed.AST,
		stages:   stages,
		metaArgs: shell.EnvsFromSlice(kvpoArgs),
		shlex:    shlex,
	}, nil
}

func (c *Client) isStageName(name string) bool {
	for _, stage := range c.stages {
		if stage.Name == name {
			return true
		}
	}
	return false
}
