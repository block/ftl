package schema

import (
	"fmt"
)

//protobuf:6
type DatabaseRuntime struct {
	Connections *DatabaseRuntimeConnections `parser:"" protobuf:"1,optional"`
}

var _ Runtime = (*DatabaseRuntime)(nil)

func (m *DatabaseRuntime) runtimeElement() {}

var _ Symbol = (*DatabaseRuntime)(nil)

func (d *DatabaseRuntime) Position() Position { return d.Connections.Read.Position() }
func (d *DatabaseRuntime) schemaSymbol()      {}
func (d *DatabaseRuntime) String() string {
	return fmt.Sprintf("read: %s, write: %s", d.Connections.Read, d.Connections.Write)
}
func (d *DatabaseRuntime) schemaChildren() []Node {
	return []Node{d.Connections}
}

type DatabaseRuntimeConnections struct {
	Read  DatabaseConnector `parser:"" protobuf:"1"`
	Write DatabaseConnector `parser:"" protobuf:"2"`
}

var _ Symbol = (*DatabaseRuntimeConnections)(nil)

func (d *DatabaseRuntimeConnections) Position() Position { return d.Read.Position() }
func (d *DatabaseRuntimeConnections) schemaSymbol()      {}
func (d *DatabaseRuntimeConnections) String() string {
	return fmt.Sprintf("read: %s, write: %s", d.Read, d.Write)
}

func (d *DatabaseRuntimeConnections) schemaChildren() []Node {
	return []Node{d.Read, d.Write}
}

type DatabaseConnector interface {
	Node

	databaseConnector()
}

//protobuf:1
type DSNDatabaseConnector struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Database string `parser:"" protobuf:"2"`
	DSN      string `parser:"" protobuf:"3"`
}

var _ DatabaseConnector = (*DSNDatabaseConnector)(nil)

func (d *DSNDatabaseConnector) Position() Position     { return d.Pos }
func (d *DSNDatabaseConnector) databaseConnector()     {}
func (d *DSNDatabaseConnector) String() string         { return d.DSN }
func (d *DSNDatabaseConnector) schemaChildren() []Node { return nil }

//protobuf:2
type AWSIAMAuthDatabaseConnector struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Username string `parser:"" protobuf:"2"`
	Endpoint string `parser:"" protobuf:"3"`
	Database string `parser:"" protobuf:"4"`
}

var _ DatabaseConnector = (*AWSIAMAuthDatabaseConnector)(nil)

func (d *AWSIAMAuthDatabaseConnector) Position() Position { return d.Pos }
func (d *AWSIAMAuthDatabaseConnector) databaseConnector() {}
func (d *AWSIAMAuthDatabaseConnector) String() string {
	return fmt.Sprintf("%s@%s/%s", d.Username, d.Endpoint, d.Database)
}

func (d *AWSIAMAuthDatabaseConnector) schemaChildren() []Node { return nil }
