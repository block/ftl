package deploymentcontext

import (
	"context"
	sha "crypto/sha256"
	"encoding/binary"
	"hash"
	"sort"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"golang.org/x/exp/maps"

	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/routing"
)

type SecretsProvider func(ctx context.Context) map[string][]byte
type ConfigProvider func(ctx context.Context) map[string][]byte

func NewAdminSecretsProvider(key key.Deployment, adminClient adminpbconnect.AdminServiceClient) SecretsProvider {
	return func(ctx context.Context) map[string][]byte {
		secretsResp, err := adminClient.MapSecretsForModule(ctx, &connect.Request[adminpb.MapSecretsForModuleRequest]{Msg: &adminpb.MapSecretsForModuleRequest{Module: key.Payload.Module}})
		if err != nil {
			log.FromContext(ctx).Errorf(err, "could not get secrets")
			return map[string][]byte{}
		}
		return secretsResp.Msg.Values

	}
}
func NewAdminConfigProvider(key key.Deployment, adminClient adminpbconnect.AdminServiceClient) ConfigProvider {
	return func(ctx context.Context) map[string][]byte {
		configResp, err := adminClient.MapConfigsForModule(ctx, &connect.Request[adminpb.MapConfigsForModuleRequest]{Msg: &adminpb.MapConfigsForModuleRequest{Module: key.Payload.Module}})
		if err != nil {
			log.FromContext(ctx).Errorf(err, "could not get config")
			return map[string][]byte{}
		}
		return configResp.Msg.Values
	}
}

// NewProvider retrieves config, secrets and DSNs for a module.
func NewProvider(ctx context.Context, key key.Deployment, routeTable *routing.RouteTable, deployment *schema.Module, secretsProvider SecretsProvider, configProvider ConfigProvider) (DeploymentContextProvider, error) {
	ret := make(chan DeploymentContext)
	logger := log.FromContext(ctx)
	updates := routeTable.Subscribe()
	module := deployment.Name

	// Initialize checksum to -1; a zero checksum does occur when the context contains no settings
	lastChecksum := int64(-1)

	callableModules := map[string]bool{}
	egress := map[string]string{}
	for _, decl := range deployment.Decls {
		switch entry := decl.(type) {
		case *schema.Verb:
			for _, md := range entry.Metadata {
				if calls, ok := md.(*schema.MetadataCalls); ok {
					for _, call := range calls.Calls {
						callableModules[call.Module] = true
					}
				}
			}
			if entry.Runtime != nil {
				if entry.Runtime.EgressRuntime != nil {
					for _, er := range entry.Runtime.EgressRuntime.Targets {
						egress[er.Expression] = er.Target
					}
				}
			}
		default:

		}
	}
	callableModuleNames := maps.Keys(callableModules)
	callableModuleNames = slices.Sort(callableModuleNames)
	logger.Debugf("Modules %s can call %v", module, callableModuleNames)
	go func() {
		defer routeTable.Unsubscribe(updates)

		for {
			h := sha.New()

			configs := configProvider(ctx)
			secrets := secretsProvider(ctx)
			routeView := routeTable.Current()

			routeTable := map[string]string{}
			for _, module := range callableModuleNames {
				if module == deployment.Name {
					continue
				}
				deployment, ok := routeView.GetDeployment(module).Get()
				if !ok {
					continue
				}
				if route, ok := routeView.Get(deployment).Get(); ok && route.String() != "" {
					routeTable[deployment.String()] = route.String()
				}
			}

			if err := hashConfigurationMap(h, configs); err != nil {
				logger.Errorf(err, "could not detect change on configs")
			}
			if err := hashConfigurationMap(h, secrets); err != nil {
				logger.Errorf(err, "could not detect change on secrets")
			}
			if err := hashRoutesTable(h, routeTable); err != nil {
				logger.Errorf(err, "could not detect change on routes")
			}

			checksum := int64(binary.BigEndian.Uint64((h.Sum(nil))[0:8])) //nolint

			if checksum != lastChecksum {
				logger.Debugf("Sending module context for: %s routes: %v", module, routeTable)
				response := NewBuilder(module).AddConfigs(configs).AddSecrets(secrets).AddEgress(egress).AddRoutes(routeTable).Build()
				ret <- response
				lastChecksum = checksum
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second * 10):
			case <-updates:

			}
		}
	}()
	return ret, nil
}

// hashConfigurationMap computes an order invariant checksum on the configuration
// settings supplied in the map.
func hashConfigurationMap(h hash.Hash, m map[string][]byte) error {
	keys := maps.Keys(m)
	sort.Strings(keys)
	for _, k := range keys {
		_, err := h.Write(append([]byte(k), m[k]...))
		if err != nil {
			return errors.Wrap(err, "error hashing configuration")
		}
	}
	return nil
}

// hashRoutesTable computes an order invariant checksum on the routes
func hashRoutesTable(h hash.Hash, m map[string]string) error {
	keys := maps.Keys(m)
	sort.Strings(keys)
	for _, k := range keys {
		_, err := h.Write(append([]byte(k), m[k]...))
		if err != nil {
			return errors.Wrap(err, "error hashing routes")
		}
	}
	return nil
}

func getDeployment(ctx context.Context, dkey key.Deployment, schemaClient ftlv1connect.SchemaServiceClient) (*schema.Module, error) {
	deployments, err := schemaClient.GetDeployments(ctx, &connect.Request[ftlv1.GetDeploymentsRequest]{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get deployments")
	}
	deploymentMap := map[string]*schema.Module{}
	for _, deployment := range deployments.Msg.Schema {
		module, err := schema.ModuleFromProto(deployment.Schema)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get module from proto")
		}
		deploymentMap[deployment.DeploymentKey] = module
	}

	deployment, ok := deploymentMap[dkey.String()]
	if !ok {
		return nil, errors.WithStack(connect.NewError(connect.CodeInternal, errors.Errorf("could not retrieve deployment: %s", dkey)))
	}
	return deployment, nil
}
