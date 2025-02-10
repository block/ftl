package scaling

import (
	"context"
	"net/url"

	"github.com/block/ftl/common/schema"
)

type RunnerScaling interface {
	Start(ctx context.Context) error

	StartDeployment(ctx context.Context, deployment string, sch *schema.Module, hasCron bool, hasIngress bool) (url.URL, error)

	UpdateDeployment(ctx context.Context, deployment string, sch *schema.Module) error

	TerminateDeployment(ctx context.Context, deployment string) error
}
