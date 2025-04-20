package auth

import (
	"context"
	"net/http"
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
	req = req.WithContext(ctx)
	return http.DefaultClient.Do(req)
}

func (w Wrapper) Host() string {
	return w.host
}
