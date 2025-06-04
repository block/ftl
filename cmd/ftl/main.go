package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/errors"

	_ "github.com/block/ftl/internal/prodinit" // Set GOMAXPROCS to match Linux container CPU quota.
)

func main() {
	ctx, cancel := context.WithCancelCause(context.Background())
	// Handle signals.
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigch
		cancel(errors.Wrapf(context.Canceled, "FTL terminating with signal %s", sig))
		_ = syscall.Kill(-syscall.Getpid(), sig.(syscall.Signal)) //nolint:forcetypeassert,errcheck // best effort
		os.Exit(0)
	}()
	app, err := New(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ftl: error: %s\n", err)
		os.Exit(1)
	}
	err = app.Run(ctx, os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ftl: error: %s\n", err)
		os.Exit(1)
	}
}
