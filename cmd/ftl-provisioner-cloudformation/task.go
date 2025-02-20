package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alecthomas/atomic"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/jpillora/backoff"

	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/provisioner"
	"github.com/block/ftl/internal/provisioner/state"
)

type task struct {
	runner *provisioner.Runner

	err     atomic.Value[error]
	outputs atomic.Value[[]state.State]
}

func waitForStackReady(ctx context.Context, stackID string, client *cloudformation.Client) ([]types.Output, error) {
	retry := backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
	}
	for {
		desc, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
			StackName: &stackID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe stack: %w", err)
		}
		stack := desc.Stacks[0]

		switch stack.StackStatus {
		// noop while running
		case types.StackStatusCreateInProgress:
		case types.StackStatusUpdateInProgress:
		case types.StackStatusUpdateCompleteCleanupInProgress:
		case types.StackStatusUpdateRollbackInProgress:

		// success
		case types.StackStatusCreateComplete:
			return stack.Outputs, nil
		case types.StackStatusDeleteComplete:
			return stack.Outputs, nil
		case types.StackStatusUpdateComplete:
			return stack.Outputs, nil

		// failures
		case types.StackStatusCreateFailed:
			return nil, fmt.Errorf("stack creation failed: %s", *stack.StackStatusReason)
		case types.StackStatusRollbackInProgress:
			return nil, fmt.Errorf("stack rollback in progress: %s", *stack.StackStatusReason)
		case types.StackStatusRollbackFailed:
			return nil, fmt.Errorf("stack rollback failed: %s", *stack.StackStatusReason)
		case types.StackStatusRollbackComplete:
			return nil, fmt.Errorf("stack rollback complete: %s", *stack.StackStatusReason)
		case types.StackStatusDeleteInProgress:
		case types.StackStatusDeleteFailed:
			return nil, fmt.Errorf("stack deletion failed: %s", *stack.StackStatusReason)
		case types.StackStatusUpdateFailed:
			return nil, fmt.Errorf("stack update failed: %s", *stack.StackStatusReason)
		default:
			return nil, fmt.Errorf("unsupported Cloudformation status code: %s", string(stack.StackStatus))
		}

		time.Sleep(retry.Duration())
	}
}

func (t *task) Start(oldCtx context.Context, client *cloudformation.Client, secrets *secretsmanager.Client, changeSetID string) {
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
