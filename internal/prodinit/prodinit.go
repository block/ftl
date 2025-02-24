// Package prodinit" initializes the runtime environment for production.
package prodinit

import (
	"fmt"
	"os"
	"runtime/debug"

	"go.uber.org/automaxprocs/maxprocs"
)

func init() {
	_, err := maxprocs.Set()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ftl:warning: non-fatal error setting GOMAXPROCS: %v\n", err)
	}
	debug.SetTraceback("all")
}
