package auth

import (
	"context"

	"github.com/hse-telescope/gateway/internal/clients/auth"
)

type Provider struct {
	auth auth.Wrapper
}

func New(auth auth.Wrapper) Provider {
	return Provider{
		auth: auth,
	}
}

func (p Provider) CheckToken(ctx context.Context, token string) bool {
	return false
}
