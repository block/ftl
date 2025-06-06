//nolint:nakedret
package schema

import (
	"go/token"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/cron"
	ftlerrors "github.com/block/ftl/common/errors"
	ftlreflect "github.com/block/ftl/common/reflect"
	islices "github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/egress"
)

var (
	// Primitive types can't be used as identifiers.
	//
	// We don't need Array/Map/Ref here because there are no
	// keywords associated with these types.
	primitivesScope = Scope{
		"Int":    ModuleDecl{Symbol: &Int{}},
		"Float":  ModuleDecl{Symbol: &Float{}},
		"String": ModuleDecl{Symbol: &String{}},
		"Bytes":  ModuleDecl{Symbol: &Bytes{}},
		"Bool":   ModuleDecl{Symbol: &Bool{}},
		"Time":   ModuleDecl{Symbol: &Time{}},
		"Unit":   ModuleDecl{Symbol: &Unit{}},
		"Any":    ModuleDecl{Symbol: &Any{}},
	}
)

type ValidatedNode interface {
	Node
	Validate() error
}

// MustValidate panics if a schema is invalid.
//
// This is useful for testing.
func MustValidate(schema *Schema) *Schema {
	clone, err := schema.Validate()
	if err != nil {
		panic(err)
	}
	return clone
}

// Validate Schema clones, normalises and semantically validates a schema.
func (s *Schema) Validate() (*Schema, error) {
	return errors.WithStack2(ValidateModuleInSchema(s, optional.None[*Module]()))
}

// Validate Realm clones, normalises and semantically validates a realm.
func (r *Realm) Validate() (*Realm, error) {
	return errors.WithStack2(ValidateModuleInRealm(r, optional.None[*Module]()))
}

// ValidateModuleInSchema clones and normalises a schema and semantically validates a single module in it's internal realm.
// m can be a new or updated module that will be added to the schema before validation (in the internal realm).
func ValidateModuleInSchema(original *Schema, m optional.Option[*Module]) (*Schema, error) {
	schema := &Schema{Pos: original.Pos}

	realms := map[string]bool{}
	merr := []error{}
	moduleAdded := false
	for _, realm := range original.Realms {
		if _, ok := realms[realm.Name]; ok {
			merr = append(merr, errorf(realm, "duplicate realm %q", realm.Name))
		}
		realms[realm.Name] = true
		var mod optional.Option[*Module]
		if !moduleAdded && !realm.External {
			if v, ok := m.Get(); ok {
				mod = optional.Some(v)
				moduleAdded = true
			}
		}
		validatedRealm, err := ValidateModuleInRealm(realm, mod)
		if err != nil {
			merr = append(merr, err)
		}
		schema.Realms = append(schema.Realms, validatedRealm)
	}
	return schema, errors.WithStack(errors.Join(merr...))
}

