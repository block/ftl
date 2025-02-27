// Package types contains examples of FTL types.
package types

//ftl:enum
type NumericEnum int

const (
	NumericEnumOne NumericEnum = iota
	NumericEnumTwo
	NumericEnumThree
)

//ftl:enum
type StringEnum string

const (
	StringEnumOne   StringEnum = "one"
	StringEnumTwo   StringEnum = "two"
	StringEnumThree StringEnum = "three"
)

//ftl:enum
type SumType interface{ sumtype() }

type SumTypeOne struct {
	Value int
}

func (SumTypeOne) sumtype() {}

type SumTypeTwo struct {
	Value string
}

func (SumTypeTwo) sumtype() {}

//ftl:typealias
type UserMap = string
