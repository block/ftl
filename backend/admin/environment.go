package admin

import (
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	. "github.com/alecthomas/types/optional"

	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/encoding"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	configuration "github.com/block/ftl/internal/config"
	"github.com/block/ftl/internal/maps"
)

type SchemaClient interface {
	GetSchema(ctx context.Context) (*schema.Schema, error)
}

// EnvironmentManager is a client that reads and writes secrets and config entries
type EnvironmentManager struct {
	schr SchemaClient
	cm   configuration.Provider[configuration.Configuration]
	sm   configuration.Provider[configuration.Secrets]
}

func NewEnvironmentClient(cm configuration.Provider[configuration.Configuration], sm configuration.Provider[configuration.Secrets], schr SchemaClient) *EnvironmentManager {
	return &EnvironmentManager{
		schr: schr,
		cm:   cm,
		sm:   sm,
	}
}

// ConfigList returns the list of configuration values, optionally filtered by module.
func (s *EnvironmentManager) ConfigList(ctx context.Context, req *connect.Request[adminpb.ConfigListRequest]) (*connect.Response[adminpb.ConfigListResponse], error) {
	listing, err := s.cm.List(ctx, req.Msg.GetIncludeValues(), None[string]())
	if err != nil {
		return nil, errors.Wrap(err, "failed to list configs")
	}

	configs := []*adminpb.ConfigListResponse_Config{}
	for _, config := range listing {
		module, ok := config.Module.Get()
		if req.Msg.Module != nil && *req.Msg.Module != "" && module != *req.Msg.Module {
			continue
		}

		ref := config.Name
		if ok {
			ref = fmt.Sprintf("%s.%s", module, config.Name)
		}

		var cv []byte
		if req.Msg.GetIncludeValues() {
			cv = config.Value.MustGet()
		}

		configs = append(configs, &adminpb.ConfigListResponse_Config{
			RefPath: ref,
			Value:   cv,
		})
	}
	return connect.NewResponse(&adminpb.ConfigListResponse{Configs: configs}), nil
}

// ConfigGet returns the configuration value for a given ref string.
func (s *EnvironmentManager) ConfigGet(ctx context.Context, req *connect.Request[adminpb.ConfigGetRequest]) (*connect.Response[adminpb.ConfigGetResponse], error) {
	ref := refFromConfigRef(req.Msg.GetRef())
	vb, err := s.cm.Load(ctx, ref)
	if err != nil {
		return nil, errors.Wrap(err, ref.String())
	}
	return connect.NewResponse(&adminpb.ConfigGetResponse{Value: vb}), nil
}

// ConfigSet sets the configuration at the given ref to the provided value.
func (s *EnvironmentManager) ConfigSet(ctx context.Context, req *connect.Request[adminpb.ConfigSetRequest]) (*connect.Response[adminpb.ConfigSetResponse], error) {
	err := s.validateAgainstSchema(ctx, false, refFromConfigRef(req.Msg.GetRef()), req.Msg.Value)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = s.cm.Store(ctx, refFromConfigRef(req.Msg.GetRef()), req.Msg.Value)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set config")
	}
	return connect.NewResponse(&adminpb.ConfigSetResponse{}), nil
}

// ConfigUnset unsets the config value at the given ref.
func (s *EnvironmentManager) ConfigUnset(ctx context.Context, req *connect.Request[adminpb.ConfigUnsetRequest]) (*connect.Response[adminpb.ConfigUnsetResponse], error) {
	err := s.cm.Delete(ctx, refFromConfigRef(req.Msg.GetRef()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to unset config")
	}
	return connect.NewResponse(&adminpb.ConfigUnsetResponse{}), nil
}

// SecretsList returns the list of secrets, optionally filtered by module.
func (s *EnvironmentManager) SecretsList(ctx context.Context, req *connect.Request[adminpb.SecretsListRequest]) (*connect.Response[adminpb.SecretsListResponse], error) {
	listing, err := s.sm.List(ctx, req.Msg.GetIncludeValues(), None[string]())
	if err != nil {
		return nil, errors.Wrap(err, "failed to list secrets")
	}
	secrets := []*adminpb.SecretsListResponse_Secret{}
	for _, secret := range listing {
		module, ok := secret.Module.Get()
		if req.Msg.Module != nil && *req.Msg.Module != "" && module != *req.Msg.Module {
			continue
		}
		ref := secret.Name
		if ok {
			ref = fmt.Sprintf("%s.%s", module, secret.Name)
		}
		var sv []byte
		if req.Msg.GetIncludeValues() {
			sv = secret.Value.MustGet()
		}
		secrets = append(secrets, &adminpb.SecretsListResponse_Secret{
			RefPath: ref,
			Value:   sv,
		})
	}
	return connect.NewResponse(&adminpb.SecretsListResponse{Secrets: secrets}), nil
}

// SecretGet returns the secret value for a given ref string.
func (s *EnvironmentManager) SecretGet(ctx context.Context, req *connect.Request[adminpb.SecretGetRequest]) (*connect.Response[adminpb.SecretGetResponse], error) {
	ref := refFromConfigRef(req.Msg.GetRef())
	vb, err := s.sm.Load(ctx, ref)
	if err != nil {
		return nil, errors.Wrap(err, ref.String())
	}
	return connect.NewResponse(&adminpb.SecretGetResponse{Value: vb}), nil
}

// SecretSet sets the secret at the given ref to the provided value.
func (s *EnvironmentManager) SecretSet(ctx context.Context, req *connect.Request[adminpb.SecretSetRequest]) (*connect.Response[adminpb.SecretSetResponse], error) {
	err := s.validateAgainstSchema(ctx, true, refFromConfigRef(req.Msg.GetRef()), req.Msg.Value)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = s.sm.Store(ctx, refFromConfigRef(req.Msg.GetRef()), req.Msg.Value)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set secret")
	}
	return connect.NewResponse(&adminpb.SecretSetResponse{}), nil
}

