package eventstream

import (
	"context"
	"iter"
	"sync"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/pubsub"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/iterops"
)

// EventView is a materialized view of an event stream.
type EventView[View any, E Event[View]] interface {
	View(ctx context.Context) (View, error)

	Publish(ctx context.Context, event E) error
	Changes(ctx context.Context) (iter.Seq[View], error)
}

// EventStream is a stream of events that can be published and subscribed to, that update a materialized view
type EventStream[View any, E Event[View]] interface {
	EventView[View, E]

	Updates() *pubsub.Topic[E]
}

// StreamView is a view of an event stream that can be subscribed to, without modifying the stream.
type StreamView[View any] interface {
	View(ctx context.Context) (View, error)

	// Subscribe to the event stream. The channel will only receive events that are published after the subscription.
	Subscribe(ctx context.Context) <-chan Event[View]
}

type Event[View any] interface {

	// Handle applies the event to the view
	Handle(view View) (View, error)
}

func NewInMemory[View any, E Event[View]](initial View) EventStream[View, E] {
	return &inMemoryEventStream[View, E]{
		view:  initial,
		topic: pubsub.New[E](),
	}
}

type inMemoryEventStream[View any, E Event[View]] struct {
	view  View
	lock  sync.Mutex
	topic *pubsub.Topic[E]
}

func (i *inMemoryEventStream[T, E]) Publish(ctx context.Context, e E) error {
	i.lock.Lock()
	defer i.lock.Unlock()
	logger := log.FromContext(ctx)

	if _, ok := any(e).(VerboseMessage); ok {
		logger.Tracef("Publishing event %T%v", e, e)
	} else {
		logger.Debugf("Publishing event %T%v", e, e)
	}
	newView, err := e.Handle(reflect.DeepCopy(i.view))
	if err != nil {
		return errors.Wrap(err, "failed to handle event")
	}
	i.view = newView
	i.topic.Publish(e)
	return nil
}

func (i *inMemoryEventStream[T, E]) Changes(ctx context.Context) (iter.Seq[T], error) {
	updates := i.Updates().Subscribe(nil)
	context.AfterFunc(ctx, func() {
		i.Updates().Unsubscribe(updates)
	})
	iter := channels.IterContext(ctx, updates)

	return iterops.Map(iter, func(e E) T {
		return i.view
	}), nil
}

func (i *inMemoryEventStream[T, E]) View(ctx context.Context) (T, error) {
	i.lock.Lock()
	defer i.lock.Unlock()
	out := reflect.DeepCopy(i.view)
	return out, nil
}

func (i *inMemoryEventStream[T, E]) Updates() *pubsub.Topic[E] {
	return i.topic
}

type VerboseMessage interface {
	VerboseMessage()
}
