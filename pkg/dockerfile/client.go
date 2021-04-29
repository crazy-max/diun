package dockerfile

import (
	"bytes"
	"io/ioutil"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/shell"
	"github.com/pkg/errors"
)

// Client represents an active dockerfile object
type Client struct {
	ast      *parser.Node
	stages   []instructions.Stage
	metaArgs []instructions.KeyValuePairOptional
	shlex    *shell.Lex
}

// Options holds dockerfile client object options
type Options struct {
	Filename string
}

// New initializes a new dockerfile client
func New(opts Options) (*Client, error) {
	b, err := ioutil.ReadFile(opts.Filename)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot read Dockerfile %s", opts.Filename)
	}

	parsed, err := parser.Parse(bytes.NewReader(b))
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot parse Dockerfile %s", opts.Filename)
	}

	stages, metaArgs, err := instructions.Parse(parsed.AST)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot parse stages for Dockerfile %s", opts.Filename)
	}

	var kvpoArgs []instructions.KeyValuePairOptional
	shlex := shell.NewLex(parsed.EscapeToken)
	for _, cmd := range metaArgs {
		for _, metaArg := range cmd.Args {
			if metaArg.Value != nil {
				*metaArg.Value, _ = shlex.ProcessWordWithMap(*metaArg.Value, metaArgsToMap(kvpoArgs))
			}
			kvpoArgs = append(kvpoArgs, metaArg)
		}
	}

	return &Client{
		ast:      parsed.AST,
		stages:   stages,
		metaArgs: kvpoArgs,
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

func metaArgsToMap(metaArgs []instructions.KeyValuePairOptional) map[string]string {
	m := map[string]string{}
	for _, arg := range metaArgs {
		m[arg.Key] = arg.ValueString()
	}
	return m
}
