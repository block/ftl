package reflection

import (
	"reflect"

	"github.com/alecthomas/types/optional"
)

// singletonTypeRegistry is the global type registry that all public functions in this
// package interface with. It is not truly threadsafe. However, everything is initialized
// in init() calls, which are safe, and the type registry is never mutated afterwards.
var singletonTypeRegistry = newTypeRegistry()

// ResetTypeRegistry clears the contents of the singleton type registry for tests to
// guarantee determinism.
func ResetTypeRegistry() {
	singletonTypeRegistry = newTypeRegistry()
}

// Register applies all the provided options to the singleton TypeRegistry
func Register(options ...Registree) {
	for _, o := range options {
		o(singletonTypeRegistry)
	}
}

// GetVariantByType returns the variant name for the given discriminator and variant type.
func GetVariantByType(discriminator reflect.Type, variantType reflect.Type) optional.Option[string] {
	return singletonTypeRegistry.getVariantByType(discriminator, variantType)
}

// GetVariantByName returns the variant type for the given discriminator and variant name.
func GetVariantByName(discriminator reflect.Type, name string) optional.Option[reflect.Type] {
	return singletonTypeRegistry.getVariantByName(discriminator, name)
}

func GetDatabase[T any]() *ReflectedDatabaseHandle {
	return singletonTypeRegistry.dbHandles[reflect.TypeFor[T]()]
}

func GetAllDatabases() []ReflectedDatabase {
	var dbs []ReflectedDatabase
	for _, db := range singletonTypeRegistry.dbHandles {
		dbs = append(dbs, db.ReflectedDatabase)
	}
	return dbs
}

func IsQueryVerb(ref Ref) bool {
	_, ok := singletonTypeRegistry.queryVerbs[ref]
	return ok
}

func GetTransactionDatabase(ref Ref) optional.Option[ReflectedDatabase] {
	ref, ok := singletonTypeRegistry.transactionVerbs[ref]
	if !ok {
		return optional.None[ReflectedDatabase]()
	}
	db, ok := singletonTypeRegistry.databases[ref]
	if !ok {
		return optional.None[ReflectedDatabase]()
	}
	return optional.Some(db)
}

func GetQueryVerbDatabases() []ReflectedDatabase {
	var dbs []ReflectedDatabase
	for _, db := range singletonTypeRegistry.queryVerbs {
		db, ok := singletonTypeRegistry.databases[db]
		if !ok {
			continue
		}
		dbs = append(dbs, db)
	}
	return dbs
}

// GetDiscriminatorByVariant returns the discriminator type for the given variant type.
func GetDiscriminatorByVariant(variant reflect.Type) optional.Option[reflect.Type] {
	return singletonTypeRegistry.getDiscriminatorByVariant(variant)
}

func CallVerb(ref Ref) VerbExec {
	return singletonTypeRegistry.verbCalls[ref].Exec
}

// IsSumTypeDiscriminator returns true if the given type is a sum type discriminator.
func IsSumTypeDiscriminator(discriminator reflect.Type) bool {
	return singletonTypeRegistry.isSumTypeDiscriminator(discriminator)
}

func IsKnownExternalType(t reflect.Type) bool {
	_, ok := singletonTypeRegistry.externalTypes[t]
	return ok
}
