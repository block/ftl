package executor

import "fmt"

type PostgresInputState struct {
	ResourceID string
	Cluster    string
	Module     string
}

func (s *PostgresInputState) Stage() Stage {
	return StageProvisioning
}

func (s *PostgresInputState) DebugString() string {
	return fmt.Sprintf("PostgresInputState{ResourceID: %s, Cluster: %s, Module: %s}", s.ResourceID, s.Cluster, s.Module)
}

type MySQLInputState struct {
	ResourceID string
	Cluster    string
	Module     string
}

func (s *MySQLInputState) Stage() Stage {
	return StageProvisioning
}

func (s *MySQLInputState) DebugString() string {
	return fmt.Sprintf("MySQLInputState{ResourceID: %s, Cluster: %s, Module: %s}", s.ResourceID, s.Cluster, s.Module)
}

type PostgresInstanceReadyState struct {
	PostgresInputState

	MasterUserSecretARN string
	WriteEndpoint       string
	ReadEndpoint        string
}

func (s *PostgresInstanceReadyState) DebugString() string {
	return fmt.Sprintf("PostgresInstanceReadyState{%s}", s.PostgresInputState.DebugString())
}

func (s *PostgresInstanceReadyState) Stage() Stage {
	return StageSetup
}

type MySQLInstanceReadyState struct {
	MySQLInputState

	MasterUserSecretARN string
	WriteEndpoint       string
	ReadEndpoint        string
}

func (s *MySQLInstanceReadyState) Stage() Stage {
	return StageSetup
}

func (s *MySQLInstanceReadyState) DebugString() string {
	return fmt.Sprintf("MySQLInstanceReadyState{%s}", s.MySQLInputState.DebugString())
}
