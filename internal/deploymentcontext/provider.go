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
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/routing"
)

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

func NewRouteTableProvider(table *routing.RouteTable) RouteProvider {
	return &routeTableRouting{table: *table}
}

var _ RouteProvider = (*routeTableRouting)(nil)

type routeTableRouting struct {
	table routing.RouteTable
}

// Route implements RouteProvider.
func (r *routeTableRouting) Route(module string) string {
	route := r.table.Current().GetForModule(module)
	if r, ok := route.Get(); ok {
		return r.String()
	}
	return ""
}

// Subscribe implements RouteProvider.
func (r *routeTableRouting) Subscribe() chan string {
	return r.table.Subscribe()
}

// Unsubscribe implements RouteProvider.
func (r *routeTableRouting) Unsubscribe(c chan string) {
	r.table.Unsubscribe(c)
}

// NewProvider retrieves config, secrets and DSNs for a module.
func NewProvider(key key.Deployment, routeProvider RouteProvider, moduleSchema *schema.Module, secretsProvider SecretsProvider, configProvider ConfigProvider) DeploymentContextProvider {
	return func(ctx context.Context) <-chan DeploymentContext {

		ret := make(chan DeploymentContext, 16)
		logger := log.FromContext(ctx)
		updates := routeProvider.Subscribe()
		module := moduleSchema.Name

		// Initialize checksum to -1; a zero checksum does occur when the context contains no settings
		lastChecksum := int64(-1)

		callableModules := map[string]bool{}
		egress := map[string]string{}
		for _, decl := range moduleSchema.Decls {
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
			defer routeProvider.Unsubscribe(updates)

			for {
				h := sha.New()

				configs := configProvider(ctx)
				secrets := secretsProvider(ctx)

				routeTable := map[string]string{}
				for _, module := range callableModuleNames {
					if module == moduleSchema.Name {
						continue
					}
					route := routeProvider.Route(module)
					if route == "" {
						continue
					}
					routeTable[module] = route
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
					select {
					case <-ctx.Done():
						return
					case ret <- response:
					}
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
		return ret
	}
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
