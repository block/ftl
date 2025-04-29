package common

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/block/ftl/common/schema"
)

func TestDirectiveParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected Directive
	}{{
		input:    "ftl:verb",
		expected: &DirectiveVerb{Verb: true, Visibility: Visibility(schema.VisibilityScopeNone)},
	}, {
		input:    "ftl:verb export",
		expected: &DirectiveVerb{Verb: true, Visibility: Visibility(schema.VisibilityScopeModule)},
	}, {
		input:    "ftl:verb export:realm",
		expected: &DirectiveVerb{Verb: true, Visibility: Visibility(schema.VisibilityScopeRealm)},
	}}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			directive, err := DirectiveParser.ParseString("test.go", test.input)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, directive.Directive)
		})
	}
}
