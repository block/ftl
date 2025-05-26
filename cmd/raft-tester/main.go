package main

import (
	"context"
	"fmt"
	"net/url"
	"time"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/kong"
	"golang.org/x/exp/rand"
	"golang.org/x/sync/errgroup"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/raft"
	sm "github.com/block/ftl/internal/statemachine"
)

var cli struct {
	RaftConfig raft.RaftConfig `embed:"" prefix:"raft-"`

	Start startCmd `cmd:"" help:"Start the raft tester cluster."`
	Join  joinCmd  `cmd:"" help:"Join the raft tester cluster."`
}

type startCmd struct{}

func (s *startCmd) Run() error {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	builder := raft.NewBuilder(&cli.RaftConfig)
	shard := raft.AddShard(ctx, builder, 1, &IntStateMachine{})
	cluster := builder.Build(ctx)

	if err := cluster.Start(ctx); err != nil {
		return errors.Wrap(err, "failed to start cluster")
	}
	defer cluster.Stop(ctx)

	return errors.WithStack(run(ctx, shard))
}

type joinCmd struct {
	ControlAddress *url.URL `help:"Control address to use to join the cluster."`
}

func (j *joinCmd) Run() error {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	builder := raft.NewBuilder(&cli.RaftConfig)
	shard := raft.AddShard(ctx, builder, 1, &IntStateMachine{})
	cluster := builder.Build(ctx)

	if err := cluster.Join(ctx, j.ControlAddress.String()); err != nil {
		return errors.Wrap(err, "failed to join cluster")
	}
	defer cluster.Stop(ctx)

	return errors.WithStack(run(ctx, shard))
}

func run(ctx context.Context, shard sm.Handle[int64, int64, IntEvent]) error {
	messages := make(chan int)

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		defer close(messages)
		// send a random number every 10 seconds
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				messages <- rand.Intn(1000)
			case <-ctx.Done():
				return nil
			}
		}
	})

	changes, err := shard.StateIter(ctx, 1)
	if err != nil {
		return errors.Wrap(err, "failed to get changes")
	}

	wg.Go(func() error {
		for {
			select {
			case msg := <-messages:
				err := shard.Publish(ctx, IntEvent(msg))
				if err != nil {
					return errors.Wrap(err, "failed to propose event")
				}
			case <-ctx.Done():
				return nil
			}
		}
	})

	go func() {
		for c := range changes {
			fmt.Println("state: ", c)
		}
	}()

	if err := wg.Wait(); err != nil {
		return errors.Wrap(err, "failed to run")
	}

	return nil
}

func main() {
	kctx := kong.Parse(&cli)

	if err := kctx.Run(); err != nil {
		kctx.FatalIfErrorf(err)
	}
}
