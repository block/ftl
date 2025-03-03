package state

import (
	"fmt"
)

// Input states are general inputs that are used to provision a resource.
// They are not specific to a provisioner.

type InputPostgres struct {
	ResourceID string
	Cluster    string
	Module     string
}

func (s InputPostgres) DebugString() string {
	return fmt.Sprintf("%T{ResourceID: %s, Cluster: %s, Module: %s}", s, s.ResourceID, s.Cluster, s.Module)
}

type InputMySQL struct {
	ResourceID string
	Cluster    string
	Module     string
}

func (s InputMySQL) DebugString() string {
	return fmt.Sprintf("%T{ResourceID: %s, Cluster: %s, Module: %s}", s, s.ResourceID, s.Cluster, s.Module)
}

type InputTopic struct {
	Topic      string
	Module     string
	Partitions int
}

func (s InputTopic) DebugString() string {
	return fmt.Sprintf("%T{Topic: %s, Module: %s}", s, s.Topic, s.Module)
}
