package log

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
)

var _ Sink = (*jsonSink)(nil)

type jsonEntry struct {
	Entry
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
		Time:  entry.Time.Format(time.RFC3339Nano),
		Error: errStr,
		Entry: entry,
	}
	return j.enc.Encode(jentry)
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
	return err
}
