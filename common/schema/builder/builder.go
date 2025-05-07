package builder

import (
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/must"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
)

type SchemaBuilder struct{ s *schema.Schema }

func Schema(realms ...*schema.Realm) *SchemaBuilder {
	return &SchemaBuilder{
		s: &schema.Schema{
			Realms: realms,
		},
	}
}

func (s *SchemaBuilder) Realm(realm ...*schema.Realm) *SchemaBuilder {
	s.s.Realms = append(s.s.Realms, realm...)
	return s
}

// MustBuild is a convenience method for tests. DO NOT USE IN PRODUCTION
func (s *SchemaBuilder) MustBuild() *schema.Schema {
	return must.Get(s.Build())
}

func (s *SchemaBuilder) Build() (*schema.Schema, error) {
	return errors.WithStack2(s.s.Validate())
}

type RealmBuilder struct {
	r *schema.Realm
}

func Realm(name string, modules ...*schema.Module) *RealmBuilder {
	return &RealmBuilder{
		r: &schema.Realm{
			Name:    name,
			Modules: modules,
		},
	}
}

func (r *RealmBuilder) External(external bool) *RealmBuilder {
	r.r.External = external
	return r
}

func (r *RealmBuilder) Module(module ...*schema.Module) *RealmBuilder {
	r.r.Modules = append(r.r.Modules, module...)
	return r
}

// MustBuild is a convenience method for tests. DO NOT USE IN PRODUCTION
func (r *RealmBuilder) MustBuild() *schema.Realm {
	return must.Get(r.Build())
}

func (r *RealmBuilder) Build() (*schema.Realm, error) {
	// TODO: Validate Realm
	return r.r, nil
}

type ModuleBuilder struct {
	m *schema.Module
}

func Module(name string) *ModuleBuilder {
	return &ModuleBuilder{
		m: &schema.Module{
			Name: name,
		},
	}
}

func (m *ModuleBuilder) Pos(pos schema.Position) *ModuleBuilder {
	m.m.Pos = pos
	return m
}

func (m *ModuleBuilder) Comment(comments ...string) *ModuleBuilder {
	m.m.Comments = append(m.m.Comments, comments...)
	return m
}

func (m ModuleBuilder) Runtime(runtime *schema.ModuleRuntime) *ModuleBuilder {
	m.m.Runtime = runtime
	return &m
}

func (m *ModuleBuilder) DeploymentKey(deploymentKey key.Deployment) *ModuleBuilder {
	if m.m.Runtime == nil {
		m.m.Runtime = &schema.ModuleRuntime{}
	}
	if m.m.Runtime.Deployment == nil {
		m.m.Runtime.Deployment = &schema.ModuleRuntimeDeployment{}
	}
	m.m.Runtime.Deployment.DeploymentKey = deploymentKey
	return m
}

func (m *ModuleBuilder) Decl(decl ...schema.Decl) *ModuleBuilder {
	m.m.Decls = append(m.m.Decls, decl...)
	return m
}

func (m *ModuleBuilder) Metadata(meta ...schema.Metadata) *ModuleBuilder {
	m.m.Metadata = append(m.m.Metadata, meta...)
	return m
}

// MustBuild is a convenience for tests. DO NOT USE IN PRODUCTION
func (m *ModuleBuilder) MustBuild() *schema.Module { return must.Get(m.Build()) }

func (m *ModuleBuilder) Build() (*schema.Module, error) {
	err := m.m.Validate()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return m.m, nil
}
