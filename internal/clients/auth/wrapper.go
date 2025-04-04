package auth

import (
	"context"
	"net/http"
)

type Wrapper struct {
	url string
}

func New(url string) Wrapper {
	return Wrapper{
		url: url,
	}
}

func (w Wrapper) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	return http.DefaultClient.Do(req)
}
