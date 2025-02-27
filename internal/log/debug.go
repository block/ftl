package log

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/types/optional"
)

const logFileName = "log.json"

func SetupDebugFileLogging(ctx context.Context, root string, maxLogs int) {
	logDir := filepath.Join(root, ".ftl-dev-logs")
	logger := FromContext(ctx)
	if err := os.MkdirAll(logDir, 0755); err != nil { //nolint:gosec
		logger.Errorf(err, "Could not create log directory %s", logDir)
		return
	}
	if err := rotateLogs(logDir, maxLogs); err != nil {
		logger.Errorf(err, "Could not rotate logs")
		return
	}
	logFile, err := os.Create(filepath.Join(logDir, logFileName))
	if err != nil {
		logger.Errorf(err, "Could not create log file")
		return
	}
	context.AfterFunc(ctx, func() {
		if logFile != nil {
			logFile.Close() // #nolint: errcheck
		}
	})
	sink := newJSONSink(logFile)
	logger.debugLogger = optional.Some(devModeDebugLogger{delegate: New(Debug, sink), file: logFile.Name()})
}

func rotateLogs(workingDir string, maxLogs int) error {
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
				return fmt.Errorf("could not rename log file %s to %s: %w", oldName, newName, err)
			}
		}
	}

	// Rename the current log file if it exists
	currentLog := filepath.Join(workingDir, logFileName)
	if _, err := os.Stat(currentLog); err == nil {
		err := os.Rename(currentLog, filepath.Join(workingDir, "log.1.json"))
		if err != nil {
			return fmt.Errorf("could not rename log file %s: %w", currentLog, err)
		}
	}

	return nil
}

// ReplayLogs replays the debug logs to the delegate logger
func ReplayLogs(ctx context.Context) {
	logger := FromContext(ctx)
	if t, ok := logger.debugLogger.Get(); ok {
		go func() {
			fileName := t.file
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
	file     string
}
