package schema

import (
	"fmt"
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
	addExpectedNode(expected, "builtin.Empty",
		[]string{},
		[]string{},
	)
	addExpectedNode(expected, "builtin.Ref",
		[]string{"builtin.CatchRequest"},
		[]string{},
	)
	addExpectedNode(expected, "builtin.HttpRequest",
		[]string{"a.inboundWithExternalTypes"},
		[]string{},
	)
	addExpectedNode(expected, "builtin.HttpResponse",
		[]string{"a.inboundWithExternalTypes"},
		[]string{},
	)
	addExpectedNode(expected, "builtin.CatchRequest",
		[]string{},
		[]string{"builtin.Ref"},
	)
	addExpectedNode(expected, "builtin.FailedEvent",
		[]string{},
		[]string{},
	)

	// module a
	addExpectedNode(expected, "a.employeeOfTheMonth",
		[]string{}, []string{"a.User"})
	addExpectedNode(expected, "a.myFavoriteChild",
		[]string{}, []string{"a.User"})
	addExpectedNode(expected, "a.db",
		[]string{"a.getUsers"},
		[]string{})
	addExpectedNode(expected, "a.User",
		[]string{"a.Event", "a.myFavoriteChild", "a.employeeOfTheMonth", "a.getUsers", "c.AliasedUser", "c.end"},
		[]string{},
	)
	addExpectedNode(expected, "a.Event",
		[]string{"a.postEvent"},
		[]string{"a.User"})
	addExpectedNode(expected, "a.empty",
		[]string{},
		[]string{})
	addExpectedNode(expected, "a.postEvent",
		[]string{}, []string{"a.Event"})
	addExpectedNode(expected, "a.getUsers",
		[]string{},
		[]string{"a.User", "a.db"},
	)
	addExpectedNode(expected, "a.inboundWithExternalTypes",
		[]string{},
		[]string{"b.Location", "b.Address", "builtin.HttpRequest", "builtin.HttpResponse"},
	)

	// module b
	addExpectedNode(expected, "b.Location",
		[]string{"a.inboundWithExternalTypes", "b.consume", "b.locations", "c.middle", "c.start"},
		[]string{},
	)
	addExpectedNode(expected, "b.Address",
		[]string{"a.inboundWithExternalTypes", "c.middle"},
		[]string{},
	)
	addExpectedNode(expected, "b.locations",
		[]string{"b.consume"},
		[]string{"b.Location"},
	)
	addExpectedNode(expected, "b.consume",
		[]string{},
		[]string{"b.Location", "b.locations"},
	)

	// module c
	addExpectedNode(expected, "c.start",
		[]string{"c.end"},
		[]string{"c.AliasedUser", "b.Location", "c.middle"},
	)
	addExpectedNode(expected, "c.middle",
		[]string{"c.start"},
		[]string{"b.Address", "b.Location", "c.end"},
	)
	addExpectedNode(expected, "c.end",
		[]string{"c.middle"},
		[]string{"a.User", "c.start"},
	)
	addExpectedNode(expected, "c.Color",
		[]string{},
		[]string{},
	)
	addExpectedNode(expected, "c.AliasedUser",
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
func addExpectedNode(m map[RefKey]GraphNode, refStr string, in, out []string) error {
	ref, err := ParseRef(refStr)
	if err != nil {
		return fmt.Errorf("could not parse ref %q: %w", refStr, err)
	}
	inRefs := []RefKey{}
	for _, r := range in {
		inRef, err := ParseRef(r)
		if err != nil {
			return fmt.Errorf("could not parse ref %q: %w", r, err)
		}
		inRefs = append(inRefs, inRef.ToRefKey())
	}
	slices.SortFunc(inRefs, func(i, j RefKey) int {
		return strings.Compare(i.String(), j.String())
	})
	outRefs := []RefKey{}
	for _, r := range out {
		outRef, err := ParseRef(r)
		if err != nil {
			return fmt.Errorf("could not parse ref %q: %w", r, err)
		}
		outRefs = append(outRefs, outRef.ToRefKey())
	}
	slices.SortFunc(outRefs, func(i, j RefKey) int {
		return strings.Compare(i.String(), j.String())
	})
	m[ref.ToRefKey()] = GraphNode{
		In:  inRefs,
		Out: outRefs,
	}
	return nil
}
