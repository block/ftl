package schema

// ModuleRefKey is a map key for a module in a realm.
// TODO: remove. Assume internal for now
type ModuleRefKey struct {
	Realm  string `parser:"(@Ident '.')?"`
	Module string `parser:"@Ident"`
}
