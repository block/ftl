package state

import "fmt"

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

	MasterUserSecretARN string
	WriteEndpoint       string
	ReadEndpoint        string
}

func (s RDSInstanceReadyMySQL) DebugString() string {
	return fmt.Sprintf("%T{Module: %s, ResourceID: %s}", s, s.Module, s.ResourceID)
}

type KafkaClusterReady struct {
	InputTopic

	Brokers []string
}

func (s *KafkaClusterReady) DebugString() string {
	return fmt.Sprintf("%T{Topic: %s, Module: %s, Partitions: %d}", s, s.Topic, s.Module, s.Partitions)
}
