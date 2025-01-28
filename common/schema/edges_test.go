package schema

import (
	"slices"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestEdges(t *testing.T) {
	input := Builtins().String() + `
module a {
    config employeeOfTheMonth a.User
    secret myFavoriteChild a.User

    database mysql db

    export data User {
        name String
    }

    export data Event {
        user User
    }

    verb empty(Unit) Unit
    verb postEvent(Event) Unit
    verb getUsers([String]) [a.User]
	    +database calls a.db

    export verb inboundWithExternalTypes(builtin.HttpRequest<Unit, b.Location, Unit>) builtin.HttpResponse<b.Address, String>
        +ingress http GET /todo/destroy/{name}
}
module b {
	export data Location {
		latitude Float
		longitude Float
	}
	export data Address {
		name String
		street String
		city String
		state String
		country String
	}

	topic locations b.Location

	verb consume(b.Location) Unit
	    +subscribe b.locations from=latest

}
module c {
    // cyclic verbs
    verb start(c.AliasedUser) b.Location
        +calls c.middle
    verb middle(b.Address) b.Location
        +calls c.end
    verb end(Unit) a.User
        +calls c.start
    

	enum Color: String {
		Red = "Red"
		Blue = "Blue"
		Green = "Green"
	}

	typealias AliasedUser a.User
}
`

	expected := map[RefKey]GraphNode{}
	// builtins
	addExpectedNode(t, expected, "builtin.Empty",
		[]string{},
		[]string{},
	)
	addExpectedNode(t, expected, "builtin.Ref",
		[]string{"builtin.CatchRequest"},
		[]string{},
	)
	addExpectedNode(t, expected, "builtin.HttpRequest",
		[]string{"a.inboundWithExternalTypes"},
		[]string{},
	)
	addExpectedNode(t, expected, "builtin.HttpResponse",
		[]string{"a.inboundWithExternalTypes"},
		[]string{},
	)
	addExpectedNode(t, expected, "builtin.CatchRequest",
		[]string{},
		[]string{"builtin.Ref"},
	)
	addExpectedNode(t, expected, "builtin.FailedEvent",
		[]string{},
		[]string{},
	)

	// module a
	addExpectedNode(t, expected, "a.employeeOfTheMonth",
		[]string{}, []string{"a.User"})
	addExpectedNode(t, expected, "a.myFavoriteChild",
		[]string{}, []string{"a.User"})
	addExpectedNode(t, expected, "a.db",
		[]string{"a.getUsers"},
		[]string{})
	addExpectedNode(t, expected, "a.User",
		[]string{"a.Event", "a.myFavoriteChild", "a.employeeOfTheMonth", "a.getUsers", "c.AliasedUser", "c.end"},
		[]string{},
	)
	addExpectedNode(t, expected, "a.Event",
		[]string{"a.postEvent"},
		[]string{"a.User"})
	addExpectedNode(t, expected, "a.empty",
		[]string{},
		[]string{})
	addExpectedNode(t, expected, "a.postEvent",
		[]string{}, []string{"a.Event"})
	addExpectedNode(t, expected, "a.getUsers",
		[]string{},
		[]string{"a.User", "a.db"},
	)
	addExpectedNode(t, expected, "a.inboundWithExternalTypes",
		[]string{},
		[]string{"b.Location", "b.Address", "builtin.HttpRequest", "builtin.HttpResponse"},
	)

	// module b
	addExpectedNode(t, expected, "b.Location",
		[]string{"a.inboundWithExternalTypes", "b.consume", "b.locations", "c.middle", "c.start"},
		[]string{},
	)
	addExpectedNode(t, expected, "b.Address",
		[]string{"a.inboundWithExternalTypes", "c.middle"},
		[]string{},
	)
	addExpectedNode(t, expected, "b.locations",
		[]string{"b.consume"},
		[]string{"b.Location"},
	)
	addExpectedNode(t, expected, "b.consume",
		[]string{},
		[]string{"b.Location", "b.locations"},
	)

	// module c
	addExpectedNode(t, expected, "c.start",
		[]string{"c.end"},
		[]string{"c.AliasedUser", "b.Location", "c.middle"},
	)
	addExpectedNode(t, expected, "c.middle",
		[]string{"c.start"},
		[]string{"b.Address", "b.Location", "c.end"},
	)
	addExpectedNode(t, expected, "c.end",
		[]string{"c.middle"},
		[]string{"a.User", "c.start"},
	)
	addExpectedNode(t, expected, "c.Color",
		[]string{},
		[]string{},
	)
	addExpectedNode(t, expected, "c.AliasedUser",
		[]string{"c.start"},
		[]string{"a.User"},
	)

	sch, err := ParseString("", input)
	assert.NoError(t, err)

	graph := Graph(sch)

	refs := map[RefKey]bool{}
	for ref := range graph {
		refs[ref] = true
	}
	for ref := range expected {
		refs[ref] = true
	}
	for ref := range refs {
		t.Run(ref.String(), func(t *testing.T) {
			expectedNode, ok := expected[ref]
			assert.True(t, ok, "did not expect node %s, but got:\nIn: %v\nOut: %v", ref, graph[ref].In, graph[ref].Out)
			graphNode, ok := graph[ref]
			assert.True(t, ok, "expected node %s but graph did not include it", ref)
			assert.Equal(t, expectedNode.In, graphNode.In, "inbound edges for %s should match", ref)
			assert.Equal(t, expectedNode.Out, graphNode.Out, "outbound edges for %s should match", ref)
		})
	}
}

// allows easy building of expected graph nodes
func addExpectedNode(t *testing.T, m map[RefKey]GraphNode, refStr string, in, out []string) {
	t.Helper()

	ref, err := ParseRef(refStr)
	assert.NoError(t, err)
	inRefs := []RefKey{}
	for _, r := range in {
		inRef, err := ParseRef(r)
		assert.NoError(t, err)
		inRefs = append(inRefs, inRef.ToRefKey())
	}
	slices.SortFunc(inRefs, func(i, j RefKey) int {
		return strings.Compare(i.String(), j.String())
	})
	outRefs := []RefKey{}
	for _, r := range out {
		outRef, err := ParseRef(r)
		assert.NoError(t, err)
		outRefs = append(outRefs, outRef.ToRefKey())
	}
	slices.SortFunc(outRefs, func(i, j RefKey) int {
		return strings.Compare(i.String(), j.String())
	})
	m[ref.ToRefKey()] = GraphNode{
		In:  inRefs,
		Out: outRefs,
	}
}
