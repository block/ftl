package state

import (
	"fmt"

	"github.com/block/ftl/common/schema"
)

// Output states are general outputs that are used to provision a resource.
// They are not specific to a provisioner.

type OutputPostgres struct {
	Module     string
	ResourceID string

	Connector schema.DatabaseConnector
}

func (s OutputPostgres) DebugString() string {
	return fmt.Sprintf("%T{Module: %s, ResourceID: %s}", s, s.Module, s.ResourceID)
}

type OutputMySQL struct {
	Module     string
	ResourceID string

	Connector schema.DatabaseConnector
}

func (s OutputMySQL) DebugString() string {
	return fmt.Sprintf("%T{Module: %s, ResourceID: %s}", s, s.Module, s.ResourceID)
}
