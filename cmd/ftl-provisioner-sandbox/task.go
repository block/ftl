package main

import (
	"context"

	"github.com/alecthomas/atomic"

	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/provisioner"
	"github.com/block/ftl/internal/provisioner/state"
)

type task struct {
	runner *provisioner.Runner

	err     atomic.Value[error]
	outputs atomic.Value[[]state.State]
}

func (t *task) Start(oldCtx context.Context) {
	ctx := context.WithoutCancel(oldCtx)
	logger := log.FromContext(ctx)
	go func() {
		outputs, err := t.runner.Run(ctx)
		if err != nil {
			logger.Errorf(err, "failed to execute provisioner")
			t.err.Store(err)
			return
		}
		t.outputs.Store(outputs)
	}()
}
