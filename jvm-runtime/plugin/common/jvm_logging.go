package common

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/alecthomas/atomic"

	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/log"
)

var _ io.Writer = &errorDetector{}

// errorDetector is a writer that forwards output to stdout, while capturing output until a safe state is reached.
// When safe state is reached, the output is then parsed for errors.
type errorDetector struct {
	logger     *log.Logger
	output     string
	ended      atomic.Value[bool]
	errors     []builderrors.Error
	jsonBuffer string
}

func (o *errorDetector) Write(p []byte) (n int, err error) {
	// We track the last log level for things like stack traces that don't have a prefix
	var last = func(format string, args ...interface{}) {
		o.logger.Debugf(format, args...)
	}
	val := string(p)
	if o.jsonBuffer != "" {
		val = o.jsonBuffer + val
		o.jsonBuffer = ""
	}
	for _, line := range strings.Split(val, "\n") {
		skip := false
		if len(line) == 0 {
			continue
		}

		if line[0] == '{' {
			record := JvmLogRecord{}
			err := json.Unmarshal([]byte(line), &record)
			if err == nil {
				// We do not want to dump any raw json output
				skip = true
				entry := record.ToEntry()
				// Huge hack, Quarkus can currently NPE when being pinged at startup, due to a race generating the 404 page before it has started
				// this should be fixed in a later version of Quarkus
				if !strings.Contains(entry.Message, "Request to /xyz.block.ftl.v1.VerbService/Ping failed") {
					o.logger.Log(entry)
					last = func(format string, args ...interface{}) {
						o.logger.Logf(entry.Level, format, args...)
					}
				}
			} else if strings.Contains(err.Error(), "unexpected end of JSON input") {
				o.jsonBuffer = line
			} else {
				o.logger.Errorf(err, "Log Parse Failure: %s", line) //nolint
			}
		} else if cleanLine, ok := strings.CutPrefix(line, "[ERROR] "); ok {
			o.logger.Logf(log.Error, "%s", cleanLine)
			last = func(format string, args ...interface{}) {
				o.logger.Logf(log.Error, format, args...)
			}
		} else if _, cleanLine, ok := strings.Cut(line, "ERROR"); ok {
			o.logger.Logf(log.Error, "%s", cleanLine)
			last = func(format string, args ...interface{}) {
				o.logger.Logf(log.Error, format, args...)
			}
		} else if cleanLine, ok := strings.CutPrefix(line, "[WARNING] "); ok {
			o.logger.Warnf("%s", cleanLine)
			last = o.logger.Warnf
		} else if _, cleanLine, ok := strings.Cut(line, "WARN"); ok {
			o.logger.Warnf("%s", cleanLine)
			last = o.logger.Warnf
		} else if cleanLine, ok := strings.CutPrefix(line, "[INFO] "); ok {
			// We downgrade maven errors, as maven is very verbose
			// Basically we just log maven Info as our Debug
			o.logger.Debugf("%s", cleanLine)
			last = o.logger.Debugf
		} else if _, cleanLine, ok := strings.Cut(line, "INFO"); ok {
			o.logger.Debugf("%s", cleanLine)
			last = o.logger.Debugf
		} else if cleanLine, ok := strings.CutPrefix(line, "[DEBUG] "); ok {
			o.logger.Debugf("%s", cleanLine)
			last = o.logger.Debugf
		} else if cleanLine, _, ok := strings.Cut(line, "DEBUG"); ok {
			o.logger.Debugf("%s", cleanLine)
			last = o.logger.Debugf
		} else if cleanLine, ok := strings.CutPrefix(line, "[TRACE] "); ok {
			o.logger.Tracef("%s", cleanLine)
			last = o.logger.Tracef
		} else if _, cleanLine, ok := strings.Cut(line, "TRACE"); ok {
			o.logger.Tracef("%s", cleanLine)
			last = o.logger.Tracef
		} else if _, cleanLine, ok := strings.Cut(line, "FINE"); ok {
			o.logger.Tracef("%s", cleanLine)
			last = o.logger.Tracef
		} else {
			last("%s", line)
		}
		if !o.ended.Load() && !skip {
			o.output += line + "\n"
		}
	}
	return len(p), nil
}