// ValidateModuleInRealm clones and normalises a realm and semantically validates a single module within it.
// If no module is provided, all modules in the realm are validated.
// m can be a new or updated module that will be added to the realm before validation.
func ValidateModuleInRealm(realm *Realm, m optional.Option[*Module]) (*Realm, error) { //nolint:maintidx
	realm = ftlreflect.DeepCopy(realm)

	if m, ok := m.Get(); ok {
		// Replace original version of module with new version in case they differ
		var found bool
		for i, module := range realm.Modules {
			if module.Name == m.Name {
				realm.Modules[i] = m
				found = true
			}
		}
		if !found {
			realm.Modules = append(realm.Modules, m)
		}
	}

	modules := map[string]bool{}
	merr := []error{}
	ingress := map[string]*Verb{}

	// Inject builtins.
	builtins := Builtins()
	// Move builtins to the front of the list.
	realm.Modules = slices.DeleteFunc(realm.Modules, func(m *Module) bool { return m.Name == builtins.Name })
	realm.Modules = append([]*Module{builtins}, realm.Modules...)

	scopes := NewScopes()

	// Validate dependencies (adapted for a single realm)
	if err := validateRealmDependencies(realm); err != nil {
		merr = append(merr, err)
	}

	// First pass, add all the modules.
	for _, module := range realm.Modules {
		if module == builtins {
			continue
		}
		if err := scopes.Add(optional.None[*Module](), module.Name, module); err != nil {
			merr = append(merr, err)
		}
	}

	// Validate modules.
	for _, module := range realm.Modules {
		// Skip builtin module, it's already been validated.
		if module.Name == "builtin" {
			continue
		}
		if v, ok := m.Get(); ok && v.Name != module.Name {
			// Don't validate other modules when validating a single module.
			continue
		}

		if _, seen := modules[module.Name]; seen {
			merr = append(merr, errorf(module, "duplicate module %q", module.Name))
		}
		modules[module.Name] = true
		if err := module.Validate(); err != nil {
			merr = append(merr, err)
		}

		indent := 0
		err := Visit(module, func(n Node, next func() error) error {
			save := scopes
			if scoped, ok := n.(Scoped); ok {
				scopes = scopes.PushScope(scoped.Scope())
				defer func() { scopes = save }()
			}
			indent++
			defer func() { indent-- }()

			// Validate all children before validating the current node.
			if err := next(); err != nil {
				return errors.WithStack(err)
			}

			if n, ok := n.(ValidatedNode); ok {
				if err := n.Validate(); err != nil {
					merr = append(merr, err)
				}
			}

			switch n := n.(type) {
			case *Ref:
				mdecl := scopes.Resolve(*n)
				if mdecl == nil {
					merr = append(merr, errorf(n, "unknown reference %q, is the type annotated and exported?", n))
					break
				}

				if decl, ok := mdecl.Symbol.(Decl); ok {
					if mod, ok := mdecl.Module.Get(); ok {
						n.Module = mod.Name
					}

					if n.Module != module.Name && !decl.GetVisibility().Exported() {
						merr = append(merr, errorf(n, "%s %q must be exported", typeName(decl), n.String()))
					}

					if dataDecl, ok := decl.(*Data); ok {
						if len(n.TypeParameters) != len(dataDecl.TypeParameters) {
							merr = append(merr, errorf(n, "reference to data structure %s has %d type parameters, but %d were expected",
								n.Name, len(n.TypeParameters), len(dataDecl.TypeParameters)))
						}
					} else if len(n.TypeParameters) != 0 && !decl.GetVisibility().Exported() {
						merr = append(merr, errorf(n, "reference to %s %q cannot have type parameters", typeName(decl), n.Name))
					}
				} else {
					if _, ok := mdecl.Symbol.(*TypeParameter); !ok {
						merr = append(merr, errorf(n, "invalid reference %q at %q", n, mdecl.Symbol.Position()))
					}
				}

			case *Verb:
				hasSubscribed := false
				// if contains MetadataSQLQuery, must also contain MetadataDatabases
				isSQLQuery := false
				injectsTransactions := false
				for _, md := range n.Metadata {
					switch md := md.(type) {
					case *MetadataSubscriber:
						if hasSubscribed {
							merr = append(merr, errorf(md, "verb can not subscribe to multiple topics"))
							continue
						}
						hasSubscribed = true
						subErrs := validateVerbSubscriptions(module, n, md, scopes)
						merr = append(merr, subErrs...)

					case *MetadataIngress:
						// Check for duplicate ingress keys
						key := md.Method + " "
						for _, path := range md.Path {
							switch path := path.(type) {
							case *IngressPathLiteral:
								key += "/" + path.Text

							case *IngressPathParameter:
								key += "/{}"
							}
						}
						if existing, ok := ingress[key]; ok {
							merr = append(merr, errorf(md, "duplicate %s ingress %s for %s:%q and %s:%q", md.Type, key, existing.Pos, existing.Name, n.Pos, n.Name))
						}
						ingress[key] = n

					case *MetadataRetry:
						validateRetries(module, md, optional.Some(n.Request), scopes, true)

					case *MetadataSQLQuery:
						isSQLQuery = true

					case *MetadataCalls:
						for _, call := range md.Calls {
							resolved := scopes.Resolve(*call)
							if resolved == nil {
								merr = append(merr, errorf(call, "unknown call %q", call))
								continue
							}
							verb, ok := resolved.Symbol.(*Verb)
							if !ok {
								merr = append(merr, errorf(call, "expected verb, got %T", resolved.Symbol))
								continue
							}
							if verb.IsTransaction() {
								injectsTransactions = true
							}
						}
					case *MetadataCronJob, *MetadataConfig, *MetadataDatabases, *MetadataAlias, *MetadataTypeMap,
						*MetadataEncoding, *MetadataSecrets, *MetadataPublisher, *MetadataSQLMigration, *MetadataArtefact,
						*MetadataSQLColumn, DatabaseConnector, *MetadataGenerated, *MetadataGit, *MetadataFixture,
						*MetadataTransaction, *MetadataEgress, *MetadataPartitions, *MetadataImage:
					}

					merr = append(merr, validateVisibility(scopes, n.Visibility, RefKey{Module: module.Name, Name: n.GetName()}, n.Request, n.Response)...)
				}
				if isSQLQuery {
					dbSet := n.ResolveDatabaseUses(realm, module.Name)
					numDbs := dbSet.Cardinality()
					if isSQLQuery && numDbs == 0 {
						merr = append(merr, errorf(n, "query verb must specify a corresponding datasource"))
					}
					if isSQLQuery && numDbs > 1 {
						merr = append(merr, errorf(n, "query verbs can only access a single datasource"))
					}
				}
				if n.IsTransaction() {
					dbSet := n.ResolveDatabaseUses(realm, module.Name)
					numDbs := dbSet.Cardinality()
					if numDbs == 0 {
						merr = append(merr, errorf(n, "transaction verbs must access a datasource"))
					}
					if numDbs > 1 {
						merr = append(merr, errorf(n, "transaction verbs can only access a single datasource"))
					}
					if injectsTransactions {
						merr = append(merr, errorf(n, "transaction verbs cannot inject nested transactions"))
					}
					for _, verbRef := range n.ResolveCalls(realm, module.Name).ToSlice() {
						if verbRef.Module != module.Name {
							merr = append(merr, errorf(n, "transaction verbs cannot call verbs in external modules; %s.%s calls %s.%s", module.Name, n.Name, verbRef.Module, verbRef.Name))
						}
					}
				}

			case *Database:
				found := false
				for _, md := range n.Metadata {
					switch md := md.(type) {
					case *MetadataSQLMigration:
						if found {
							merr = append(merr, errors.Errorf("database %q has multiple migration metadata", n.Name))
						}
						found = true
					default:
						merr = append(merr, errors.Errorf("metadata %q is not valid on databases", strings.TrimSpace(md.String())))
					}
				}
			case *Enum:
				if n.IsValueEnum() {
					for _, v := range n.Variants {
						expected := resolveType(realm, v.Value.schemaValueType())
						actual := resolveType(realm, n.Type)
						if reflect.TypeOf(expected) != reflect.TypeOf(actual) {
							merr = append(merr, errorf(v, "enum variant %q of type %s cannot have a value of "+
								"type %q", v.Name, n.Type, v.Value.schemaValueType()))
						}
					}
				} else {
					if len(n.Variants) == 0 {
						merr = append(merr, errorf(n, "enum %q must have at least one variant", n.Name))
					}
					for _, v := range n.Variants {
						if _, ok := v.Value.(*TypeValue); !ok {
							merr = append(merr, errorf(v, "type enum variant %q value must be a type, was %T",
								v.Name, n))
						}
					}
				}
				return errors.WithStack(next())

			case IngressPathComponent, Metadata, Value, Type, DatabaseConnector,
				*Module, *Optional, *Schema, *TypeAlias, *String, *Time, *Unit, *Any, *TypeParameter,
				*EnumVariant, *Config, *Secret, *Topic, *DatabaseRuntime, *DatabaseRuntimeConnections,
				*Data, *Field, *MetadataPartitions, *MetadataSQLQuery, *MetadataSQLColumn, *Realm:
				// no-op
			}
			// Declared types must have all child refs maintain at least the same level of visibility.
			if n, ok := n.(Type); ok {
				if n, ok := n.(Decl); ok {
					merr = append(merr, validateVisibility(scopes, n.GetVisibility(), RefKey{Module: module.Name, Name: n.GetName()}, n.schemaChildren()...)...)
				}
			}
			return nil
		})
		if err != nil {
			merr = append(merr, err)
		}
	}

	merr = cleanErrors(merr)
	return realm, errors.WithStack(errors.Join(merr...))
}

var validNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ValidateName validates an FTL name.
func ValidateName(name string) bool {
	return validNameRe.MatchString(name)
}

// ValidateModuleName validates an FTL module name.
func ValidateModuleName(name string) bool {
	if token.Lookup(name).IsKeyword() {
		return false
	}
	return ValidateName(name)
}

// Validate performs the subset of semantic validation possible on a single module.
//
// It ignores references to other modules.
func (m *Module) Validate() error {
	merr := []error{}

	scopes := NewScopes()

	if !ValidateModuleName(m.Name) {
		merr = append(merr, errorf(m, "module name %q is invalid", m.Name))
	}
	if m.Builtin && m.Name != "builtin" {
		merr = append(merr, errorf(m, "the \"builtin\" module is reserved"))
	}
	if err := scopes.Add(optional.None[*Module](), m.Name, m); err != nil {
		merr = append(merr, err)
	}
	scopes = scopes.Push()

	generateDeadLetterTopics(m)

	// Key is <type>:<name>
	duplicateDecls := map[string]Decl{}

	_ = Visit(m, func(n Node, next func() error) error { //nolint:errcheck
		if scoped, ok := n.(Scoped); ok {
			pop := scopes
			scopes = scopes.PushScope(scoped.Scope())
			defer func() { scopes = pop }()
		}

		// Validate all children before validating the current node.
		if err := next(); err != nil {
			return errors.WithStack(err)
		}

		if n, ok := n.(Decl); ok {
			tname := typeName(n)
			duplKey := tname + ":" + n.GetName()
			if dupl, ok := duplicateDecls[duplKey]; ok {
				merr = append(merr, errorf(n, "duplicate %s %q, first defined at %s", tname, n.GetName(), dupl.Position()))
			} else {
				duplicateDecls[duplKey] = n
			}
			if !ValidateName(n.GetName()) {
				merr = append(merr, errorf(n, "%s name %q is invalid", tname, n.GetName()))
			} else if _, ok := primitivesScope[n.GetName()]; ok {
				merr = append(merr, errorf(n, "%s name %q is a reserved word", tname, n.GetName()))
			}
		}
		switch n := n.(type) {
		case *Ref:
			mdecl := scopes.Resolve(*n)
			if mdecl == nil && n.Module == "" {
				// Fully qualify refs with module == "" to the current module if it can be resolved.
				if internalDecl := scopes.Resolve(Ref{Module: m.Name, Name: n.Name, TypeParameters: n.TypeParameters}); internalDecl != nil {
					n.Module = m.Name
					mdecl = scopes.Resolve(*n)
				}
			}
			if mdecl != nil {
				switch decl := mdecl.Symbol.(type) {
				case *Verb, *Enum, *Database, *Config, *Secret, *Topic:
					if len(n.TypeParameters) != 0 {
						merr = append(merr, errorf(n, "reference to %s %q cannot have type parameters", typeName(decl), n.Name))
					}
				case *Data:
					if len(n.TypeParameters) != len(decl.TypeParameters) {
						merr = append(merr, errorf(n, "reference to data structure %s has %d type parameters, but %d were expected",
							n.Name, len(n.TypeParameters), len(decl.TypeParameters)))
					}
				default:
				}
			}

		case *Verb:
			n.SortMetadata()
			merr = append(merr, validateVerbMetadata(scopes, m, n)...)

		case *Data:
			for _, md := range n.Metadata {
				switch md.(type) {
				case *MetadataCalls, *MetadataSecrets, *MetadataConfig, *MetadataSQLColumn, *MetadataSQLQuery:
					merr = append(merr, errorf(md, "metadata %q is not valid on data structures", strings.TrimSpace(md.String())))
				default:

				}
			}

		case *Field:
			for _, md := range n.Metadata {
				switch md := md.(type) {
				case *MetadataAlias:
				case *MetadataSQLColumn:
					continue
				default:
					merr = append(merr, errorf(md, "metadata %q is not valid on fields", strings.TrimSpace(md.String())))
				}
			}

		case *Topic:
			// Topic names must:
			// - be idents: this allows us to generate external module files with the variable name as the topic name (with first letter uppercased, so it is visible to the module)
			// - start with a lower case letter: this allows us to deterministically derive the topic name from the generated variable name
			if !ValidateName(n.Name) || !unicode.IsLower(rune(n.Name[0])) {
				merr = append(merr, errorf(n, "invalid name: must consist of only letters, numbers and underscores, and start with a lowercase letter."))
			}

		case *TypeAlias:
			for _, md := range n.Metadata {
				if _, ok := md.(*MetadataTypeMap); !ok {
					merr = append(merr, errorf(md, "metadata %q is not valid on type aliases", strings.TrimSpace(md.String())))
				}
			}

		case Type, Metadata, Value, IngressPathComponent, DatabaseConnector,
			*Module, *Schema, *Optional, *TypeParameter, *EnumVariant,
			*Config, *Secret, *DatabaseRuntime, *DatabaseRuntimeConnections,
			*Database, *Enum, *Realm:
		}
		return nil
	})

	merr = cleanErrors(merr)
	SortModuleDecls(m)
	return errors.WithStack(errors.Join(merr...))
}

