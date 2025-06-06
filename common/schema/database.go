package schema

import (
	"fmt"
	"strings"

	"github.com/block/ftl/common/slices"
)

const PostgresDatabaseType = "postgres"
const MySQLDatabaseType = "mysql"

//protobuf:3
type Database struct {
	Pos     Position         `parser:"" protobuf:"1,optional"`
	Runtime *DatabaseRuntime `parser:"" protobuf:"31634,optional"`

	Comments []string   `parser:"@Comment*" protobuf:"2"`
	Type     string     `parser:"'database' @('postgres'|'mysql')" protobuf:"4"`
	Name     string     `parser:"@Ident" protobuf:"3"`
	Metadata []Metadata `parser:"@@*" protobuf:"5"`
}

var _ Decl = (*Database)(nil)
var _ Symbol = (*Database)(nil)
var _ Provisioned = (*Database)(nil)

func (d *Database) Position() Position { return d.Pos }
func (*Database) schemaDecl()          {}
func (*Database) schemaSymbol()        {}
func (d *Database) provisioned()       {}
func (d *Database) IsGenerated() bool  { return true }

func (d *Database) schemaChildren() []Node {
	children := []Node{}
	for _, c := range d.Metadata {
		children = append(children, c)
	}
	return children
}
func (d *Database) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(d.Comments))
	fmt.Fprintf(w, "database %s %s", d.Type, d.Name)
	fmt.Fprint(w, indent(encodeMetadata(d.Metadata)))
	return w.String()
}

func (d *Database) GetName() string           { return d.Name }
func (d *Database) GetVisibility() Visibility { return VisibilityScopeNone }

func (d *Database) GetProvisioned() ResourceSet {
	kind := ResourceTypeMysql
	if d.Type == PostgresDatabaseType {
		kind = ResourceTypePostgres
	}
	result := []*ProvisionedResource{{
		Kind:   kind,
		Config: &Database{Type: d.Type},
		State:  d.Runtime,
	}}

	migration, ok := slices.FindVariant[*MetadataSQLMigration](d.Metadata)
	if ok {
		result = append(result, &ProvisionedResource{
			Kind:               ResourceTypeSQLMigration,
			Config:             &Database{Type: d.Type, Metadata: []Metadata{migration}},
			DeploymentSpecific: true,
		})
	}

	return result

}

func (d *Database) ResourceID() string {
	return d.Name
}
