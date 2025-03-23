package log

import (
	"errors"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
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
{"message":"trace: trace","level":"trace","time":"1970-01-01T00:00:00Z"}
{"message":"debug: debug","level":"debug","time":"1970-01-01T00:00:00Z"}
{"message":"info: info","level":"info","time":"1970-01-01T00:00:00Z"}
{"message":"warn: warn","level":"warn","time":"1970-01-01T00:00:00Z"}
{"message":"error: error: error","level":"error","time":"1970-01-01T00:00:00Z","error":"error"}
{"attributes":{"key":"value","scope":"scoped"},"message":"trace: trace","level":"trace","time":"1970-01-01T00:00:00Z"}
{"attributes":{"key":"value","scope":"scoped"},"message":"trace: trace","level":"trace","time":"1970-01-01T00:00:00Z"}
`)+"\n", w.String())
}