// SortModuleDecls sorts the declarations in a module.
func SortModuleDecls(module *Module) {
	sort.SliceStable(module.Decls, func(i, j int) bool {
		iDecl := module.Decls[i]
		jDecl := module.Decls[j]
		iPriority := getDeclSortingPriority(iDecl)
		jPriority := getDeclSortingPriority(jDecl)
		if iPriority == jPriority {
			return iDecl.GetName() < jDecl.GetName()
		}
		return iPriority < jPriority
	})
}

// getDeclSortingPriority (used for schema sorting) is pulled out into it's own switch so the Go sumtype check will fail
// if a new Decl is added but not explicitly prioritized
func getDeclSortingPriority(decl Decl) int {
	priority := 0
	switch decl.(type) {
	case *Config:
		priority = 1
	case *Secret:
		priority = 2
	case *Database:
		priority = 3
	case *Topic:
		priority = 4
	case *TypeAlias:
		priority = 5
	case *Enum:
		priority = 6
	case *Data:
		priority = 7
	case *Verb:
		priority = 8
	}
	return priority
}

func sortMetadata(md []Metadata) {
	sort.SliceStable(md, func(i, j int) bool {
		iMd := md[i]
		jMd := md[j]
		sortMetadataType(iMd)
		sortMetadataType(jMd)
		iPriority := getMetadataSortingPriority(iMd)
		jPriority := getMetadataSortingPriority(jMd)
		return iPriority < jPriority
	})
}

func sortMetadataType(md Metadata) {
	sortRefs := func(refs []*Ref) {
		sort.SliceStable(refs, func(i, j int) bool {
			if refs[i].Module == refs[j].Module {
				return refs[i].Name < refs[j].Name
			}
			return refs[i].Module < refs[j].Module
		})
	}

	switch m := md.(type) {
	case *MetadataAlias:
		return
	case *MetadataCalls:
		sortRefs(m.Calls)
	case *MetadataConfig:
		sortRefs(m.Config)
	case *MetadataCronJob:
		return
	case *MetadataDatabases:
		sortRefs(m.Uses)
	case *MetadataEncoding:
		return
	case *MetadataIngress:
		return
	case *MetadataRetry:
		return
	case *MetadataSecrets:
		sortRefs(m.Secrets)
	case *MetadataSubscriber:
		return
	case *MetadataTypeMap:
		return
	case *MetadataPublisher:
		sortRefs(m.Topics)
	case *MetadataSQLMigration:
		return
	case *MetadataArtefact:
		return
	case *MetadataImage:
		return
	case *MetadataSQLColumn:
		return
	case *MetadataSQLQuery:
		return
	case *MetadataPartitions:
		return
	case *MetadataGenerated:
		return
	case *MetadataGit:
		return
	case *MetadataFixture:
		return
	case *MetadataTransaction:
		return
	case *MetadataEgress:
		slices.Sort(m.Targets)
	}
}

func getMetadataSortingPriority(metadata Metadata) int {
	priority := 0
	switch metadata.(type) {
	case *MetadataIngress:
		priority = 1
	case *MetadataAlias:
		priority = 2
	case *MetadataEncoding:
		priority = 3
	case *MetadataCalls:
		priority = 4
	case *MetadataDatabases:
		priority = 5
	case *MetadataSecrets:
		priority = 6
	case *MetadataConfig:
		priority = 7
	case *MetadataCronJob:
		priority = 8
	case *MetadataPublisher:
		priority = 9
	case *MetadataSubscriber:
		priority = 10
	case *MetadataRetry:
		priority = 11
	case *MetadataTypeMap:
		priority = 12
	case *MetadataSQLMigration:
		priority = 13
	case *MetadataArtefact:
		priority = 14
	case *MetadataSQLColumn:
		priority = 15
	case *MetadataSQLQuery:
		priority = 16
	case *MetadataPartitions:
		priority = 17
	case *MetadataGenerated:
		priority = 18
	case *MetadataGit:
		priority = 19
	case *MetadataFixture:
		priority = 20
	case *MetadataTransaction:
		priority = 21
	case *MetadataEgress:
		priority = 22
	case *MetadataImage:
		priority = 23
	}
	return priority
}

// Sort and de-duplicate errors.
func cleanErrors(merr []error) []error {
	if len(merr) == 0 {
		return nil
	}
	unwrapped := errors.UnwrapAllInnermost(errors.Join(merr...))
	merr = ftlerrors.Deduplicate(unwrapped)
	// Sort by position.
	sort.Slice(merr, func(i, j int) bool {
		var ipe, jpe participle.Error
		if errors.As(merr[i], &ipe) && errors.As(merr[j], &jpe) {
			ipp := ipe.Position()
			jpp := jpe.Position()
			return ipp.Line < jpp.Line || (ipp.Line == jpp.Line && ipp.Column < jpp.Column)
		}
		return merr[i].Error() < merr[j].Error()
	})
	return merr
}

type dependencyVertex struct {
	from, to string
}

type dependencyVertexState int

const (
	notExplored dependencyVertexState = iota
	exploring
	fullyExplored
)

func validateRealmDependencies(realm *Realm) error {
	// go through schema's modules, find cycles in modules' dependencies

	// First pass, set up direct imports and vertex states for each module
	// We need each import array and vertex array to be sorted to make the output deterministic
	imports := map[string][]string{}
	vertexes := []dependencyVertex{}
	vertexStates := map[dependencyVertex]dependencyVertexState{}

	for _, module := range realm.Modules {
		currentImports := module.Imports()
		sort.Strings(currentImports)
		imports[module.Name] = currentImports

		for _, imp := range currentImports {
			v := dependencyVertex{module.Name, imp}
			vertexes = append(vertexes, v)
			vertexStates[v] = notExplored
		}
	}

	sort.Slice(vertexes, func(i, j int) bool {
		lhs := vertexes[i]
		rhs := vertexes[j]
		return lhs.from < rhs.from || (lhs.from == rhs.from && lhs.to < rhs.to)
	})

	// DFS to find cycles
	for _, v := range vertexes {
		if cycle := dfsForDependencyCycle(imports, vertexStates, v); cycle != nil {
			return errors.Errorf("found cycle in dependencies: %s", strings.Join(cycle, " -> "))
		}
	}

	return nil
}

