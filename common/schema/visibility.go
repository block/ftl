package schema

import (
	"github.com/alecthomas/errors"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// buf:lint:ignore ENUM_ZERO_VALUE_SUFFIX
type Visibility int

const (
	// VisibilityScopeNone is visible only within the same module
	VisibilityScopeNone Visibility = iota
	// VisibilityScopeModule is visible in all modules in the same realm
	VisibilityScopeModule
	// VisibilityScopeRealm is visible in external realms
	VisibilityScopeRealm
)

func (v Visibility) Exported() bool {
	return v != VisibilityScopeNone
}

func (v Visibility) String() string {
	switch v {
	case VisibilityScopeModule:
		return "export"
	case VisibilityScopeRealm:
		return "export realm"
	default:
		return ""
	}
}

func (v *Visibility) Parse(lex *lexer.PeekingLexer) error {
	type export struct {
		Exported      bool `parser:"@'export'" protobuf:"2"`
		RealmExported bool `parser:"(' ' @'realm')?" protobuf:"3"`
	}
	parser := participle.MustBuild[export]()
	r, err := parser.ParseFromLexer(lex, participle.AllowTrailing(true))
	if err != nil {
		return errors.Wrapf(err, "failed to parse visibility")
	}

	if r.RealmExported {
		*v = VisibilityScopeRealm
	} else if r.Exported {
		*v = VisibilityScopeModule
	} else {
		*v = VisibilityScopeNone
	}
	return nil
}
