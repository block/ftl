package admin

import (
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/encoding"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/manager"
	"github.com/block/ftl/internal/log"
)

type SchemaClient interface {
	GetSchema(ctx context.Context) (*schema.Schema, error)
}

// EnvironmentManager is a client that reads and writes secrets and config entries
type EnvironmentManager struct {
	schr SchemaClient
	cm   *manager.Manager[configuration.Configuration]
	sm   *manager.Manager[configuration.Secrets]
}

func NewEnvironmentClient(cm *manager.Manager[configuration.Configuration], sm *manager.Manager[configuration.Secrets], schr SchemaClient) *EnvironmentManager {
	return &EnvironmentManager{
		schr: schr,
		cm:   cm,
		sm:   sm,
	}
}

// ConfigList returns the list of configuration values, optionally filtered by module.
func (s *EnvironmentManager) ConfigList(ctx context.Context, req *connect.Request[ftlv1.ConfigListRequest]) (*connect.Response[ftlv1.ConfigListResponse], error) {
	listing, err := s.cm.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list configs: %w", err)
	}

	configs := []*ftlv1.ConfigListResponse_Config{}
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
		if *req.Msg.IncludeValues {
			var value any
			err := s.cm.Get(ctx, config.Ref, &value)
			if err != nil {
				return nil, fmt.Errorf("failed to get value for %v: %w", ref, err)
			}
			cv, err = json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal value for %s: %w", ref, err)
			}
		}

		configs = append(configs, &ftlv1.ConfigListResponse_Config{
			RefPath: ref,
			Value:   cv,
		})
	}
	return connect.NewResponse(&ftlv1.ConfigListResponse{Configs: configs}), nil
}

// ConfigGet returns the configuration value for a given ref string.
func (s *EnvironmentManager) ConfigGet(ctx context.Context, req *connect.Request[ftlv1.ConfigGetRequest]) (*connect.Response[ftlv1.ConfigGetResponse], error) {
	var value any
	err := s.cm.Get(ctx, refFromConfigRef(req.Msg.GetRef()), &value)
	if err != nil {
		return nil, fmt.Errorf("failed to get from config manager: %w", err)
	}
	vb, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
	}
	return connect.NewResponse(&ftlv1.ConfigGetResponse{Value: vb}), nil
}

// ConfigSet sets the configuration at the given ref to the provided value.
func (s *EnvironmentManager) ConfigSet(ctx context.Context, req *connect.Request[ftlv1.ConfigSetRequest]) (*connect.Response[ftlv1.ConfigSetResponse], error) {
	err := s.validateAgainstSchema(ctx, false, refFromConfigRef(req.Msg.GetRef()), req.Msg.Value)
	if err != nil {
		return nil, err
	}

	err = s.cm.SetJSON(ctx, refFromConfigRef(req.Msg.GetRef()), req.Msg.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to set config: %w", err)
	}
	return connect.NewResponse(&ftlv1.ConfigSetResponse{}), nil
}

// ConfigUnset unsets the config value at the given ref.
func (s *EnvironmentManager) ConfigUnset(ctx context.Context, req *connect.Request[ftlv1.ConfigUnsetRequest]) (*connect.Response[ftlv1.ConfigUnsetResponse], error) {
	err := s.cm.Unset(ctx, refFromConfigRef(req.Msg.GetRef()))
	if err != nil {
		return nil, fmt.Errorf("failed to unset config: %w", err)
	}
	return connect.NewResponse(&ftlv1.ConfigUnsetResponse{}), nil
}

// SecretsList returns the list of secrets, optionally filtered by module.
func (s *EnvironmentManager) SecretsList(ctx context.Context, req *connect.Request[ftlv1.SecretsListRequest]) (*connect.Response[ftlv1.SecretsListResponse], error) {
	listing, err := s.sm.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}
	secrets := []*ftlv1.SecretsListResponse_Secret{}
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
		if *req.Msg.IncludeValues {
			var value any
			err := s.sm.Get(ctx, secret.Ref, &value)
			if err != nil {
				return nil, fmt.Errorf("failed to get value for %v: %w", ref, err)
			}
			sv, err = json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal value for %s: %w", ref, err)
			}
		}
		secrets = append(secrets, &ftlv1.SecretsListResponse_Secret{
			RefPath: ref,
			Value:   sv,
		})
	}
	return connect.NewResponse(&ftlv1.SecretsListResponse{Secrets: secrets}), nil
}

