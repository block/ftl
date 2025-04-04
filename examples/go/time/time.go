package time

import (
	"context"
	"time"

	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/go-runtime/ftl"
)

// Simple string configuration
type Greeting = ftl.Config[string]

type TimeRequest struct{}
type TimeResponse struct {
	Time time.Time
}

// Time returns the current time.
//
//ftl:verb export
func Time(ctx context.Context, req TimeRequest, ic InternalClient) (TimeResponse, error) {
	internalTime, err := ic(ctx, req)
	if err != nil {
		return TimeResponse{}, errors.WithStack(err)
	}
	return TimeResponse{Time: internalTime.Time}, nil
}

//ftl:verb export
func Internal(ctx context.Context, req TimeRequest) (TimeResponse, error) {
	return TimeResponse{Time: time.Now()}, nil
}

//ftl:verb export
//ftl:egress url="example.com"
func Hello(ctx context.Context, req string, url ftl.EgressTarget) (string, error) {
	return req + " " + url.GetString(ctx), nil
}