func (o *errorDetector) FinalizeCapture(dump bool) []builderrors.Error {
	if o.ended.Load() {
		return o.errors
	}
	o.ended.Store(true)

	// Example error output:
	// [ERROR] /example/path/StockPricePublisher.kt: (31, 24) 'public' function exposes its 'internal' parameter type 'StockPricesTopic'.
	// [ERROR] Failed to execute goal io.quarkus:quarkus-maven-plugin:3.18.1:dev (default-cli) on project stockprices: Unable to execute mojo: Compilation failure
	// [ERROR] /example/path/StockPricePublisher.kt:[31,24] 'public' function exposes its 'internal' parameter type 'StockPricesTopic'.
	// [ERROR] -> [Help 1]
	// [ERROR]
	// [ERROR] To see the full stack trace of the errors, re-run Maven with the -e switch.
	// [ERROR] Re-run Maven using the -X switch to enable full debug logging.
	// [ERROR]
	// [ERROR] For more information about the errors and possible solutions, please read the following articles:
	// [ERROR] [Help 1] http://cwiki.apache.org/confluence/display/MAVEN/MojoExecutionException

	lines := slices.Filter(strings.Split(o.output, "\n"), func(s string) bool {
		return strings.HasPrefix(s, "[ERROR]")
	})
	for _, line := range lines {
		line = strings.TrimSpace(strings.TrimPrefix(line, "[ERROR]"))
		if len(line) == 0 {
			continue
		}
		if strings.HasPrefix(line, "To see the full") || strings.HasPrefix(line, "->") {
			// This looks like the end of the errors and the start of help text
			break
		}
		o.errors = append(o.errors, builderrors.Error{
			Msg:   line,
			Type:  builderrors.COMPILER,
			Level: builderrors.ERROR,
		})
	}
	if dump && len(o.errors) == 0 {
		// If we got no useful error output do a dump
		for _, l := range strings.Split(o.output, "\n") {
			o.logger.Warnf("%s", l)
		}
	}
	return o.errors
}

type JvmLogRecord struct {
	Timestamp       time.Time        `json:"timestamp"`
	Sequence        int              `json:"sequence"`
	LoggerClassName string           `json:"loggerClassName"`
	LoggerName      string           `json:"loggerName"`
	Level           string           `json:"level"`
	Message         string           `json:"message"`
	ThreadName      string           `json:"threadName"`
	ThreadID        int              `json:"threadId"`
	Mdc             any              `json:"mdc"`
	Ndc             string           `json:"ndc"`
	HostName        string           `json:"hostName"`
	ProcessName     string           `json:"processName"`
	ProcessID       int              `json:"processId"`
	Exception       *ExceptionRecord `json:"exception"`
	StackTrace      string           `json:"stackTrace,omitempty"`
}

type ExceptionRecord struct {
	RefID         int              `json:"refId"`
	ExceptionType string           `json:"exceptionType"`
	Message       string           `json:"message"`
	Frames        []FrameRecord    `json:"frames"`
	CausedBy      *ExceptionRecord `json:"causedBy"`
}

type FrameRecord struct {
	Class  string `json:"class"`
	Method string `json:"method"`
	Line   int    `json:"line"`
}

func (r *JvmLogRecord) ToEntry() log.Entry {
	level := log.Info
	switch r.Level {
	case "DEBUG":
		level = log.Debug
	case "INFO":
		level = log.Info
	case "WARN", "WARNING":
		level = log.Warn
	case "ERROR", "SEVERE":
		level = log.Error
	case "FINE", "FINER":
		level = log.Debug
	case "FINEST", "TRACE":
		level = log.Trace
	default:
		r.Message = r.Level + ": " + r.Message
	}
	if r.Exception != nil {
		r.Message += r.Exception.ExceptionType + " " + r.Exception.Message
		for _, f := range r.Exception.Frames {
			r.Message += fmt.Sprintf("\n\t%s#%s:%d", f.Class, f.Method, f.Line)
		}
	}
	if r.StackTrace != "" {
		r.Message += "\n" + r.StackTrace
	}

	ret := log.Entry{
		Time:    r.Timestamp,
		Level:   level,
		Message: r.Message,
	}
	return ret
}
