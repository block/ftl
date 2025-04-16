package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/akamensky/base58"
	errors "github.com/alecthomas/errors"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/smithy-go"
	goformation "github.com/awslabs/goformation/v7/cloudformation"
	"github.com/jpillora/backoff"
)

// ensureStackExists and if not, creates an empty stack with the given name
//
// Returns, when the stack is ready
func ensureStackExists(ctx context.Context, client *cloudformation.Client, name string) error {
	_, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: &name,
	})
	var ae smithy.APIError
	// Not Found is returned as ValidationError from the AWS API
	if errors.As(err, &ae) && ae.ErrorCode() == "ValidationError" {
		empty := `
		{
			"Resources": {
				  "NullResource": {
					"Type": "AWS::CloudFormation::WaitConditionHandle"
				}
			}
		}
		`
		if _, err := client.CreateStack(ctx, &cloudformation.CreateStackInput{
			StackName:    &name,
			TemplateBody: &empty,
		}); err != nil {
			return errors.Wrapf(err, "failed to create stack %s", name)
		}

		if err := waitStackReady(ctx, client, name); err != nil {
			return errors.Wrapf(err, "stack %s did not become ready", name)
		}

	} else if err != nil {
		return errors.Wrapf(err, "failed to describe stack %s", name)
	}
	return nil
}

// waitChangeSetReady returns when the given changeset either became ready, or resulted into no changes error.
func waitChangeSetReady(ctx context.Context, client *cloudformation.Client, changeSet, stack string) (hadChanges bool, err error) {
	retry := backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
	}
	for {
		desc, err := client.DescribeChangeSet(ctx, &cloudformation.DescribeChangeSetInput{
			ChangeSetName: &changeSet,
			StackName:     &stack,
		})
		if err != nil {
			return false, errors.Wrap(err, "failed to describe change-set")
		}
		if desc.Status == types.ChangeSetStatusFailed {
			// Unfortunately, there does not seem to be a better way to do this
			if *desc.StatusReason == "The submitted information didn't contain changes. Submit different information to create a change set." {
				// clean up the changeset if there were no changes
				_, err := client.DeleteChangeSet(ctx, &cloudformation.DeleteChangeSetInput{
					ChangeSetName: &changeSet,
					StackName:     &stack,
				})
				if err != nil {
					return false, errors.Wrapf(err, "failed to delete change-set %s", changeSet)
				}
				return false, nil
			}
			return false, errors.WithStack(errors.New(*desc.StatusReason))
		}
		if desc.Status != types.ChangeSetStatusCreatePending && desc.Status != types.ChangeSetStatusCreateInProgress {
			return true, nil
		}
		time.Sleep(retry.Duration())
	}
}

// createClient for interacting with Cloudformation
func createClient(ctx context.Context) (*cloudformation.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load default aws config")
	}

	return cloudformation.New(
		cloudformation.Options{
			Credentials: cfg.Credentials,
			Region:      cfg.Region,
		},
	), nil
}

func createSecretsClient(ctx context.Context) (*secretsmanager.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load default aws config")
	}
	return secretsmanager.New(
		secretsmanager.Options{
			Credentials: cfg.Credentials,
			Region:      cfg.Region,
		},
	), nil
}

// CloudformationOutputKey is structured key to be used as an output from a CF stack
type CloudformationOutputKey struct {
	ResourceID   string `json:"r"`
	ResourceKind string `json:"k"`
	PropertyName string `json:"p"`
}

// decodeOutputKey reads the structured CloudformationOutputKey from the given stack output
func decodeOutputKey(output types.Output) (*CloudformationOutputKey, error) {
	rawKey := *output.OutputKey
	bytes, err := base58.Decode(rawKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode cloudformation output key")
	}
	key := CloudformationOutputKey{}
	if err := json.Unmarshal(bytes, &key); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal cloudformation output key")
	}
	return &key, nil
}

// addOutput to the given goformation.Outputs
//
// Encodes the given CloudformationOutputKey, and uses the goformation value as the value.
func addOutput(to goformation.Outputs, value any, key *CloudformationOutputKey) {
	desc := string(outputKeyJSON(key))
	to[base58.Encode(outputKeyJSON(key))] = goformation.Output{
		Value:       value,
		Description: &desc,
	}
}

func outputKeyJSON(key *CloudformationOutputKey) []byte {
	bytes, err := json.Marshal(key)
	if err != nil {
		panic(err)
	}
	return bytes
}

func waitStackReady(ctx context.Context, client *cloudformation.Client, name string) error {
	retry := backoff.Backoff{
		Min:    100 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
	}
	for {
		state, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
			StackName: &name,
		})
		if err != nil {
			return errors.Wrap(err, "failed to describe stack")
		}
		if state.Stacks[0].StackStatus == types.StackStatusCreateFailed {
			return errors.WithStack(errors.New(*state.Stacks[0].StackStatusReason))
		}
		if state.Stacks[0].StackStatus != types.StackStatusCreateInProgress {
			return nil
		}
		time.Sleep(retry.Duration())
	}
}

func getStackOutputs(ctx context.Context, stackID string, client *cloudformation.Client) ([]types.Output, error) {
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
			return nil, errors.Wrap(err, "failed to describe stack")
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
			return nil, errors.Errorf("stack creation failed: %s", *stack.StackStatusReason)
		case types.StackStatusRollbackInProgress:
			return nil, errors.Errorf("stack rollback in progress: %s", *stack.StackStatusReason)
		case types.StackStatusRollbackFailed:
			return nil, errors.Errorf("stack rollback failed: %s", *stack.StackStatusReason)
		case types.StackStatusRollbackComplete:
			return nil, errors.Errorf("stack rollback complete: %s", *stack.StackStatusReason)
		case types.StackStatusDeleteInProgress:
		case types.StackStatusDeleteFailed:
			return nil, errors.Errorf("stack deletion failed: %s", *stack.StackStatusReason)
		case types.StackStatusUpdateFailed:
			return nil, errors.Errorf("stack update failed: %s", *stack.StackStatusReason)
		default:
			return nil, errors.Errorf("unsupported Cloudformation status code: %s", string(stack.StackStatus))
		}

		time.Sleep(retry.Duration())
	}
}
