package log

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/errors"
	"github.com/benbjohnson/clock"
)

func TestLogger(t *testing.T) {
	w := &strings.Builder{}
	log := New(Trace, newJSONSink(w))
	log.clock = clock.NewMock()
	log.Tracef("trace: %s", "trace")
	log.Debugf("debug: %s", "debug")
	log.Infof("info: %s", "info")
	log.Warnf("warn: %s", "warn")
	log.Errorf(errors.New("error"), "error: %s", "error")
	log = log.Scope("scoped").Attrs(map[string]string{"key": "value"})
	log.Tracef("trace: %s", "trace")
	log.Log(Entry{Level: Trace, Message: "trace: trace"})
	assert.Equal(t, strings.TrimSpace(`
{"level":"trace","time":"1970-01-01T00:00:00Z","timestamp":"0001-01-01T00:00:00Z","message":"trace: trace"}
{"level":"debug","time":"1970-01-01T00:00:00Z","timestamp":"0001-01-01T00:00:00Z","message":"debug: debug"}
{"level":"info","time":"1970-01-01T00:00:00Z","timestamp":"0001-01-01T00:00:00Z","message":"info: info"}
{"level":"warn","time":"1970-01-01T00:00:00Z","timestamp":"0001-01-01T00:00:00Z","message":"warn: warn"}
{"level":"error","time":"1970-01-01T00:00:00Z","error":"error","timestamp":"0001-01-01T00:00:00Z","message":"error: error: error"}
{"attributes":{"key":"value","scope":"scoped"},"level":"trace","time":"1970-01-01T00:00:00Z","timestamp":"0001-01-01T00:00:00Z","message":"trace: trace"}
{"attributes":{"key":"value","scope":"scoped"},"level":"trace","time":"1970-01-01T00:00:00Z","timestamp":"0001-01-01T00:00:00Z","message":"trace: trace"}

`)+"\n", w.String())
}