func dfsForDependencyCycle(imports map[string][]string, vertexStates map[dependencyVertex]dependencyVertexState, v dependencyVertex) []string {
	switch vertexStates[v] {
	case notExplored:
		vertexStates[v] = exploring

		for _, toModule := range imports[v.to] {
			nextV := dependencyVertex{v.to, toModule}
			if cycle := dfsForDependencyCycle(imports, vertexStates, nextV); cycle != nil {
				// found cycle. prepend current module to cycle and return
				cycle = append([]string{nextV.from}, cycle...)
				return cycle
			}
		}
		vertexStates[v] = fullyExplored
		return nil
	case exploring:
		return []string{v.to}
	case fullyExplored:
		return nil
	}

	return nil
}

func errorf(pos interface{ Position() Position }, format string, args ...any) error {
	p := pos.Position()
	errPos := builderrors.Position{
		Filename:    p.Filename,
		Line:        p.Line,
		StartColumn: p.Column,
		EndColumn:   p.Column,
	}
	return errors.WithStack(builderrors.Errorf(errPos, format, args...))
}

// validateVisibility checks that child nodes do not have less visibility than their parent.
func validateVisibility(scopes Scopes, requiredVisibility Visibility, parent RefKey, nodes ...Node) (merr []error) {
	if requiredVisibility == VisibilityScopeNone {
		return nil
	}
	for _, n := range nodes {
		switch n := n.(type) {
		case *Ref:
			mdecl := scopes.Resolve(*n)
			if mdecl == nil {
				continue
			}
			if t, ok := mdecl.Symbol.(Type); ok {
				ts := append([]Node{t}, islices.Map(n.TypeParameters, func(t Type) Node { return t })...)
				merr = append(merr, validateVisibility(scopes, requiredVisibility, parent, ts...)...)
			}
		case Decl:
			if n.GetVisibility() < requiredVisibility {
				merr = append(merr, errorf(n, "%s %q must have visibility of %q due to %s", typeName(n), n.GetName(), requiredVisibility.String(), parent))
			}
		default:
			// For built in types, validate their children
			merr = append(merr, validateVisibility(scopes, requiredVisibility, parent, n.schemaChildren()...)...)
		}
	}
	return merr
}

func validateVerbMetadata(scopes Scopes, module *Module, n *Verb) (merr []error) {
	// Validate metadata
	metadataTypes := map[reflect.Type]bool{}
	for _, md := range n.Metadata {
		reflected := reflect.TypeOf(md)
		if _, allowsMultiple := md.(*MetadataSubscriber); !allowsMultiple {
			if _, seen := metadataTypes[reflected]; seen {
				merr = append(merr, errorf(md, "verb can not have multiple instances of %s", strings.ToLower(strings.TrimPrefix(reflected.String(), "*schema.Metadata"))))
				continue
			}
		}
		metadataTypes[reflected] = true

		switch md := md.(type) {
		case *MetadataIngress:
			reqInfo, errs := validateIngressRequest(scopes, module, n, "request", n.Request, md.Method == http.MethodGet)
			merr = append(merr, errs...)
			errs = validateIngressResponse(scopes, module, n, "response", n.Response)
			merr = append(merr, errs...)

			if reqInfo.pathParamSymbol != nil {
				// If this is nil it has already failed validation

				hasParameters := false
				// Validate path
				for _, path := range md.Path {
					switch path := path.(type) {
					case *IngressPathParameter:
						hasParameters = true
						switch dataType := reqInfo.pathParamSymbol.(type) {
						case *Data:
							if dataType.FieldByName(path.Name) == nil {
								merr = append(merr, errorf(path, "ingress verb %s: request pathParameter type %s does not contain a field corresponding to the parameter %q", n.Name, reqInfo.pathParamType, path.Name))
							}
						case *Map:
							if keyType, ok := dataType.Key.(*String); !ok {
								merr = append(merr, errorf(path, "ingress verb %s: request pathParameter map key time type %s does not contain a field corresponding to the parameter %q", n.Name, keyType, path.Name))
							}
						case *String, *Int, *Bool, *Float:
							// Only valid for a single path parameter
							count := 0
							for _, p := range md.Path {
								if _, ok := p.(*IngressPathParameter); ok {
									count++
								}
							}
							if count != 1 {
								merr = append(merr, errorf(path, "ingress verb %s: cannot use path parameter %q with request type %s as it has multiple path parameters, expected Data or Map type", n.Name, path.Name, reqInfo.pathParamType))
							}
						default:
							merr = append(merr, errorf(path, "ingress verb %s: cannot use path parameter %q with request type %s, expected Data or Map type", n.Name, path.Name, reqInfo.pathParamType))
						}
					case *IngressPathLiteral:
					}
				}
				if !hasParameters {
					// We still allow map even with no path parameters
					switch reqInfo.pathParamSymbol.(type) {
					case *Unit, *Map:
					default:
						merr = append(merr, errorf(reqInfo.pathParamSymbol, "ingress verb %s: cannot use path parameter type %s, expected Unit or Map as ingress has no path parameters", n.Name, reqInfo.pathParamType))
					}
				}

			}
		case *MetadataCronJob:
			_, err := cron.Parse(md.Cron)
			if err != nil {
				merr = append(merr, errorf(md, "verb %s: invalid cron expression %q: %v", n.Name, md.Cron, err))
			}
			if _, ok := n.Request.(*Unit); !ok {
				merr = append(merr, errorf(md, "verb %s: cron job can not have a request type", n.Name))
			}
			if _, ok := n.Response.(*Unit); !ok {
				merr = append(merr, errorf(md, "verb %s: cron job can not have a response type", n.Name))
			}
		case *MetadataRetry:
			// Only allow retries on pubsub subscribers for now
			_, isSubscriber := islices.FindVariant[*MetadataSubscriber](n.Metadata)
			if !isSubscriber {
				merr = append(merr, errorf(md, `retries can only be added to subscribers`))
				return
			}

			subErrs := validateRetries(module, md, optional.Some(n.Request), scopes, false)
			merr = append(merr, subErrs...)

		case *MetadataSubscriber:
			subErrs := validateVerbSubscriptions(module, n, md, scopes)
			merr = append(merr, subErrs...)
		case *MetadataSQLColumn:
			merr = append(merr, errorf(md, "metadata %q is not valid on verbs", strings.TrimSpace(md.String())))

		case *MetadataEgress:
			config := islices.FilterVariants[*Config, Decl](module.Decls)
			configNames := map[string]bool{}
			for i := range config {
				configNames[i.Name] = true
			}
			for _, target := range md.Targets {
				extract := egress.ExtractConfigItems(target)
				for _, item := range extract {
					if !configNames[item] {
						merr = append(merr, errorf(md, "egress target %q references unknown config %q", target, item))
					}
				}
			}
		case *MetadataCalls, *MetadataConfig, *MetadataDatabases, *MetadataAlias, *MetadataTypeMap, *MetadataEncoding,
			*MetadataSecrets, *MetadataPublisher, *MetadataSQLMigration, *MetadataArtefact, *MetadataSQLQuery, *MetadataPartitions, *MetadataGenerated,
			*MetadataGit, *MetadataFixture, *MetadataTransaction, *MetadataImage:
		}
	}
	return
}

