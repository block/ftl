package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	goformation "github.com/awslabs/goformation/v7/cloudformation"
	"github.com/block/ftl/cmd/ftl-provisioner-cloudformation/executor"
)

type CFExecutor struct {
	cfn      *cloudformation.Client
	secrets  *secretsmanager.Client
	template *goformation.Template

	inputs []executor.State
}

func NewCFExecutor(ctx context.Context, cfn *cloudformation.Client, secrets *secretsmanager.Client) (*CFExecutor, error) {
	e := &CFExecutor{
		cfn:      cfn,
		secrets:  secrets,
		template: goformation.NewTemplate(),
	}

	return e, nil
}

func (e *CFExecutor) Stage() executor.Stage {
	return executor.StageProvisioning
}

func (e *CFExecutor) Resources() []executor.ResourceKind {
	return executor.AllResources
}

func (e *CFExecutor) Prepare(ctx context.Context, input executor.State) error {
	e.inputs = append(e.inputs, input)
	return nil
}
