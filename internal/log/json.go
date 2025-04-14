package log

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"time"

	errors "github.com/alecthomas/errors"
)

var _ Sink = (*jsonSink)(nil)

type jsonEntry struct {
	Entry
	Level string `json:"level,omitempty"`
	Time  string `json:"time,omitempty"`
	Error string `json:"error,omitempty"`
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
		Level: entry.Level.String(),
		Time:  entry.Time.Format(time.RFC3339Nano),
		Error: errStr,
		Entry: entry,
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
			if entry.Error != "" {
				entry.Entry.Error = errors.New(entry.Error)
			}
			switch strings.ToLower(entry.Level) {
			case "fine", "finer", "config":
				entry.Entry.Level = Debug
			case "finest":
				entry.Entry.Level = Trace
			case "severe":
				entry.Entry.Level = Error
			case "":
				entry.Entry.Level = Info
			default:
				entry.Entry.Level, err = ParseLevel(entry.Level)
				if err != nil {
					log.Warnf("Invalid log level: %s", err)
					entry.Entry.Level = defaultLevel
				}
			}
			entry.Entry.Time, err = time.Parse(time.RFC3339Nano, entry.Time)
			if err != nil {
				entry.Entry.Time = time.Now()
			}
			log.Log(entry.Entry)
		}
	}
	err := scan.Err()
	if errors.Is(err, io.EOF) || (err != nil && strings.Contains(err.Error(), "already closed")) {
		return nil
	}
	return errors.WithStack(err)
}
