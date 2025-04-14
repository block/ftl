package ftl

import (
	"context"
	"net/url"
	"strings"

	errors "github.com/alecthomas/errors"
)

type spiffeIDKey struct{}

func New(spiffeID Option[url.URL]) *SpiffeIdentity {
	return &SpiffeIdentity{url: spiffeID}
}

func WorkloadIdentity(ctx context.Context) SpiffeIdentity {
	value := ctx.Value(spiffeIDKey{})
	if value == nil {
		return SpiffeIdentity{}
	}
	if ptr, ok := value.(*url.URL); ok {
		return SpiffeIdentity{url: Ptr(ptr)}
	}
	return SpiffeIdentity{}
}

func ContextWithSpiffeIdentity(ctx context.Context, id *url.URL) context.Context {
	return context.WithValue(ctx, spiffeIDKey{}, id)
}

type SpiffeIdentity struct {
	url Option[url.URL]
}

// SpiffeID returns the SPIFFE ID of the workload in URI format.
func (r *SpiffeIdentity) SpiffeID() (url.URL, error) {
	if id, ok := r.url.Get(); ok {
		return id, nil
	}
	return url.URL{}, errors.Errorf("no workload identity found")
}

// AppID returns the application ID of the workload, which is inferred from the namespace of the spiffe ID.
func (r *SpiffeIdentity) AppID() (string, error) {
	if id, ok := r.url.Get(); ok {
		path := id.Path
		elements := strings.Split(path, "/")
		elementCount := len(elements)
		if elementCount < 4 {
			return "", errors.Errorf("no app ID found, expected istio /ns/.../sa/... path")
		}
		if elements[elementCount-2] != "sa" && elements[elementCount-4] != "ns" {
			return "", errors.Errorf("no app ID found, expected istio /ns/.../sa/... path")
		}
		return elements[elementCount-3], nil
	}
	return "", errors.Errorf("no workload identity found")
}
