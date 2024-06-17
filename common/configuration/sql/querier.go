// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package sql

import (
	"context"

	"github.com/alecthomas/types/optional"
)

type Querier interface {
	GetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) ([]byte, error)
	ListModuleConfiguration(ctx context.Context) ([]ModuleConfiguration, error)
	SetModuleConfiguration(ctx context.Context, module optional.Option[string], name string, value []byte) error
	UnsetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) error
}

var _ Querier = (*Queries)(nil)
