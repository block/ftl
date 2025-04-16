package timeline

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"google.golang.org/protobuf/proto"

	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
)

func TestFilterModule_Multiple(t *testing.T) {
	// Create test events
	event1 := &timelinepb.Event{
		Entry: &timelinepb.Event_Call{
			Call: &timelinepb.CallEvent{
				DestinationVerbRef: &schemapb.Ref{Module: "module1", Name: "verb1"},
			},
		},
	}
	event2 := &timelinepb.Event{
		Entry: &timelinepb.Event_Log{
			Log: &timelinepb.LogEvent{
				Attributes: map[string]string{"module": "module2", "verb": "verb2"},
			},
		},
	}
	event3 := &timelinepb.Event{
		Entry: &timelinepb.Event_Call{
			Call: &timelinepb.CallEvent{
				DestinationVerbRef: &schemapb.Ref{Module: "module3", Name: "verb3"},
			},
		},
	}

	// Create multiple module filters
	filters := []*timelinepb.TimelineQuery_ModuleFilter{
		{Module: "module1", Verb: proto.String("verb1")},
		{Module: "module2", Verb: proto.String("verb2")},
	}

	filter := FilterModule(filters)

	// Test that events matching either filter pass through
	assert.True(t, filter(event1), "event1 should match first filter")
	assert.True(t, filter(event2), "event2 should match second filter")
	assert.False(t, filter(event3), "event3 should not match any filter")

	// Test with one filter having no verb specified
	filters = []*timelinepb.TimelineQuery_ModuleFilter{
		{
			Module: "module1",
		},
		{
			Module: "module2",
			Verb:   proto.String("verb2"),
		},
	}

	filter = FilterModule(filters)

	// Test that module1 matches regardless of verb
	assert.True(t, filter(event1), "event1 should match first filter (any verb)")
	assert.True(t, filter(event2), "event2 should match second filter")
	assert.False(t, filter(event3), "event3 should not match any filter")
}
