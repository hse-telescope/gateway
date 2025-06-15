package auth

import (
	"context"
	"net/http"

	"github.com/hse-telescope/tracer"
	"go.opentelemetry.io/otel/propagation"
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
	propagator := propagation.TraceContext{}
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))
	req = req.WithContext(ctx)
	return http.DefaultClient.Do(req)
}

func (w Wrapper) Host() string {
	return w.host
}
