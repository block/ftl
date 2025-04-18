package state

import (
	"fmt"

	"github.com/alecthomas/types/optional"
)

// AWS states are intermediate provisioner states that are specific to AWS.

type RDSInstanceReadyPostgres struct {
	Module     string
	ResourceID string

	MasterUserSecretARN string
	WriteEndpoint       string
	ReadEndpoint        string
}

func (s RDSInstanceReadyPostgres) DebugString() string {
	return fmt.Sprintf("%T{Module: %s, ResourceID: %s}", s, s.Module, s.ResourceID)
}

type RDSInstanceReadyMySQL struct {
	Module     string
	ResourceID string

	MasterUserSecretARN optional.Option[string]
	WriteEndpoint       string
	ReadEndpoint        string
}

func (s RDSInstanceReadyMySQL) DebugString() string {
	return fmt.Sprintf("%T{Module: %s, ResourceID: %s}", s, s.Module, s.ResourceID)
}

type TopicClusterReady struct {
	InputTopic

	Brokers []string
}

func (s TopicClusterReady) DebugString() string {
	return fmt.Sprintf("%T{Topic: %s, Module: %s, Partitions: %d}", s, s.Topic, s.Module, s.Partitions)
}
