package scaling

import (
	"context"
	"net/url"

	"github.com/block/ftl/common/schema"
)

type RunnerScaling interface {
	Start(ctx context.Context) error

	StartDeployment(ctx context.Context, deployment string, sch *schema.Module, hasCron bool, hasIngress bool, subscriptionProcessor bool) (url.URL, error)

	UpdateDeployment(ctx context.Context, deployment string, sch *schema.Module, subscriptionProcessor bool) error

	TerminateDeployment(ctx context.Context, deployment string, subscriptionProcessor bool) error
}
