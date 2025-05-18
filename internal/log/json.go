package log

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	errors "github.com/alecthomas/errors"
)

var _ Sink = (*jsonSink)(nil)

type jsonEntry struct {
	Attributes      map[string]string `json:"attributes,omitempty"`
	Level           string            `json:"level,omitempty"`
	Time            string            `json:"time,omitempty"`
	Error           string            `json:"error,omitempty"`
	Timestamp       time.Time         `json:"timestamp,omitempty"`
	Sequence        int               `json:"sequence,omitempty"`
	LoggerClassName string            `json:"loggerClassName,omitempty"`
	LoggerName      string            `json:"loggerName,omitempty"`
	Message         string            `json:"message,omitempty"`
	ThreadName      string            `json:"threadName,omitempty"`
	ThreadID        int               `json:"threadId,omitempty"`
	Mdc             any               `json:"mdc,omitempty"`
	Ndc             string            `json:"ndc,omitempty"`
	HostName        string            `json:"hostName,omitempty"`
	ProcessName     string            `json:"processName,omitempty"`
	ProcessID       int               `json:"processId,omitempty"`
	Exception       *exceptionRecord  `json:"exception,omitempty"`
	StackTrace      string            `json:"stackTrace,omitempty"`
}

func newJSONSink(w io.Writer) *jsonSink {
	return &jsonSink{
		w:   w,
		enc: json.NewEncoder(w),
	}
}

type jsonSink struct {
	w   io.Writer
	enc *json.Encoder
}

func (j *jsonSink) Log(entry Entry) error {
	var errStr string
	if entry.Error != nil {
		errStr = entry.Error.Error()
	}
	jentry := jsonEntry{
		Level:      entry.Level.String(),
		Time:       entry.Time.Format(time.RFC3339Nano),
		Error:      errStr,
		Attributes: entry.Attributes,
		Message:    entry.Message,
	}
	return errors.WithStack(j.enc.Encode(jentry))
}

// JSONStreamer reads a stream of JSON log entries from r and logs them to log.
//
// If a line of JSON is invalid an entry is created at the defaultLevel.
func JSONStreamer(r io.Reader, log *Logger, defaultLevel Level) error {
	scan := bufio.NewScanner(r)
	scan.Buffer(nil, 1024*1024) // 1MB buffer
	for scan.Scan() {
		var entry jsonEntry
		line := scan.Bytes()
		err := json.Unmarshal(line, &entry)
		if err != nil {
			if len(line) > 0 && line[0] == '{' {
				log.Warnf("Invalid JSON log entry: %s", err)
				log.Warnf("Entry: %s", line)
			}
			log.Log(Entry{Level: defaultLevel, Time: time.Now(), Message: string(line)})
		} else {
			log.Log(entry.ToEntry())
		}
	}
	err := scan.Err()
	if errors.Is(err, io.EOF) || (err != nil && strings.Contains(err.Error(), "already closed")) {
		return nil
	}
	return errors.WithStack(err)
}

type exceptionRecord struct {
	RefID         int              `json:"refId"`
	ExceptionType string           `json:"exceptionType"`
	Message       string           `json:"message"`
	Frames        []frameRecord    `json:"frames"`
	CausedBy      *exceptionRecord `json:"causedBy"`
}

type frameRecord struct {
	Class  string `json:"class"`
	Method string `json:"method"`
	Line   int    `json:"line"`
}

func (r *jsonEntry) ToEntry() Entry {
	var level Level
	var err error
	switch strings.ToLower(r.Level) {
	case "fine", "finer", "config", "debug":
		level = Debug
	case "finest", "trace":
		level = Trace
	case "severe", "error":
		level = Error
	case "warning", "warn":
		level = Warn
	case "", "info":
		level = Info
	default:
		level, err = ParseLevel(r.Level)
		if err != nil {
			level = Warn
		}
	}

	ret := Entry{
		Time:    r.Timestamp,
		Level:   level,
		Message: r.Message,
	}
	if r.StackTrace != "" {
		ret.Message += "\n" + r.StackTrace
	}
	if r.Exception != nil {
		em := fmt.Sprintf("%s %s %d", r.Exception.ExceptionType, r.Exception.Message, len(r.Exception.Frames))
		for _, f := range r.Exception.Frames {
			em += fmt.Sprintf("\n\t%s#%s:%d", f.Class, f.Method, f.Line)
		}
		ret.Error = errors.New(em)
	} else if r.Error != "" {
		ret.Error = errors.New(r.Error)
	}

	// Go is time, JVM is timestamp
	if r.Time != "" {
		ret.Time, err = time.Parse(time.RFC3339Nano, r.Time)
		if err != nil {
			ret.Time = time.Now()
		}
	}
	return ret
}
