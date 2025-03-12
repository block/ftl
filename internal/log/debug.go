package log

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl-golang-tools/go/ssa/interp/testdata/src/sync"
)

const logFileName = "log.json"

func SetupDebugFileLogging(ctx context.Context, root string, maxLogs int) {
	logDir := filepath.Join(root, ".ftl-dev-logs")
	logger := FromContext(ctx)
	if err := os.MkdirAll(logDir, 0755); err != nil { //nolint:gosec
		logger.Errorf(err, "Could not create log directory %s", logDir)
		return
	}
	sink := &logRotatingSink{maxLogs: maxLogs, logDir: logDir}
	context.AfterFunc(ctx, func() {
		sink.lock.Lock()
		defer sink.lock.Unlock()
		if sink.file != nil {
			sink.file.Close() // #nolint: errcheck
		}
	})
	logger.debugLogger = optional.Some(devModeDebugLogger{delegate: New(Debug, sink), sink: sink})
}

var _ Sink = (*logRotatingSink)(nil)

type logRotatingSink struct {
	maxLogs int
	count   int
	lock    sync.Mutex
	file    *os.File
	logDir  string
	sink    Sink
}

func (r *logRotatingSink) Log(entry Entry) error {
	var err error
	r.lock.Lock()
	defer r.lock.Unlock()
	r.count++
	if r.file == nil {
		r.file, err = rotateLogs(r.logDir, r.maxLogs)
		r.sink = newJSONSink(r.file)
		if err != nil {
			return fmt.Errorf("could not rotate logs: %w", err)
		}
	} else if r.count%100 == 0 {
		// We only check the size every hundred logs to avoid a performance hit
		size, err := r.file.Stat()
		if err != nil {
			return fmt.Errorf("could not stat log file: %w", err)
		}
		if size.Size() > 1024*1024*10 {
			r.file.Close() // #nolint: errcheck
			r.file, err = rotateLogs(r.logDir, r.maxLogs)
			r.sink = newJSONSink(r.file)
			if err != nil {
				return fmt.Errorf("could not rotate logs: %w", err)
			}
		}
	}
	err = r.sink.Log(entry)
	if err != nil {
		return fmt.Errorf("could not write log entry: %w", err)
	}
	return nil
}

func rotateLogs(workingDir string, maxLogs int) (*os.File, error) {
	// Remove the oldest log file if it exists
	oldestLog := filepath.Join(workingDir, fmt.Sprintf("log.%d.json", maxLogs))
	if _, err := os.Stat(oldestLog); err == nil {
		os.Remove(oldestLog)
	}

	// Shift log files up
	for i := maxLogs - 1; i > 0; i-- {
		oldName := filepath.Join(workingDir, fmt.Sprintf("log.%d.json", i))
		newName := filepath.Join(workingDir, fmt.Sprintf("log.%d.json", i+1))
		if _, err := os.Stat(oldName); err == nil {
			err := os.Rename(oldName, newName)
			if err != nil {
				return nil, fmt.Errorf("could not rename log file %s to %s: %w", oldName, newName, err)
			}
		}
	}

	// Rename the current log file if it exists
	currentLog := filepath.Join(workingDir, logFileName)
	if _, err := os.Stat(currentLog); err == nil {
		err := os.Rename(currentLog, filepath.Join(workingDir, "log.1.json"))
		if err != nil {
			return nil, fmt.Errorf("could not rename log file %s: %w", currentLog, err)
		}
	}

	logFile, err := os.Create(filepath.Join(workingDir, logFileName))
	if err != nil {
		return nil, fmt.Errorf("could not create log file: %w", err)
	}

	return logFile, nil
}

// ReplayLogs replays the debug logs to the delegate logger
func ReplayLogs(ctx context.Context) {
	logger := FromContext(ctx)
	if t, ok := logger.debugLogger.Get(); ok {
		go func() {
			fileName := t.sink.file.Name()
			// Make a copy of the logger and remove the debug logger
			// This is so we don't re-write all the messages to the log file
			logCopy := *logger
			logCopy.debugLogger = optional.None[devModeDebugLogger]()
			file, err := os.Open(fileName)
			if err != nil {
				logger.Errorf(err, "Could not open log file")
			}
			defer file.Close() // #nolint: errcheck
			err = JSONStreamer(file, &logCopy, Debug)
			if err != nil {
				logger.Errorf(err, "Could not replay log file")
			}
		}()
	}
}

type devModeDebugLogger struct {
	delegate *Logger
	sink     *logRotatingSink
}
