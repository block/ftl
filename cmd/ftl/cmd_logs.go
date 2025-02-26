package main

type logsCmd struct {
	Level logsSetLevelCmd `cmd:"" help:"Set the current log level"`
}