// SecretUnset unsets the secret value at the given ref.
func (s *EnvironmentManager) SecretUnset(ctx context.Context, req *connect.Request[adminpb.SecretUnsetRequest]) (*connect.Response[adminpb.SecretUnsetResponse], error) {
	err := s.sm.Delete(ctx, refFromConfigRef(req.Msg.GetRef()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to unset secret")
	}
	return connect.NewResponse(&adminpb.SecretUnsetResponse{}), nil
}

// MapConfigsForModule combines all configuration values visible to the module.
func (s *EnvironmentManager) MapConfigsForModule(ctx context.Context, req *connect.Request[adminpb.MapConfigsForModuleRequest]) (*connect.Response[adminpb.MapConfigsForModuleResponse], error) {
	values, err := s.cm.List(ctx, true, Some(req.Msg.Module))
	if err != nil {
		return nil, errors.Wrap(err, "failed to map configs for module")
	}
	return connect.NewResponse(&adminpb.MapConfigsForModuleResponse{
		Values: maps.FromSlice(values, func(value configuration.Value) (string, []byte) {
			return value.Name, value.Value.MustGet()
		}),
	}), nil
}

// MapSecretsForModule combines all secrets visible to the module.
func (s *EnvironmentManager) MapSecretsForModule(ctx context.Context, req *connect.Request[adminpb.MapSecretsForModuleRequest]) (*connect.Response[adminpb.MapSecretsForModuleResponse], error) {
	values, err := s.sm.List(ctx, true, Some(req.Msg.Module))
	if err != nil {
		return nil, errors.Wrap(err, "failed to map secrets for module")
	}
	return connect.NewResponse(&adminpb.MapSecretsForModuleResponse{
		Values: maps.FromSlice(values, func(value configuration.Value) (string, []byte) {
			return value.Name, value.Value.MustGet()
		}),
	}), nil
}

func refFromConfigRef(cr *adminpb.ConfigRef) configuration.Ref {
	return configuration.NewRef(Ptr(cr.Module), cr.Name)
}

func (s *EnvironmentManager) validateAgainstSchema(ctx context.Context, isSecret bool, ref configuration.Ref, value json.RawMessage) error {
	logger := log.FromContext(ctx)

	// Globals aren't in the module schemas, so we have nothing to validate against.
	if !ref.Module.Ok() {
		return nil
	}

	// If we can't retrieve an active schema, skip validation.
	sch, err := s.schr.GetSchema(ctx)
	if err != nil {
		logger.Debugf("skipping validation; could not get the active schema: %v", err)
		return nil
	}

	r := schema.RefKey{Module: ref.Module.Default(""), Name: ref.Name}.ToRef()
	decl, ok := sch.Resolve(r).Get()
	if !ok {
		logger.Debugf("skipping validation; declaration %q not found", ref.Name)
		return nil
	}

	var fieldType schema.Type
	if isSecret {
		decl, ok := decl.(*schema.Secret)
		if !ok {
			return errors.Errorf("%q is not a secret declaration", ref.Name)
		}
		fieldType = decl.Type
	} else {
		decl, ok := decl.(*schema.Config)
		if !ok {
			return errors.Errorf("%q is not a config declaration", ref.Name)
		}
		fieldType = decl.Type
	}

	var v any
	err = encoding.Unmarshal(value, &v)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal JSON value")
	}

	err = schema.ValidateJSONValue(fieldType, []string{ref.Name}, v, sch)
	if err != nil {
		return errors.Wrap(err, "JSON validation failed")
	}

	return nil
}

func (s *EnvironmentManager) GetSchema(ctx context.Context, c *connect.Request[ftlv1.GetSchemaRequest]) (*connect.Response[ftlv1.GetSchemaResponse], error) {
	sch, err := s.schr.GetSchema(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get schema")
	}
	return connect.NewResponse(&ftlv1.GetSchemaResponse{Schema: sch.ToProto()}), nil
}
