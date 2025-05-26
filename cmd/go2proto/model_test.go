package main_test

import (
	"bytes"
	"net/url"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/must"
	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/proto"

	"github.com/block/ftl/cmd/go2proto/testdata"
	"github.com/block/ftl/cmd/go2proto/testdata/external"
	"github.com/block/ftl/cmd/go2proto/testdata/testdatapb"
	"github.com/block/ftl/common/key"
)

func TestModel(t *testing.T) {
	intv := 1
	// UTC, as proto conversion does not preserve timezone
	now := time.Now().UTC()
	model := testdata.Root{
		Int:                1,
		String:             "foo",
		MessagePtr:         &testdata.Message{Time: now},
		Enum:               testdata.EnumA,
		SumType:            &testdata.SumTypeA{A: "bar"},
		OptionalInt:        2,
		OptionalIntPtr:     &intv,
		OptionalMsg:        &testdata.Message{Time: now},
		OptionalWrapper:    optional.Some("foo"),
		RepeatedInt:        []int{1, 2, 3},
		RepeatedMsg:        []*testdata.Message{&testdata.Message{Time: now}, &testdata.Message{Time: now}},
		URL:                must.Get(url.Parse("http://127.0.0.1")),
		Key:                key.NewDeploymentKey("test", "echo"),
		ExternalRoot:       external.Root{Prefix: "abc", Suffix: "xyz"},
		OptionalTime:       optional.Some(now),
		OptionalMessage:    optional.Some[testdata.Message](testdata.Message{}),
		OptionalMessagePtr: optional.Some[*testdata.Message](&testdata.Message{}),
		Map:                map[string]time.Time{now.Format(time.RFC3339): now},
	}
	pb := model.ToProto()
	data, err := proto.Marshal(pb)
	assert.NoError(t, err)
	assert.True(t, bytes.Contains(data, []byte("http://127.0.0.1")), "missing url")
	assert.True(t, bytes.Contains(data, []byte("dpl-test-echo-")), "missing deployment key")
	assert.True(t, bytes.Contains(data, []byte("bar")), "missing sum type value")
	out := &testdatapb.Root{}
	err = proto.Unmarshal(data, out)
	assert.NoError(t, err)
	assert.Equal(t, pb.String(), out.String())

	testModelRoundtrip(t, &model)

	t.Run("test optional.None", func(t *testing.T) {
		testModelRoundtrip(t, &testdata.Root{
			// TODO: ToProto crashes if these 2 are missing
			OptionalWrapper: optional.None[string](),
			URL:             must.Get(url.Parse("http://127.0.0.1")),
			OptionalIntPtr:  &intv,
		})
	})
}

func TestValidate(t *testing.T) {
	msg := &testdata.Message{Invalid: true}
	pb := msg.ToProto()
	_, err := testdata.MessageFromProto(pb)
	assert.EqualError(t, err, "invalid message")
}

func testModelRoundtrip(t *testing.T, model *testdata.Root) {
	t.Helper()

	pb := model.ToProto()
	model2, err := testdata.RootFromProto(pb)
	assert.NoError(t, err)
	assert.Equal(t, model, model2)
}
