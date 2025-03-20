package provisioner

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/mysql"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	_ "github.com/go-sql-driver/mysql" // SQL driver
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
)

// NewFixtureProvisioner creates a new provisioner that provisions fixtures in dev mode
func NewFixtureProvisioner() *InMemProvisioner {
	return NewEmbeddedProvisioner(map[schema.ResourceType]InMemResourceProvisionerFn{
		schema.ResourceTypeFixture: provisionFixture(),
	}, make(map[schema.ResourceType]InMemResourceProvisionerFn))
}

func provisionFixture() InMemResourceProvisionerFn {
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, resource schema.Provisioned, module *schema.Module) (*schema.RuntimeElement, error) {
		logger := log.FromContext(ctx)

		verb, ok := resource.(*schema.Verb)
		if !ok {
			return nil, fmt.Errorf("expected verb, got %T", resource)
		}
		fixtures := []*schema.MetadataFixture{}
		for fixture := range slices.FilterVariants[*schema.MetadataFixture](verb.Metadata) {
			if !fixture.Manual {
				fixtures = append(fixtures, fixture)
			}
		}
		if len(fixtures) == 0 {
			return nil, nil
		}
		endpoint := module.GetRuntime().GetRunner().GetEndpoint()
		if endpoint == "" {
			return nil, errors.New("runner endpoint is required")
		}
		client := rpc.Dial(ftlv1connect.NewVerbServiceClient, endpoint, log.Debug)
		// We should not need to wait here, the runner should already be up
		logger.Debugf("Calling fixture verb %s", verb.Name)
		_, err := client.Call(ctx, connect.NewRequest(&ftlv1.CallRequest{Verb: &schemapb.Ref{Module: module.Name, Name: verb.Name}, Body: []byte("{}")}))
		if err != nil {
			return nil, fmt.Errorf("failed to call verb: %w", err)
		}
		return nil, nil
	}
}
