package main

import (
	"context"
	"syscall"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/kong"

	"github.com/block/ftl/internal/buildengine/languageplugin"
)

type innerCommandPanic struct{}
type KongContextBinder func(ctx context.Context, kctx *kong.Context) context.Context

// runInnerCmd runs a kong command and recovers from a panic if needed
func runInnerCmd(ctx context.Context, k *kong.Kong, binder KongContextBinder, args []string, additionalExit func(int)) error {
	// Overload the exit function to avoid exiting the process
	k.Exit = func(i int) {
		if i != 0 {
			if additionalExit != nil {
				additionalExit(i)
			}
			_ = syscall.Kill(-syscall.Getpid(), syscall.SIGINT) //nolint:forcetypeassert,errcheck
		}
		// For a normal exit from an interactive command we need a special panic
		// we recover from this and continue the loop
		panic(innerCommandPanic{})
	}

	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(innerCommandPanic); ok {
				return
			}
			panic(r)
		}
	}()
	// Dynamically update the kong app with language specific flags for the "ftl module new" command.
	err := languageplugin.PrepareNewCmd(ctx, k, args)
	if err != nil {
		return errors.Wrap(err, "could not prepare for command")
	}
	kctx, err := k.Parse(args)
	if err != nil {
		return errors.WithStack(err) //nolint:wrapcheck
	}
	subctx := binder(ctx, kctx)

	err = kctx.Run(subctx)
	if err != nil {
		return errors.WithStack(err) //nolint:wrapcheck
	}
	return nil
}
