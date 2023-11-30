package schema

// Normalise a Node.
func Normalise[T Node](n T) T {
	var zero Position
	var ni Node = n
	switch c := ni.(type) {
	case *Schema:
		c.Pos = zero
		c.Modules = normaliseSlice(c.Modules)

	case *Module:
		c.Pos = zero
		c.Decls = normaliseSlice(c.Decls)

	case *Array:
		c.Pos = zero
		c.Element = Normalise(c.Element)

	case *Bool:
		c.Bool = false
		c.Pos = zero

	case *Data:
		c.Pos = zero
		c.Fields = normaliseSlice(c.Fields)
		c.Metadata = normaliseSlice(c.Metadata)

	case *DataRef:
		c.Pos = zero

	case *Field:
		c.Pos = zero
		c.Type = Normalise(c.Type)

	case *Float:
		c.Float = false
		c.Pos = zero

	case *Int:
		c.Int = false
		c.Pos = zero

	case *Time:
		c.Time = false
		c.Pos = zero

	case *Map:
		c.Pos = zero
		c.Key = Normalise(c.Key)
		c.Value = Normalise(c.Value)

	case *String:
		c.Str = false
		c.Pos = zero

	case *Verb:
		c.Pos = zero
		c.Request = Normalise(c.Request)
		c.Response = Normalise(c.Response)
		c.Metadata = normaliseSlice(c.Metadata)

	case *VerbRef:
		c.Pos = zero

	case *MetadataCalls:
		c.Pos = zero
		c.Calls = normaliseSlice(c.Calls)

	case *MetadataIngress:
		c.Pos = zero

	case *Optional:
		c.Type = Normalise(c.Type)

	case Decl, Metadata, Type: // Can never occur in reality, but here to satisfy the sum-type check.
		panic("??")
	}
	return ni.(T) //nolint:forcetypeassert
}

func normaliseSlice[T Node](in []T) []T {
	if in == nil {
		return nil
	}
	var out []T
	for _, n := range in {
		out = append(out, Normalise(n))
	}
	return out
}
