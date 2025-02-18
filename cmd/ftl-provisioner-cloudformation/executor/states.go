package executor

import (
	"fmt"

	"github.com/block/ftl/common/schema"
)

type PostgresInputState struct {
	ResourceID string
	Cluster    string
	Module     string
}

func (s PostgresInputState) Kind() ResourceKind {
	return ResourceKindPostgres
}

func (s PostgresInputState) DebugString() string {
	return fmt.Sprintf("PostgresInputState{ResourceID: %s, Cluster: %s, Module: %s}", s.ResourceID, s.Cluster, s.Module)
}

type MySQLInputState struct {
	ResourceID string
	Cluster    string
	Module     string
}

func (s MySQLInputState) Kind() ResourceKind {
	return ResourceKindMySQL
}

func (s MySQLInputState) DebugString() string {
	return fmt.Sprintf("MySQLInputState{ResourceID: %s, Cluster: %s, Module: %s}", s.ResourceID, s.Cluster, s.Module)
}

type PostgresInstanceReadyState struct {
	PostgresInputState

	MasterUserSecretARN string
	WriteEndpoint       string
	ReadEndpoint        string
}

func (s PostgresInstanceReadyState) DebugString() string {
	return fmt.Sprintf("PostgresInstanceReadyState{%s}", s.PostgresInputState.DebugString())
}

type MySQLInstanceReadyState struct {
	MySQLInputState

	MasterUserSecretARN string
	WriteEndpoint       string
	ReadEndpoint        string
}

func (s MySQLInstanceReadyState) DebugString() string {
	return fmt.Sprintf("MySQLInstanceReadyState{%s}", s.MySQLInputState.DebugString())
}

type PostgresDBDoneState struct {
	PostgresInstanceReadyState
	Connector schema.DatabaseConnector
}

func (s PostgresDBDoneState) DebugString() string {
	return fmt.Sprintf("PostgresDBDoneState{%s}", s.PostgresInstanceReadyState.DebugString())
}

type MySQLDBDoneState struct {
	MySQLInstanceReadyState
	Connector schema.DatabaseConnector
}

func (s MySQLDBDoneState) DebugString() string {
	return fmt.Sprintf("MySQLDBDoneState{%s}", s.MySQLInstanceReadyState.DebugString())
}
