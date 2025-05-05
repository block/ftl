package schema

import (
	"fmt"
	"strings"

	"github.com/alecthomas/types/optional"
	sets "github.com/deckarep/golang-set/v2"

	"github.com/block/ftl/common/slices"
)

//protobuf:2
type Verb struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments   []string   `parser:"@Comment*" protobuf:"2"`
	Visibility Visibility `parser:"@@?" protobuf:"3"`
	Name       string     `parser:"'verb' @Ident" protobuf:"4"`
	Request    Type       `parser:"'(' @@ ')'" protobuf:"5"`
	Response   Type       `parser:"@@" protobuf:"6"`
	Metadata   []Metadata `parser:"@@*" protobuf:"7"`

	Runtime *VerbRuntime `protobuf:"31634,optional" parser:""`
}

var _ Decl = (*Verb)(nil)
var _ Symbol = (*Verb)(nil)
var _ Provisioned = (*Verb)(nil)

// VerbKind is the kind of Verb: verb, sink, source or empty.
type VerbKind string

const (
	// VerbKindVerb is a normal verb taking an input and an output of any non-unit type.
	VerbKindVerb VerbKind = "verb"
	// VerbKindSink is a verb that takes an input and returns unit.
	VerbKindSink VerbKind = "sink"
	// VerbKindSource is a verb that returns an output and takes unit.
	VerbKindSource VerbKind = "source"
	// VerbKindEmpty is a verb that takes unit and returns unit.
	VerbKindEmpty VerbKind = "empty"
)

// Kind returns the kind of Verb this is.
func (v *Verb) Kind() VerbKind {
	_, inIsUnit := v.Request.(*Unit)
	_, outIsUnit := v.Response.(*Unit)
	switch {
	case inIsUnit && outIsUnit:
		return VerbKindEmpty

	case inIsUnit:
		return VerbKindSource

	case outIsUnit:
		return VerbKindSink

	default:
		return VerbKindVerb
	}
}

func (v *Verb) Position() Position { return v.Pos }

func (v *Verb) schemaDecl()   {}
func (v *Verb) schemaSymbol() {}
func (v *Verb) provisioned()  {}
func (v *Verb) schemaChildren() []Node {
	children := []Node{}
	if v.Request != nil {
		children = append(children, v.Request)
	}
	if v.Response != nil {
		children = append(children, v.Response)
	}
	for _, c := range v.Metadata {
		children = append(children, c)
	}
	return children
}

func (v *Verb) GetName() string           { return v.Name }
func (v *Verb) GetVisibility() Visibility { return v.Visibility }

func (v *Verb) IsGenerated() bool {
	_, found := slices.FindVariant[*MetadataGenerated](v.Metadata)
	return found
}

func (v *Verb) IsTransaction() bool {
	_, found := slices.FindVariant[*MetadataTransaction](v.Metadata)
	return found
}

func (v *Verb) IsQuery() bool {
	_, found := slices.FindVariant[*MetadataSQLQuery](v.Metadata)
	return found
}

func (v *Verb) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(v.Comments))
	formatTokens(w,
		v.Visibility.String(),
		"verb",
		fmt.Sprintf("%s(%s)", v.Name, v.Request),
		v.Response.String(),
	)
	fmt.Fprint(w, indent(encodeMetadata(v.Metadata)))
	return w.String()
}

// AddCall adds a call reference to the Verb.
func (v *Verb) AddCall(verb *Ref) {
	v.Metadata = upsert[MetadataCalls](v.Metadata, verb)
}

// AddConfig adds a config reference to the Verb.
func (v *Verb) AddConfig(config *Ref) {
	v.Metadata = upsert[MetadataConfig](v.Metadata, config)
}

// AddSecret adds a config reference to the Verb.
func (v *Verb) AddSecret(secret *Ref) {
	v.Metadata = upsert[MetadataSecrets](v.Metadata, secret)
}

// AddDatabase adds a DB reference to the Verb.
func (v *Verb) AddDatabase(db *Ref) {
	v.Metadata = upsert[MetadataDatabases](v.Metadata, db)
}

func (v *Verb) AddSubscription(sub *MetadataSubscriber) {
	v.Metadata = append(v.Metadata, sub)
}

// AddTopicPublish adds a topic that this Verb publishes to.
func (v *Verb) AddTopicPublish(topic *Ref) {
	v.Metadata = upsert[MetadataPublisher](v.Metadata, topic)
}

