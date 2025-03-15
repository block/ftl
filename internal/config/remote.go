package config

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"

	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/common/slices"
)

const RemoteProviderKind ProviderKind = "remote"

// RemoteProvider provides configuration management of a remote cluster via the FTL admin service.
type RemoteProvider[R Role] struct {
	client adminpbconnect.AdminServiceClient
}

var _ Provider[Configuration] = &RemoteProvider[Configuration]{}

func NewRemoteProviderFactory[R Role](client adminpbconnect.AdminServiceClient) (ProviderKind, Factory[R]) {
	return RemoteProviderKind, func(ctx context.Context, projectRoot string, key ProviderKey) (BaseProvider[R], error) {
		return NewRemoteProvider[R](client), nil
	}
}

func NewRemoteProvider[R Role](client adminpbconnect.AdminServiceClient) *RemoteProvider[R] {
	return &RemoteProvider[R]{
		client: client,
	}
}
func (s *RemoteProvider[R]) Key() ProviderKey                { return NewProviderKey(RemoteProviderKind) }
func (s *RemoteProvider[R]) Role() R                         { var r R; return r }
func (s *RemoteProvider[R]) Close(ctx context.Context) error { return nil }

func (s *RemoteProvider[R]) Delete(ctx context.Context, ref Ref) error {
	var r R
	switch any(r).(type) {
	case Configuration:
		_, err := s.client.ConfigUnset(ctx, connect.NewRequest(&adminpb.ConfigUnsetRequest{
			Ref: &adminpb.ConfigRef{
				Module: ref.Module.Ptr(),
				Name:   ref.Name,
			},
		}))
		return fmt.Errorf("%s delete: %w", r, err)

	case Secrets:
		_, err := s.client.SecretUnset(ctx, connect.NewRequest(&adminpb.SecretUnsetRequest{
			Ref: &adminpb.ConfigRef{
				Module: ref.Module.Ptr(),
				Name:   ref.Name,
			},
		}))
		return fmt.Errorf("%s delete: %w", r, err)

	default:
		panic(fmt.Sprintf("unsupported role %T", r))
	}
}

func (s *RemoteProvider[R]) List(ctx context.Context, withValues bool) ([]Value, error) {
	var r R
	switch any(r).(type) {
	case Configuration:
		resp, err := s.client.ConfigList(ctx, connect.NewRequest(&adminpb.ConfigListRequest{
			IncludeValues: &withValues,
		}))
		if err != nil {
			return nil, fmt.Errorf("%s list: %w", r, err)
		}
		return slices.Map(resp.Msg.Configs, func(config *adminpb.ConfigListResponse_Config) Value {
			return Value{
				Ref:   ParseRef(config.RefPath),
				Value: optional.Zero(config.Value),
			}
		}), nil

	case Secrets:
		resp, err := s.client.SecretsList(ctx, connect.NewRequest(&adminpb.SecretsListRequest{
			IncludeValues: &withValues,
		}))
		if err != nil {
			return nil, fmt.Errorf("%s list: %w", r, err)
		}
		return slices.Map(resp.Msg.Secrets, func(secret *adminpb.SecretsListResponse_Secret) Value {
			return Value{
				Ref:   ParseRef(secret.RefPath),
				Value: optional.Zero(secret.Value),
			}
		}), nil

	default:
		panic(fmt.Sprintf("unsupported role %T", r))
	}
}

func (s *RemoteProvider[R]) Load(ctx context.Context, ref Ref) ([]byte, error) {
	var r R
	switch any(r).(type) {
	case Configuration:
		resp, err := s.client.ConfigGet(ctx, connect.NewRequest(&adminpb.ConfigGetRequest{
			Ref: &adminpb.ConfigRef{
				Module: ref.Module.Ptr(),
				Name:   ref.Name,
			},
		}))
		if err != nil {
			return nil, fmt.Errorf("%s load: %w", r, err)
		}
		return resp.Msg.Value, nil

	case Secrets:
		resp, err := s.client.SecretGet(ctx, connect.NewRequest(&adminpb.SecretGetRequest{
			Ref: &adminpb.ConfigRef{
				Module: ref.Module.Ptr(),
				Name:   ref.Name,
			},
		}))
		if err != nil {
			return nil, fmt.Errorf("%s load: %w", r, err)
		}
		return resp.Msg.Value, nil

	default:
		panic(fmt.Sprintf("unsupported role %T", r))
	}
}

func (s *RemoteProvider[R]) Store(ctx context.Context, ref Ref, value []byte) error {
	var r R
	switch any(r).(type) {
	case Configuration:
		_, err := s.client.ConfigSet(ctx, connect.NewRequest(&adminpb.ConfigSetRequest{
			Ref: &adminpb.ConfigRef{
				Module: ref.Module.Ptr(),
				Name:   ref.Name,
			},
			Value: value,
		}))
		if err != nil {
			return fmt.Errorf("%s store: %w", r, err)
		}
		return nil

	case Secrets:
		_, err := s.client.SecretSet(ctx, connect.NewRequest(&adminpb.SecretSetRequest{
			Ref: &adminpb.ConfigRef{
				Module: ref.Module.Ptr(),
				Name:   ref.Name,
			},
			Value: value,
		}))
		if err != nil {
			return fmt.Errorf("%s store: %w", r, err)
		}
		return nil

	default:
		panic(fmt.Sprintf("unsupported role %T", r))
	}
}
