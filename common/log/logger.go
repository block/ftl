package log

import (
	"bufio"
	"fmt"
	"io"
	"maps"
	"runtime"
	"time"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"
	"github.com/benbjohnson/clock"

	"github.com/block/ftl/common/key"
)

var _ Interface = (*Logger)(nil)

const scopeKey = "scope"
const moduleKey = "module"
const deploymentKey = "deployment"
const changesetKey = "changeset"

type Entry struct {
	Time       time.Time         `json:"-"`
	Level      Level             `json:"level"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Message    string            `json:"message"`

	Error error `json:"-"`
}

// Logger is the concrete logger.
type Logger struct {
	debugLogger optional.Option[devModeDebugLogger]
	level       *atomic.Value[Level]
	attributes  map[string]string
	sink        Sink
	clock       clock.Clock
}

// New returns a new logger.
func New(level Level, sink Sink) *Logger {
	return &Logger{
		level:      atomic.New(level),
		attributes: map[string]string{},
		sink:       sink,
		clock:      clock.New(),
	}
}

func (l Logger) Scope(scope string) *Logger {
	return l.Attrs(map[string]string{scopeKey: scope})
}

func (l Logger) Module(module string) *Logger {
	return l.Attrs(map[string]string{moduleKey: module})
}

func (l Logger) Deployment(deployment key.Deployment) *Logger {
	return l.Attrs(map[string]string{deploymentKey: deployment.String()})
}

func (l Logger) Changeset(changeset key.Changeset) *Logger {
	return l.Attrs(map[string]string{changesetKey: changeset.String()})
}

// Attrs creates a new logger with the given attributes.
func (l Logger) Attrs(attributes map[string]string) *Logger {
	attr := make(map[string]string, len(l.attributes)+len(attributes))
	maps.Copy(attr, l.attributes)
	maps.Copy(attr, attributes)
	l.attributes = attr
	return &l
}

func (l Logger) AddSink(sink Sink) *Logger {
	l.sink = Tee(l.sink, sink)
	return &l
}

func (l Logger) Level(level Level) *Logger {
	l.level = atomic.New(level)
	return &l
}

func (l Logger) GetLevel() Level {
	return l.level.Load()
}

func (l *Logger) Log(entry Entry) {
	var mergedAttributes map[string]string
	if t, ok := l.debugLogger.Get(); ok {
		mergedAttributes = make(map[string]string, len(l.attributes)+len(entry.Attributes))
		maps.Copy(mergedAttributes, l.attributes)
		maps.Copy(mergedAttributes, entry.Attributes)
		entry.Attributes = mergedAttributes
		t.delegate.Log(entry)
	}
	if entry.Level < l.level.Load() {
		return
	}
	if entry.Time.IsZero() {
		// Use UTC now
		entry.Time = l.clock.Now().UTC()
	}

	if mergedAttributes == nil {
		mergedAttributes = make(map[string]string, len(l.attributes)+len(entry.Attributes))
		maps.Copy(mergedAttributes, l.attributes)
		maps.Copy(mergedAttributes, entry.Attributes)
		entry.Attributes = mergedAttributes
	}
	_ = l.sink.Log(entry) // nolint:errcheck
}

func (l *Logger) Logf(level Level, format string, args ...interface{}) {
	l.Log(Entry{Level: level, Message: fmt.Sprintf(format, args...)})
}

func (l *Logger) Tracef(format string, args ...interface{}) {
	l.Log(Entry{Level: Trace, Message: fmt.Sprintf(format, args...)})
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Log(Entry{Level: Debug, Message: fmt.Sprintf(format, args...)})
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Log(Entry{Level: Info, Message: fmt.Sprintf(format, args...)})
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Log(Entry{Level: Warn, Message: fmt.Sprintf(format, args...)})
}

func (l *Logger) Errorf(err error, format string, args ...interface{}) {
	if err == nil {
		return
	}
	l.Log(Entry{Level: Error, Message: fmt.Sprintf(format, args...) + ": " + err.Error(), Error: err})
}

// WriterAt returns a writer that logs each line at the given level.
func (l *Logger) WriterAt(level Level) *io.PipeWriter {
	// Based on MIT licensed Logrus https://github.com/sirupsen/logrus/blob/bdc0db8ead3853c56b7cd1ac2ba4e11b47d7da6b/writer.go#L27
	reader, writer := io.Pipe()
	var printFunc func(format string, args ...interface{})

	switch level {
	case Trace:
		printFunc = l.Tracef
	case Debug:
		printFunc = l.Debugf
	case Info:
		printFunc = l.Infof
	case Warn:
		printFunc = l.Warnf
	case Error:
		printFunc = func(format string, args ...interface{}) {
			l.Errorf(nil, format, args...)
		}
	default:
		panic(level)
	}

	go l.writerScanner(reader, printFunc)
	runtime.SetFinalizer(writer, writerFinalizer)

	return writer
}

func (l *Logger) writerScanner(reader *io.PipeReader, printFunc func(format string, args ...interface{})) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(nil, 1024*1024) // 1MB buffer
	for scanner.Scan() {
		printFunc("%s", scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		l.Errorf(err, "Error while reading from Writer")
	}
	reader.Close()
}

func writerFinalizer(writer *io.PipeWriter) {
	writer.Close()
}

// SetCurrentLevel sets the current log level, it is a global operation, and will affect all contexts using this logger
func SetCurrentLevel(l *Logger, level Level) {
	l.level.Store(level)
}

// NewNoopSink returns a sink that does nothing.
func NewNoopSink() Sink {
	return noopSink{}
}

type noopSink struct{}

func (noopSink) Log(entry Entry) error { return nil }
