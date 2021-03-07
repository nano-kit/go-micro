package auth

import (
	"fmt"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/micro/go-micro/v2/auth/provider/basic"
)

var (
	DefaultAuth = NewAuth()
)

func NewAuth(opts ...Option) Auth {
	options := Options{
		Provider: basic.NewProvider(),
	}

	for _, o := range opts {
		o(&options)
	}

	return &noop{
		opts: options,
	}
}

type noop struct {
	opts Options
}

// String returns the name of the implementation
func (n *noop) String() string {
	return "noop"
}

// Init the auth
func (n *noop) Init(opts ...Option) {
	for _, o := range opts {
		o(&n.opts)
	}
}

// Options set for auth
func (n *noop) Options() Options {
	return n.opts
}

// Generate a new account
func (n *noop) Generate(id string, opts ...GenerateOption) (*Account, error) {
	options := NewGenerateOptions(opts...)

	return &Account{
		ID:       id,
		Secret:   options.Secret,
		Metadata: options.Metadata,
		Scopes:   options.Scopes,
		Issuer:   n.Options().Namespace,
	}, nil
}

// Grant access to a resource
func (n *noop) Grant(rule *Rule) error {
	return nil
}

// Revoke access to a resource
func (n *noop) Revoke(rule *Rule) error {
	return nil
}

// Rules used to verify requests
func (n *noop) Rules(opts ...RulesOption) ([]*Rule, error) {
	return []*Rule{}, nil
}

// Verify an account has access to a resource
func (n *noop) Verify(acc *Account, res *Resource, opts ...VerifyOption) error {
	return nil
}

// Inspect a token
func (n *noop) Inspect(token string) (*Account, error) {
	// Because this token has already been inspected at API gateway
	// try to decode JWT locally and do not verify signature
	if len(strings.Split(token, ".")) == 3 {
		// authClaims to be encoded in the JWT
		type authClaims struct {
			Type     string            `json:"type"`
			Scopes   []string          `json:"scopes"`
			Metadata map[string]string `json:"metadata"`

			jwt.StandardClaims
		}
		res, _, err := new(jwt.Parser).ParseUnverified(token, &authClaims{})
		if err != nil {
			return nil, fmt.Errorf("jwt parse: %v", err)
		}
		claims, ok := res.Claims.(*authClaims)
		if !ok {
			return nil, fmt.Errorf("jwt claims type is incorrect")
		}
		// return the token
		return &Account{
			ID:       claims.Subject,
			Issuer:   claims.Issuer,
			Type:     claims.Type,
			Scopes:   claims.Scopes,
			Metadata: claims.Metadata,
		}, nil
	}

	return &Account{ID: uuid.New().String(), Issuer: n.Options().Namespace}, nil
}

// Token generation using an account id and secret
func (n *noop) Token(opts ...TokenOption) (*Token, error) {
	return &Token{}, nil
}