type httpRequestExtractedTypes struct {
	fieldType        Type
	body             Symbol
	pathParamType    Type
	pathParamSymbol  Symbol
	queryParamType   Type
	queryParamSymbol Symbol
}

func validateIngressResponse(scopes Scopes, module *Module, n *Verb, reqOrResp string, r Type) (merr []error) {
	data, err := resolveValidIngressReqResp(scopes, reqOrResp, optional.None[*ModuleDecl](), r, nil)
	if err != nil {
		merr = append(merr, errorf(r, "ingress verb %s: %s type %s: %v", n.Name, reqOrResp, r, err))
		return
	}
	resp, ok := data.Get()
	if !ok {
		merr = append(merr, errorf(r, "ingress verb %s: %s type %s must be builtin.HttpResponse", n.Name, reqOrResp, r))
		return
	}

	scopes = scopes.PushScope(resp.Scope())

	_, _, merr = validateParam(resp, "body", scopes, module, n, reqOrResp, r, validateBodyPayloadType)
	return
}

func validateIngressRequest(scopes Scopes, module *Module, n *Verb, reqOrResp string, r Type, getRequest bool) (result httpRequestExtractedTypes, merr []error) {
	data, err := resolveValidIngressReqResp(scopes, reqOrResp, optional.None[*ModuleDecl](), r, nil)
	if err != nil {
		merr = append(merr, errorf(r, "ingress verb %s: %s type %s: %v", n.Name, reqOrResp, r, err))
		return
	}
	resp, ok := data.Get()
	isRequest := reqOrResp == "request"
	if !ok {
		if isRequest {
			merr = append(merr, errorf(r, "ingress verb %s: %s type %s must be builtin.HttpRequest", n.Name, reqOrResp, r))
		} else {
			merr = append(merr, errorf(r, "ingress verb %s: %s type %s must be builtin.HttpResponse", n.Name, reqOrResp, r))
		}
		return
	}

	scopes = scopes.PushScope(resp.Scope())

	var errs []error
	if getRequest {
		result.fieldType, result.body, errs = validateParam(resp, "body", scopes, module, n, reqOrResp, r, requireUnitPayloadType)
		merr = append(merr, errs...)
	} else {
		result.fieldType, result.body, errs = validateParam(resp, "body", scopes, module, n, reqOrResp, r, validateBodyPayloadType)
		merr = append(merr, errs...)
	}
	if isRequest {
		result.pathParamType, result.pathParamSymbol, errs = validateParam(resp, "pathParameters", scopes, module, n, reqOrResp, r, validatePathParamsPayloadType)
		merr = append(merr, errs...)

		result.queryParamType, result.queryParamSymbol, errs = validateParam(resp, "query", scopes, module, n, reqOrResp, r, validateQueryParamsPayloadType)
		merr = append(merr, errs...)
	}
	return
}

func validateParam(resp *Data, paramName string, scopes Scopes, module *Module, n *Verb, reqOrResp string, r Type, validationFunc func(Node, Type, *Verb, string) error) (fieldType Type, body Symbol, merr []error) {
	fieldType = resp.FieldByName(paramName).Type
	if opt, ok := fieldType.(*Optional); ok {
		fieldType = opt.Type
	}

	if ref, err := ParseRef(fieldType.String()); err == nil {
		if ref.Module != "" && ref.Module != module.Name {
			return // ignores references to other modules.
		}
	}

	bodySym := scopes.ResolveType(fieldType)
	if bodySym == nil {
		merr = append(merr, errorf(r, "ingress verb %s: couldn't resolve %s body type %s", n.Name, reqOrResp, fieldType))
		return
	}
	body = bodySym.Symbol
	err := validationFunc(bodySym.Symbol, r, n, reqOrResp)
	if err != nil {
		merr = append(merr, err)
	}
	return
}

