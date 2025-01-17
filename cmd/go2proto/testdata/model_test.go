package testdata

import (
	"bytes"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/must"
	"google.golang.org/protobuf/proto"

	"github.com/block/ftl/cmd/go2proto/testdata/testdatapb"
	"github.com/block/ftl/internal/key"
)

func TestModel(t *testing.T) {
	intv := 1
	// UTC, as proto conversion does not preserve timezone
	now := time.Now().UTC()
	model := Root{
		Int:            1,
		String:         "foo",
		MessagePtr:     &Message{Time: now},
		Enum:           EnumA,
		SumType:        &SumTypeA{A: "bar"},
		OptionalInt:    2,
		OptionalIntPtr: &intv,
		OptionalMsg:    &Message{Time: now},
		RepeatedInt:    []int{1, 2, 3},
		RepeatedMsg:    []*Message{&Message{Time: now}, &Message{Time: now}},
		URL:            must.Get(url.Parse("http://127.0.0.1")),
		Key:            key.NewDeploymentKey("echo"),
	}
	pb := model.ToProto()
	fmt.Println(pb)
	data, err := proto.Marshal(pb)
	assert.NoError(t, err)
	assert.True(t, bytes.Contains(data, []byte("http://127.0.0.1")), "missing url")
	assert.True(t, bytes.Contains(data, []byte("dpl-echo-")), "missing deployment key")
	assert.True(t, bytes.Contains(data, []byte("bar")), "missing sum type value")
	out := &testdatapb.Root{}
	err = proto.Unmarshal(data, out)
	assert.NoError(t, err)
	assert.Equal(t, pb.String(), out.String())

	testModelRoundtrip(t, &model)
}

func testModelRoundtrip(t *testing.T, model *Root) {
	t.Helper()

	pb := model.ToProto()
	model2, err := RootFromProto(pb)
	assert.NoError(t, err)
	assert.Equal(t, model, model2)
}
