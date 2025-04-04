package common

import (
	"encoding/json"
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
	logger *log.Logger
	output string
	ended  atomic.Value[bool]
	errors []builderrors.Error
}

func (o *errorDetector) Write(p []byte) (n int, err error) {
	// Forward output to stdout
	for _, line := range strings.Split(string(p), "\n") {
		if len(line) == 0 {
			continue
		}
		if cleanLine, ok := strings.CutPrefix(line, "[ERROR] "); ok {
			o.logger.Logf(log.Error, "%s", cleanLine)
		} else if cleanLine, ok := strings.CutPrefix(line, "ERROR "); ok {
			o.logger.Logf(log.Error, "%s", cleanLine)
		} else if cleanLine, ok := strings.CutPrefix(line, "[WARNING] "); ok {
			o.logger.Warnf("%s", cleanLine)
		} else if cleanLine, ok := strings.CutPrefix(line, "WARN "); ok {
			o.logger.Warnf("%s", cleanLine)
		} else if cleanLine, ok := strings.CutPrefix(line, "[INFO] "); ok {
			// We downgrade maven errors, as maven is very verbose
			// Basically we just log maven Info as our Debug
			o.logger.Debugf("%s", cleanLine)
		} else if cleanLine, ok := strings.CutPrefix(line, "INFO "); ok {
			o.logger.Debugf("%s", cleanLine)
		} else if cleanLine, ok := strings.CutPrefix(line, "[DEBUG] "); ok {
			o.logger.Debugf("%s", cleanLine)
		} else if cleanLine, ok := strings.CutPrefix(line, "DEBUG "); ok {
			o.logger.Debugf("%s", cleanLine)
		} else if cleanLine, ok := strings.CutPrefix(line, "[TRACE] "); ok {
			o.logger.Tracef("%s", cleanLine)
		} else if cleanLine, ok := strings.CutPrefix(line, "TRACE "); ok {
			o.logger.Tracef("%s", cleanLine)
		} else if cleanLine, ok := strings.CutPrefix(line, "FINE "); ok {
			o.logger.Tracef("%s", cleanLine)
		} else if line[0] == '{' {
			record := JvmLogRecord{}
			if err := json.Unmarshal([]byte(line), &record); err == nil {
				o.logger.Log(record.ToEntry())
			} else {
				o.logger.Infof("Log Parse Failure: %s", line) //nolint
			}
		} else {
			o.logger.Debugf("%s", line)
		}
	}
	if !o.ended.Load() {
		o.output += string(p)
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
		if dump {
			// If we are dumping it is because there was a failure
			// So we want to display the full maven output
			o.logger.Warnf("%s", s)
		}
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
	return o.errors
}

type JvmLogRecord struct {
	Timestamp       time.Time `json:"timestamp"`
	Sequence        int       `json:"sequence"`
	LoggerClassName string    `json:"loggerClassName"`
	LoggerName      string    `json:"loggerName"`
	Level           string    `json:"level"`
	Message         string    `json:"message"`
	ThreadName      string    `json:"threadName"`
	ThreadID        int       `json:"threadId"`
	Mdc             any       `json:"mdc"`
	Ndc             string    `json:"ndc"`
	HostName        string    `json:"hostName"`
	ProcessName     string    `json:"processName"`
	ProcessID       int       `json:"processId"`
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

	ret := log.Entry{
		Time:    r.Timestamp,
		Level:   level,
		Message: r.Message,
	}
	return ret
}