func resolveValidIngressReqResp(scopes Scopes, reqOrResp string, moduleDecl optional.Option[*ModuleDecl], n Node, parent Node) (optional.Option[*Data], error) {
	switch t := n.(type) {
	case *Ref:
		m := scopes.Resolve(*t)
		if m == nil {
			return optional.None[*Data](), errors.Errorf("unknown reference %v", t)
		}
		sym := m.Symbol
		return errors.WithStack2(resolveValidIngressReqResp(scopes, reqOrResp, optional.Some(m), sym, n))
	case *Data:
		md, ok := moduleDecl.Get()
		if !ok {
			return optional.None[*Data](), nil
		}

		m, ok := md.Module.Get()
		if !ok {
			return optional.None[*Data](), nil
		}

		if parent == nil || m.Name != "builtin" || t.Name != "Http"+strings.Title(reqOrResp) {
			return optional.None[*Data](), nil
		}

		ref, ok := parent.(*Ref)
		if !ok {
			return optional.None[*Data](), nil
		}

		result, err := t.Monomorphise(ref)
		if err != nil {
			return optional.None[*Data](), errors.WithStack(err)
		}

		return optional.Some(result), nil
	case *TypeAlias:
		return errors.WithStack2(resolveValidIngressReqResp(scopes, reqOrResp, moduleDecl, t.Type, t))
	default:
		return optional.None[*Data](), nil
	}
}

func validateBodyPayloadType(n Node, r Type, v *Verb, reqOrResp string) error {
	switch t := n.(type) {
	case *Bytes, *String, *Data, *Unit, *Float, *Int, *Bool, *Map, *Array: // Valid HTTP response payload types.
	case *TypeAlias:
		// allow aliases of external types
		for _, m := range t.Metadata {
			if _, ok := m.(*MetadataTypeMap); ok {
				return nil
			}
		}
		return errors.WithStack(validateBodyPayloadType(t.Type, r, v, reqOrResp))
	case *Enum:
		// Type enums are valid but value enums are not.
		if t.IsValueEnum() {
			return errors.WithStack(errorf(r, "ingress verb %s: %s type %s must have a body of bytes, string, data structure, unit, float, int, bool, map, or array not enum %s", v.Name, reqOrResp, r, t.Name))
		}
	default:
		return errors.WithStack(errorf(r, "ingress verb %s: %s type %s must have a body of bytes, string, data structure, unit, float, int, bool, map, or array not %s", v.Name, reqOrResp, r, n))
	}
	return nil
}

func requireUnitPayloadType(n Node, r Type, v *Verb, reqOrResp string) error {
	if _, ok := n.(*Unit); !ok {
		return errors.WithStack(errorf(r, "ingress verb %s: GET request type %s must have a body of unit not %s", v.Name, r, n))

	}
	return nil
}

func validatePathParamsPayloadType(n Node, r Type, v *Verb, reqOrResp string) error {
	switch t := n.(type) {
	case *String, *Data, *Unit, *Float, *Int, *Bool: // Valid HTTP param payload types.
	case *Map:
		if _, ok := t.Value.(*String); !ok {
			return errors.WithStack(errorf(r, "ingress verb %s: %s type %s path params can only support maps of strings, not %v", v.Name, reqOrResp, r, n))
		}
	case *TypeAlias:
		// allow aliases of external types
		for _, m := range t.Metadata {
			if _, ok := m.(*MetadataTypeMap); ok {
				return nil
			}
		}
		return errors.WithStack(validatePathParamsPayloadType(t.Type, r, v, reqOrResp))
	default:
		return errors.WithStack(errorf(r, "ingress verb %s: %s type %s must have a param of data structure, unit or map not %s", v.Name, reqOrResp, r, n))
	}
	return nil
}

func validateQueryParamsPayloadType(n Node, r Type, v *Verb, reqOrResp string) error {
	switch t := n.(type) {
	case *Data, *Unit: // Valid HTTP query payload types.
		return nil
	case *Map:
		switch val := t.Value.(type) {
		case *String:
			// Valid HTTP query payload type
			return nil
		case *Array:
			if _, ok := val.Element.(*String); ok {
				return nil
			}
		default:
			return errors.WithStack(errorf(r, "ingress verb %s: %s type %s query params can only support maps of strings or string arrays, not %v", v.Name, reqOrResp, r, n))
		}
	case *TypeAlias:
		// allow aliases of external types
		for _, m := range t.Metadata {
			if _, ok := m.(*MetadataTypeMap); ok {
				return nil
			}
		}
		return errors.WithStack(validateQueryParamsPayloadType(t.Type, r, v, reqOrResp))
	default:
		return errors.WithStack(errorf(r, "ingress verb %s: %s type %s must have a param of data structure, unit or map not %s", v.Name, reqOrResp, r, n))
	}
	return errors.WithStack(errorf(r, "ingress verb %s: %s type %s query params can only support maps of strings or string arrays, or data types not %v", v.Name, reqOrResp, r, n))
}

func validateVerbSubscriptions(module *Module, v *Verb, md *MetadataSubscriber, scopes Scopes) (merr []error) {
	merr = []error{}
	topicDecl := scopes.Resolve(*md.Topic)
	if topicDecl == nil {
		// errors for invalid refs are handled elsewhere
		return
	}
	topic, ok := topicDecl.Symbol.(*Topic)
	if !ok {
		merr = append(merr, errorf(md, "verb %s: expected topic but found %T for %q", v.Name, topicDecl, md.Topic))
		return
	}

	if !v.Request.Equal(topic.Event) {
		merr = append(merr, errorf(md, "verb %s: request type %v differs from subscription's event type %v", v.Name, v.Request, topic.Event))
	}
	if _, ok := v.Response.(*Unit); !ok {
		merr = append(merr, errorf(md, "verb %s: must be a sink to subscribe but found response type %v", v.Name, v.Response))
	}
	if md.DeadLetter {
		if err := validateDeadLetterTopic(module, v, md); err != nil {
			merr = append(merr, err)
		}
	}
	return merr
}