func (v *Verb) AddEgress(target string) {
	egress := v.GetMetadataEgress()
	if e, ok := egress.Get(); ok {
		if slices.Contains(e.Targets, target) {
			return
		}
		e.Targets = append(e.Targets, target)
		return
	}
	v.Metadata = append(v.Metadata, &MetadataEgress{
		Pos:     v.Pos,
		Targets: []string{target},
	})

}

func (v *Verb) SortMetadata() {
	sortMetadata(v.Metadata)
}

func (v *Verb) GetMetadataIngress() optional.Option[*MetadataIngress] {
	return optional.From(slices.FindVariant[*MetadataIngress](v.Metadata))
}

func (v *Verb) GetMetadataEgress() optional.Option[*MetadataEgress] {
	return optional.From(slices.FindVariant[*MetadataEgress](v.Metadata))
}
func (v *Verb) GetMetadataCronJob() optional.Option[*MetadataCronJob] {
	return optional.From(slices.FindVariant[*MetadataCronJob](v.Metadata))
}

func (v *Verb) GetProvisioned() ResourceSet {
	var result ResourceSet
	for sub := range slices.FilterVariants[*MetadataSubscriber](v.Metadata) {
		result = append(result, &ProvisionedResource{
			Kind: ResourceTypeSubscription,
			Config: &MetadataSubscriber{
				Topic:      sub.Topic,
				FromOffset: sub.FromOffset,
				DeadLetter: sub.DeadLetter,
			},
		})
	}
	for sub := range slices.FilterVariants[*MetadataFixture](v.Metadata) {
		if !sub.Manual {
			result = append(result, &ProvisionedResource{
				Kind:               ResourceTypeFixture,
				Config:             v,
				DeploymentSpecific: true,
			})
		}
	}
	for sub := range slices.FilterVariants[*MetadataEgress](v.Metadata) {
		for _, target := range sub.Targets {
			result = append(result, &ProvisionedResource{
				Kind:               ResourceTypeEgress,
				Config:             target,
				DeploymentSpecific: true,
			})
		}
	}
	return result
}

func (v *Verb) ResourceID() string {
	return v.Name
}

// GetQuery returns the query metadata for the Verb if it exists. If present, the Verb was generated from SQL.
func (v *Verb) GetQuery() (*MetadataSQLQuery, bool) {
	md, found := slices.FindVariant[*MetadataSQLQuery](v.Metadata)
	if !found {
		return nil, false
	}
	return md, true
}

// ResolveDatabaseUses resolves all datasources accessed by a verb, explicitly or implicitly.
func (v *Verb) ResolveDatabaseUses(resolver DeclResolver, module string) sets.Set[RefKey] {
	dbs := sets.NewSet[RefKey]()
	for _, md := range v.Metadata {
		switch md := md.(type) {
		case *MetadataDatabases:
			for _, db := range md.Uses {
				dbs.Add(db.ToRefKey())
			}
		case *MetadataCalls:
			for _, call := range md.Calls {
				if call.Module == module && call.Name == v.Name {
					continue
				}
				ropt, _ := resolver.ResolveWithModule(call)
				resolved, ok := ropt.Get()
				if !ok {
					continue
				}
				callee, ok := resolved.(*Verb)
				if !ok {
					continue
				}
				dbs = dbs.Union(callee.ResolveDatabaseUses(resolver, module))
			}
		default:
		}
	}
	return dbs
}

// ResolveCalls resolves all verbs called by a verb, explicitly or implicitly.
func (v *Verb) ResolveCalls(resolver DeclResolver, module string) sets.Set[RefKey] {
	verbs := sets.NewSet[RefKey]()
	visited := sets.NewSet[RefKey]()
	for _, md := range v.Metadata {
		switch md := md.(type) {
		case *MetadataCalls:
			for _, call := range md.Calls {
				if call.Module == module && call.Name == v.Name {
					continue
				}
				if visited.Contains(call.ToRefKey()) {
					continue
				}
				visited.Add(call.ToRefKey())
				ropt, _ := resolver.ResolveWithModule(call)
				resolved, ok := ropt.Get()
				if !ok {
					continue
				}
				callee, ok := resolved.(*Verb)
				if !ok {
					continue
				}
				verbs = verbs.Union(callee.ResolveCalls(resolver, module))
			}
		default:
		}
	}
	return verbs
}

// Helper function to insert or update a value in a slice of interfaces
func upsert[V any, W interface {
	*V
	Append(u U)
}, T any, U any](slice []T, u U) (out []T) {
	if c, ok := slices.FindVariant[W](slice); ok {
		c.Append(u)
		return slice
	}
	insert := any(new(V)).(W) //nolint
	insert.Append(u)
	slice = append(slice, any(insert).(T)) //nolint
	return slice
}
