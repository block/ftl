package ingress

import (
	"net/url"
	"testing"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/alecthomas/assert/v2"
)

type obj = map[string]any

func TestMatchAndExtractAllSegments(t *testing.T) {
	tests := []struct {
		pattern  string
		urlPath  string
		expected map[string]string
		matched  bool
	}{
		// valid patterns
		{"", "", map[string]string{}, true},
		{"/", "/", map[string]string{}, true},
		{"/{id}", "/123", map[string]string{"id": "123"}, true},
		{"/{id}/{userId}", "/123/456", map[string]string{"id": "123", "userId": "456"}, true},
		{"/users", "/users", map[string]string{}, true},
		{"/users/{id}", "/users/123", map[string]string{"id": "123"}, true},
		{"/users/{id}", "/users/123", map[string]string{"id": "123"}, true},
		{"/users/{id}/posts/{postId}", "/users/123/posts/456", map[string]string{"id": "123", "postId": "456"}, true},

		// invalid patterns
		{"/", "/users", map[string]string{}, false},
		{"/users/{id}", "/bogus/123", map[string]string{}, false},
	}

	for _, test := range tests {
		actual := make(map[string]string)
		match := matchSegments(test.pattern, test.urlPath, func(segment, value string) {
			actual[segment] = value
		})
		assert.Equal(t, test.matched, match, "pattern = %s, urlPath = %s", test.pattern, test.urlPath)
		assert.Equal(t, test.expected, actual, "pattern = %s, urlPath = %s", test.pattern, test.urlPath)
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		schema  string
		request obj
	}{
		{name: "int", schema: `module test { data Test { intValue Int } }`, request: obj{"intValue": 10.0}},
		{name: "float", schema: `module test { data Test { floatValue Float } }`, request: obj{"floatValue": 10.0}},
		{name: "string", schema: `module test { data Test { stringValue String } }`, request: obj{"stringValue": "test"}},
		{name: "bool", schema: `module test { data Test { boolValue Bool } }`, request: obj{"boolValue": true}},
		{name: "intString", schema: `module test { data Test { intValue Int } }`, request: obj{"intValue": "10"}},
		{name: "floatString", schema: `module test { data Test { floatValue Float } }`, request: obj{"floatValue": "10.0"}},
		{name: "boolString", schema: `module test { data Test { boolValue Bool } }`, request: obj{"boolValue": "true"}},
		{name: "array", schema: `module test { data Test { arrayValue [String] } }`, request: obj{"arrayValue": []any{"test1", "test2"}}},
		{name: "map", schema: `module test { data Test { mapValue {String: String} } }`, request: obj{"mapValue": obj{"key1": "value1", "key2": "value2"}}},
		{name: "dataRef", schema: `module test { data Nested { intValue Int } data Test { dataRef Nested } }`, request: obj{"dataRef": obj{"intValue": 10.0}}},
		{name: "optional", schema: `module test { data Test { intValue Int? } }`, request: obj{}},
		{name: "optionalProvided", schema: `module test { data Test { intValue Int? } }`, request: obj{"intValue": 10.0}},
		{name: "arrayDataRef", schema: `module test { data Nested { intValue Int } data Test { arrayValue [Nested] } }`, request: obj{"arrayValue": []any{obj{"intValue": 10.0}, obj{"intValue": 20.0}}}},
		{name: "mapDataRef", schema: `module test { data Nested { intValue Int } data Test { mapValue {String: Nested} } }`, request: obj{"mapValue": obj{"key1": obj{"intValue": 10.0}, "key2": obj{"intValue": 20.0}}}},
		{name: "otherModuleRef", schema: `module other { data Other { intValue Int } } module test { data Test { otherRef other.Other } }`, request: obj{"otherRef": obj{"intValue": 10.0}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sch, err := schema.ParseString("", test.schema)
			assert.NoError(t, err)

			err = validateRequestMap(&schema.DataRef{Module: "test", Name: "Test"}, nil, test.request, sch)
			assert.NoError(t, err, "%v", test.name)
		})
	}
}

func TestParseQueryParams(t *testing.T) {
	data := &schema.Data{
		Fields: []*schema.Field{
			{Name: "int", Type: &schema.Int{}},
			{Name: "float", Type: &schema.Float{}},
			{Name: "string", Type: &schema.String{}},
			{Name: "bool", Type: &schema.Bool{}},
			{Name: "array", Type: &schema.Array{Element: &schema.Int{}}},
		},
	}
	tests := []struct {
		query   string
		request obj
		err     string
	}{
		{query: "", request: obj{}},
		{query: "int=1", request: obj{"int": "1"}},
		{query: "float=2.2", request: obj{"float": "2.2"}},
		{query: "string=test", request: obj{"string": "test"}},
		{query: "bool=true", request: obj{"bool": "true"}},
		{query: "array=2", request: obj{"array": []string{"2"}}},
		{query: "array=10&array=11", request: obj{"array": []string{"10", "11"}}},
		{query: "int=10&array=11&array=12", request: obj{"int": "10", "array": []string{"11", "12"}}},
		{query: "int=1&int=2", request: nil, err: "multiple values for \"int\" are not supported"},
		{query: "a=bogus", request: nil, err: "unknown query parameter \"a\""},
		{query: "[a,b]=c", request: nil, err: "complex key \"[a,b]\" is not supported, use '@json=' instead"},
		{query: "array=[1,2]", request: nil, err: "complex value \"[1,2]\" is not supported, use '@json=' instead"},
	}

	for _, test := range tests {
		parsedQuery, err := url.ParseQuery(test.query)
		assert.NoError(t, err)
		actual, err := parseQueryParams(parsedQuery, data)
		assert.EqualError(t, err, test.err)
		assert.Equal(t, test.request, actual, test.query)
	}
}

func TestParseQueryJson(t *testing.T) {
	tests := []struct {
		query   string
		request obj
		err     string
	}{
		{query: "@json=", request: nil, err: "failed to parse '@json' query parameter: unexpected end of JSON input"},
		{query: "@json=10", request: nil, err: "failed to parse '@json' query parameter: json: cannot unmarshal number into Go value of type map[string]interface {}"},
		{query: "@json=10&a=b", request: nil, err: "only '@json' parameter is allowed, but other parameters were found"},
		{query: "@json=%7B%7D", request: obj{}},
		{query: `@json=%7B%22a%22%3A%2010%7D`, request: obj{"a": 10.0}},
		{query: `@json=%7B%22a%22%3A%2010%2C%20%22b%22%3A%2011%7D`, request: obj{"a": 10.0, "b": 11.0}},
		{query: `@json=%7B%22a%22%3A%20%7B%22b%22%3A%2010%7D%7D`, request: obj{"a": obj{"b": 10.0}}},
		{query: `@json=%7B%22a%22%3A%20%7B%22b%22%3A%2010%7D%2C%20%22c%22%3A%2011%7D`, request: obj{"a": obj{"b": 10.0}, "c": 11.0}},

		// also works with non-urlencoded json
		{query: `@json={"a": {"b": 10}, "c": 11}`, request: obj{"a": obj{"b": 10.0}, "c": 11.0}},
	}

	for _, test := range tests {
		parsedQuery, err := url.ParseQuery(test.query)
		assert.NoError(t, err)
		actual, err := parseQueryParams(parsedQuery, &schema.Data{})
		assert.EqualError(t, err, test.err)
		assert.Equal(t, test.request, actual, test.query)
	}
}
