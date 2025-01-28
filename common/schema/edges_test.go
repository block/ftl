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
        +ingress http GET /todo/external/{name}

	export verb inboundWithDupes(builtin.HttpRequest<Unit, b.Location, Unit>) builtin.HttpResponse<b.Location, String>
        +ingress http GET /todo/dupes/{name}
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
		nil,
		nil,
	)
	addExpectedNode(t, expected, "builtin.Ref",
		[]string{"builtin.CatchRequest"},
		nil,
	)
	addExpectedNode(t, expected, "builtin.HttpRequest",
		[]string{"a.inboundWithExternalTypes", "a.inboundWithDupes"},
		nil,
	)
	addExpectedNode(t, expected, "builtin.HttpResponse",
		[]string{"a.inboundWithExternalTypes", "a.inboundWithDupes"},
		nil,
	)
	addExpectedNode(t, expected, "builtin.CatchRequest",
		nil,
		[]string{"builtin.Ref"},
	)
	addExpectedNode(t, expected, "builtin.FailedEvent",
		nil,
		nil,
	)

	// module a
	addExpectedNode(t, expected, "a.employeeOfTheMonth",
		nil,
		[]string{"a.User"})
	addExpectedNode(t, expected, "a.myFavoriteChild",
		nil,
		[]string{"a.User"})
	addExpectedNode(t, expected, "a.db",
		[]string{"a.getUsers"},
		nil)
	addExpectedNode(t, expected, "a.User",
		[]string{"a.Event", "a.myFavoriteChild", "a.employeeOfTheMonth", "a.getUsers", "c.AliasedUser", "c.end"},
		nil,
	)
	addExpectedNode(t, expected, "a.Event",
		[]string{"a.postEvent"},
		[]string{"a.User"})
	addExpectedNode(t, expected, "a.empty",
		nil,
		nil)
	addExpectedNode(t, expected, "a.postEvent",
		nil, []string{"a.Event"})
	addExpectedNode(t, expected, "a.getUsers",
		nil,
		[]string{"a.User", "a.db"},
	)
	addExpectedNode(t, expected, "a.inboundWithExternalTypes",
		nil,
		[]string{"b.Location", "b.Address", "builtin.HttpRequest", "builtin.HttpResponse"},
	)
	addExpectedNode(t, expected, "a.inboundWithDupes",
		nil,
		[]string{"b.Location", "builtin.HttpRequest", "builtin.HttpResponse"},
	)

	// module b
	addExpectedNode(t, expected, "b.Location",
		[]string{"a.inboundWithExternalTypes", "a.inboundWithDupes", "b.consume", "b.locations", "c.middle", "c.start"},
		nil,
	)
	addExpectedNode(t, expected, "b.Address",
		[]string{"a.inboundWithExternalTypes", "c.middle"},
		nil,
	)
	addExpectedNode(t, expected, "b.locations",
		[]string{"b.consume"},
		[]string{"b.Location"},
	)
	addExpectedNode(t, expected, "b.consume",
		nil,
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
		nil,
		nil,
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
			assertEqualOrBothEmpty(t, expectedNode.In, graphNode.In, "inbound edges for %s should match", ref)
			assertEqualOrBothEmpty(t, expectedNode.Out, graphNode.Out, "outbound edges for %s should match", ref)
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

func assertEqualOrBothEmpty(t *testing.T, a, b []RefKey, msgAndArgs ...any) {
	t.Helper()
	if len(a) == 0 && len(b) == 0 {
		return
	}
	assert.Equal(t, a, b, msgAndArgs...)
}
