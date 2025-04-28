package common

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestDirectiveParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected Directive
	}{{
		input:    "ftl:verb",
		expected: &DirectiveVerb{Verb: true, Export: Export{IsExported: false}},
	}, {
		input:    "ftl:verb export",
		expected: &DirectiveVerb{Verb: true, Export: Export{IsExported: true}},
	}, {
		input:    "ftl:verb export:realm",
		expected: &DirectiveVerb{Verb: true, Export: Export{IsExported: true, IsRealmExported: true}},
	}}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			directive, err := DirectiveParser.ParseString("test.go", test.input)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, directive.Directive)
		})
	}
}