func validateDeadLetterTopic(module *Module, v *Verb, md *MetadataSubscriber) error {
	ref := &Ref{
		Module: module.Name,
		Name:   DeadLetterNameForSubscriber(v.Name),
	}
	deadLetterDecl, exists := islices.Find(module.Decls, func(d Decl) bool {
		return d.GetName() == ref.Name
	})
	if !exists {
		return errors.WithStack(errorf(md, "could not find dead letter topic %q", DeadLetterNameForSubscriber(v.Name)))
	}
	deadLetterTopic, ok := deadLetterDecl.(*Topic)
	if !ok {
		return errors.WithStack(errorf(md, "expected %v to be a dead letter topic but was already declared at %v", ref, deadLetterDecl.Position()))
	}
	eventType := &Ref{
		Module:         "builtin",
		Name:           "FailedEvent",
		TypeParameters: []Type{v.Request},
	}
	if deadLetterTopic.Event.String() != eventType.String() {
		return errors.WithStack(errorf(md, "dead letter topic %v must have the same event type (%v) as the subscription request type (%v)", ref, deadLetterTopic.Event, eventType))
	}
	// declare that this verb publishes to the dead letter topic
	publishMetadata, ok := islices.FindVariant[*MetadataPublisher](v.Metadata)
	if !ok {
		publishMetadata = &MetadataPublisher{}
		v.Metadata = append(v.Metadata, publishMetadata)
	}
	if _, ok := islices.Find(publishMetadata.Topics, func(r *Ref) bool {
		return r.Name == ref.Name
	}); !ok {
		publishMetadata.Topics = append(publishMetadata.Topics, ref)
	}
	return nil
}

func generateDeadLetterTopics(module *Module) {
	for _, decl := range module.Decls {
		verb, ok := decl.(*Verb)
		if !ok {
			continue
		}
		for _, md := range verb.Metadata {
			sub, ok := md.(*MetadataSubscriber)
			if !ok {
				continue
			}
			if !sub.DeadLetter {
				continue
			}
			ref := Ref{
				Module: module.Name,
				Name:   DeadLetterNameForSubscriber(verb.Name),
			}
			_, exists := islices.Find(module.Decls, func(d Decl) bool {
				return d.GetName() == ref.Name
			})
			if exists {
				// we validate the existing decl later
				continue
			}
			// create a dead letter topic
			deadLetterTopic := &Topic{
				Name: ref.Name,
				Event: &Ref{
					Module:         "builtin",
					Name:           "FailedEvent",
					TypeParameters: []Type{verb.Request},
				},
			}
			module.Decls = append(module.Decls, deadLetterTopic)
		}
	}
}

func validateRetries(module *Module, retry *MetadataRetry, requestType optional.Option[Type], scopes Scopes, fullSchema bool) (merr []error) {
	// Validate count
	if retry.Count != nil && *retry.Count < 0 {
		merr = append(merr, errorf(retry, "retry count can not be negative"))
	}

	// Validate parsing of durations
	retryParams, err := retry.RetryParams()
	if err != nil {
		merr = append(merr, errorf(retry, "%s", err.Error()))
		return
	}
	if retryParams.MaxBackoff < retryParams.MinBackoff {
		merr = append(merr, errorf(retry, "max backoff duration (%s) needs to be at least as long as initial backoff (%s)", retry.MaxBackoff, retry.MinBackoff))
	}

	// validate catch
	if retry.Catch == nil {
		if retryParams.Count == 0 && retry.MinBackoff != "" {
			merr = append(merr, errorf(retry, "can not define a backoff duration when retry count is 0 and no catch is declared"))
		}
		return
	}
	req, ok := requestType.Get()
	if !ok {
		merr = append(merr, errorf(retry, "catch can only be defined on verbs"))
		return
	}
	catchDecl := scopes.Resolve(*retry.Catch)
	if catchDecl == nil {
		if retry.Catch.Module != "" && retry.Catch.Module != module.Name && !fullSchema {
			// can not validate catch ref from external modules until we have the whole schema
			return
		}
		merr = append(merr, errorf(retry, "could not resolve catch verb %q", retry.Catch))
		return
	}
	catchVerb, ok := catchDecl.Symbol.(*Verb)
	if !ok {
		merr = append(merr, errorf(retry, "expected catch to be a verb"))
		return
	}

	hasValidCatchRequest := true
	catchRequestRef, ok := catchVerb.Request.(*Ref)
	if !ok {
		hasValidCatchRequest = false
	} else if !strings.HasPrefix(catchRequestRef.String(), "builtin.CatchRequest") {
		hasValidCatchRequest = false
	} else if len(catchRequestRef.TypeParameters) != 1 {
		hasValidCatchRequest = false
	} else if _, isAny := catchRequestRef.TypeParameters[0].(*Any); !isAny && catchRequestRef.TypeParameters[0].String() != req.String() {
		hasValidCatchRequest = false
	}
	if !hasValidCatchRequest {
		merr = append(merr, errorf(retry, "catch verb must have a request type of builtin.CatchRequest<%s> or builtin.CatchRequest<Any>, but found %v", requestType, catchVerb.Request))
	}

	if _, ok := catchVerb.Response.(*Unit); !ok {
		merr = append(merr, errorf(retry, "catch verb must not have a response type but found %v", catchVerb.Response))
	}
	return merr
}

// Give a type a human-readable name.
func typeName(v any) string {
	return strings.ToLower(reflect.Indirect(reflect.ValueOf(v)).Type().Name())
}

func resolveType(resolver DeclResolver, typ Type) Type {
	ref, ok := typ.(*Ref)
	if !ok {
		return typ
	}
	ropt, _ := resolver.ResolveWithModule(ref)
	resolved, ok := ropt.Get()
	if !ok {
		return typ
	}
	ta, ok := resolved.(*TypeAlias)
	if !ok {
		return typ
	}
	return resolveType(resolver, ta.Type)
}
