package auth

import (
	"context"
	stderrors "errors"

	"github.com/crazy-max/diun/v4/internal/secret"
	"github.com/pkg/errors"
)

// StaticOptions configures fixed registry credentials, optionally read from files.
type StaticOptions struct {
	Username     string `yaml:"username,omitempty" json:"username,omitempty" validate:"required_without=UsernameFile"`
	UsernameFile string `yaml:"usernameFile,omitempty" json:"usernameFile,omitempty" validate:"omitempty,file"`
	Password     string `yaml:"password,omitempty" json:"password,omitempty"`
	PasswordFile string `yaml:"passwordFile,omitempty" json:"passwordFile,omitempty" validate:"omitempty,file"`
}

// Static supplies fixed credentials. Its value is comparable so clients using
// identical credentials can share work.
type Static Credentials

// NewStatic reads fixed credentials once. On error it returns any credentials
// successfully resolved.
func NewStatic(opts StaticOptions) (Static, error) {
	username, userErr := secret.GetSecret(opts.Username, opts.UsernameFile)
	password, passwordErr := secret.GetSecret(opts.Password, opts.PasswordFile)
	return Static{
		Username: username,
		Password: password,
	}, stderrors.Join(
		errors.Wrap(userErr, "cannot retrieve username secret"),
		errors.Wrap(passwordErr, "cannot retrieve password secret"),
	)
}

func (s Static) Credentials(context.Context, string) (Credentials, error) {
	return Credentials(s), nil
}
