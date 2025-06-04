package provisioner

import (
	"context"
	"encoding/json"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	egress2 "github.com/block/ftl/internal/egress"
)

// NewEgressProvisioner creates a new provisioner that provisions egress, it simply interpolates the egress values from config
func NewEgressProvisioner(adminClient adminpbconnect.AdminServiceClient) *InMemProvisioner {
	return NewEmbeddedProvisioner(map[schema.ResourceType]InMemResourceProvisionerFn{
		schema.ResourceTypeEgress: provisionEgress(adminClient),
	}, map[schema.ResourceType]InMemResourceProvisionerFn{})
}

func provisionEgress(adminClient adminpbconnect.AdminServiceClient) InMemResourceProvisionerFn {
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, res schema.Provisioned, module *schema.Module) (*schema.RuntimeElement, error) {
		logger := log.FromContext(ctx)
		verb, ok := res.(*schema.Verb)
		if !ok {
			panic(errors.Errorf("unexpected resource type: %T", res))
		}
		for egress := range slices.FilterVariants[*schema.MetadataEgress](verb.Metadata) {
			elements := []schema.EgressTarget{}
			logger.Debugf("Provisioning egress for verb: %s", verb.Name)
			for _, e := range egress.Targets {
				interpolated, err := egress2.Interpolate(
					e, func(s string) (string, error) {
						res, err := adminClient.ConfigGet(ctx, connect.NewRequest(&adminpb.ConfigGetRequest{Ref: &adminpb.ConfigRef{Name: s, Module: &module.Name}}))
						if err != nil {
							return "", errors.Wrapf(err, "failed to get config %q", s)
						}
						val := ""
						err = json.Unmarshal(res.Msg.Value, &val)
						if err != nil {
							return "", errors.Wrapf(err, "failed to unmarshal config %q", s)
						}
						return val, nil
					})
				if err != nil {
					return nil, errors.Wrapf(err, "failed to interpolate egress target %q", e)
				}
				elements = append(elements, schema.EgressTarget{Expression: e, Target: interpolated})
			}
			return &schema.RuntimeElement{
				Name:       optional.Some(res.ResourceID()),
				Deployment: deployment,
				Element: &schema.EgressRuntime{
					Targets: elements,
				},
			}, nil
		}
		return nil, nil
	}
}
