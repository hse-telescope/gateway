package auth

import (
	"context"
	"net/http"

	"github.com/hse-telescope/tracer"
	"github.com/hse-telescope/utils/requests"
)

type Wrapper struct {
	host string
}

func New(host string) Wrapper {
	return Wrapper{
		host: host,
	}
}

func (w Wrapper) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	ctx, span := tracer.Start(ctx, "sending auth request")
	defer span.End()
	return requests.Do(ctx, req)
}

func (w Wrapper) Host() string {
	return w.host
}