// SecretGet returns the secret value for a given ref string.
func (s *EnvironmentManager) SecretGet(ctx context.Context, req *connect.Request[ftlv1.SecretGetRequest]) (*connect.Response[ftlv1.SecretGetResponse], error) {
	var value any
	err := s.sm.Get(ctx, refFromConfigRef(req.Msg.GetRef()), &value)
	if err != nil {
		return nil, fmt.Errorf("failed to get from secret manager: %w", err)
	}
	vb, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
	}
	return connect.NewResponse(&ftlv1.SecretGetResponse{Value: vb}), nil
}

// SecretSet sets the secret at the given ref to the provided value.
func (s *EnvironmentManager) SecretSet(ctx context.Context, req *connect.Request[ftlv1.SecretSetRequest]) (*connect.Response[ftlv1.SecretSetResponse], error) {
	err := s.validateAgainstSchema(ctx, true, refFromConfigRef(req.Msg.GetRef()), req.Msg.Value)
	if err != nil {
		return nil, err
	}

	err = s.sm.SetJSON(ctx, refFromConfigRef(req.Msg.GetRef()), req.Msg.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to set secret: %w", err)
	}
	return connect.NewResponse(&ftlv1.SecretSetResponse{}), nil
}

// SecretUnset unsets the secret value at the given ref.
func (s *EnvironmentManager) SecretUnset(ctx context.Context, req *connect.Request[ftlv1.SecretUnsetRequest]) (*connect.Response[ftlv1.SecretUnsetResponse], error) {
	err := s.sm.Unset(ctx, refFromConfigRef(req.Msg.GetRef()))
	if err != nil {
		return nil, fmt.Errorf("failed to unset secret: %w", err)
	}
	return connect.NewResponse(&ftlv1.SecretUnsetResponse{}), nil
}

// MapConfigsForModule combines all configuration values visible to the module.
func (s *EnvironmentManager) MapConfigsForModule(ctx context.Context, req *connect.Request[ftlv1.MapConfigsForModuleRequest]) (*connect.Response[ftlv1.MapConfigsForModuleResponse], error) {
	values, err := s.cm.MapForModule(ctx, req.Msg.Module)
	if err != nil {
		return nil, fmt.Errorf("failed to map configs for module: %w", err)
	}
	return connect.NewResponse(&ftlv1.MapConfigsForModuleResponse{Values: values}), nil
}

// MapSecretsForModule combines all secrets visible to the module.
func (s *EnvironmentManager) MapSecretsForModule(ctx context.Context, req *connect.Request[ftlv1.MapSecretsForModuleRequest]) (*connect.Response[ftlv1.MapSecretsForModuleResponse], error) {
	values, err := s.sm.MapForModule(ctx, req.Msg.Module)
	if err != nil {
		return nil, fmt.Errorf("failed to map secrets for module: %w", err)
	}
	return connect.NewResponse(&ftlv1.MapSecretsForModuleResponse{Values: values}), nil
}

func refFromConfigRef(cr *ftlv1.ConfigRef) configuration.Ref {
	return configuration.NewRef(cr.GetModule(), cr.GetName())
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
			return fmt.Errorf("%q is not a secret declaration", ref.Name)
		}
		fieldType = decl.Type
	} else {
		decl, ok := decl.(*schema.Config)
		if !ok {
			return fmt.Errorf("%q is not a config declaration", ref.Name)
		}
		fieldType = decl.Type
	}

	var v any
	err = encoding.Unmarshal(value, &v)
	if err != nil {
		return fmt.Errorf("could not unmarshal JSON value: %w", err)
	}

	err = schema.ValidateJSONValue(fieldType, []string{ref.Name}, v, sch)
	if err != nil {
		return fmt.Errorf("JSON validation failed: %w", err)
	}

	return nil
}

func (s *EnvironmentManager) GetSchema(ctx context.Context, c *connect.Request[ftlv1.GetSchemaRequest]) (*connect.Response[ftlv1.GetSchemaResponse], error) {
	sch, err := s.schr.GetSchema(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}
	return connect.NewResponse(&ftlv1.GetSchemaResponse{Schema: sch.ToProto()}), nil
}
