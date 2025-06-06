package ftltest

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/go-runtime/ftl"
	"github.com/block/ftl/go-runtime/internal"
	configuration "github.com/block/ftl/internal/config"
)

// pubSubEvent is a sum type for all events that can be published to the pubsub system.
// not to be confused with an event that gets published to a topic
//
//sumtype:decl
type pubSubEvent interface {
	// cronJobEvent is a marker to ensure that all events implement the interface.
	pubSubEvent()
}

// publishEvent holds an event to be published to a topic
type publishEvent struct {
	topic   *schema.Ref
	content any
}

func (publishEvent) pubSubEvent() {}

// subscriptionDidConsumeEvent indicates that a call to a subscriber has completed
type subscriptionDidConsumeEvent struct {
	subscription string
	err          error
}

func (subscriptionDidConsumeEvent) pubSubEvent() {}

type subscription struct {
	name        string
	topic       *schema.Ref
	cursor      optional.Option[int]
	isExecuting bool
	errors      map[int]error
}

type subscriber func(context.Context, any) error

type fakeFTL struct {

	// We store the options used to construct this fake, so they can be
	// replayed to extend the fake with new options
	options []Option

	mockMaps      map[uintptr]mapImpl
	allowMapCalls bool
	configValues  map[string][]byte
	secretValues  map[string][]byte
	egressValues  map[string]string
	pubSub        *fakePubSub
}

func (f *fakeFTL) GetEgress(ctx context.Context, name string) (string, error) {
	return f.egressValues[name], nil
}

// mapImpl is a function that takes an object and returns an object of a potentially different
// type but is not constrained by input/output type like ftl.Map.
type mapImpl func(context.Context) (any, error)

func contextWithFakeFTL(ctx context.Context, options ...Option) context.Context {
	fake := &fakeFTL{
		mockMaps:      map[uintptr]mapImpl{},
		allowMapCalls: false,
		configValues:  map[string][]byte{},
		secretValues:  map[string][]byte{},
		egressValues:  map[string]string{},
		options:       options,
	}
	ctx = internal.WithContext(ctx, fake)

	// TODO: revisit this and all unused funcs and types
	// fake.pubSub = newFakePubSub(ctx)
	return ctx
}

func (f *fakeFTL) setConfig(name string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return errors.WithStack(err)
	}
	f.configValues[name] = data
	return nil
}

func (f *fakeFTL) GetConfig(ctx context.Context, name string, dest any) error {
	data, ok := f.configValues[name]
	if !ok {
		return errors.Wrapf(configuration.ErrNotFound, "config value %q not found, did you remember to ctx := ftltest.Context(ftltest.WithDefaultProjectFile()) ?", name)
	}
	return errors.WithStack(json.Unmarshal(data, dest))
}

func (f *fakeFTL) setSecret(name string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return errors.WithStack(err)
	}
	f.secretValues[name] = data
	return nil
}

func (f *fakeFTL) GetSecret(ctx context.Context, name string, dest any) error {
	data, ok := f.secretValues[name]
	if !ok {
		return errors.Wrapf(configuration.ErrNotFound, "config value %q not found, did you remember to ctx := ftltest.Context(ftltest.WithDefaultProjectFile()) ?", name)
	}
	return errors.WithStack(json.Unmarshal(data, dest))
}

// addMapMock saves a new mock of ftl.Map to the internal map in fakeFTL.
//
// mockMap provides the whole mock implemention, so it gets called in place of both `fn`
// and `getter` in ftl.Map.
func addMapMock[T, U any](f *fakeFTL, mapper *ftl.MapHandle[T, U], mockMap func(context.Context) (U, error)) {
	key := makeMapKey(mapper)
	f.mockMaps[key] = func(ctx context.Context) (any, error) {
		return errors.WithStack2(mockMap(ctx))
	}
}

func (f *fakeFTL) startAllowingMapCalls() {
	f.allowMapCalls = true
}

func (f *fakeFTL) CallMap(ctx context.Context, mapper any, value any, mapImpl func(context.Context) (any, error)) any {
	key := makeMapKey(mapper)
	mockMap, ok := f.mockMaps[key]
	if ok {
		return actuallyCallMap(ctx, mockMap)
	}
	if f.allowMapCalls {
		return actuallyCallMap(ctx, mapImpl)
	}
	panic("map calls not allowed in tests by default, ftltest.Context should be instantiated with either ftltest.WithMapsAllowed() or a mock for the specific map being called using ftltest.WhenMap(...)")
}

func makeMapKey(mapper any) uintptr {
	v := reflect.ValueOf(mapper)
	if v.Kind() != reflect.Pointer {
		panic("fakeFTL received object that was not a pointer, expected *MapHandle")
	}
	underlying := v.Elem().Type().Name()
	if !strings.HasPrefix(underlying, "MapHandle[") {
		panic(fmt.Sprintf("fakeFTL received *%s, expected *MapHandle", underlying))
	}
	return v.Pointer()
}

func actuallyCallMap(ctx context.Context, impl mapImpl) any {
	out, err := impl(ctx)
	if err != nil {
		panic(err)
	}
	return out
}

func (f *fakeFTL) PublishEvent(ctx context.Context, topic *schema.Ref, event any, key string) error {
	return errors.WithStack(f.pubSub.publishEvent(topic, event))
}
