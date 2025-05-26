package main

import (
	"context"
	"fmt"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	goformation "github.com/awslabs/goformation/v7/cloudformation"
	cf "github.com/awslabs/goformation/v7/cloudformation/cloudformation"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/provisioner/state"
)

type CloudFormationExecutor struct {
	cfn      *cloudformation.Client
	secrets  *secretsmanager.Client
	template *goformation.Template
	config   *Config
	stack    string

	inputs []state.State
}

func NewCloudFormationExecutor(stack string, cfn *cloudformation.Client, secrets *secretsmanager.Client, config *Config) *CloudFormationExecutor {
	return &CloudFormationExecutor{
		cfn:      cfn,
		secrets:  secrets,
		template: goformation.NewTemplate(),
		stack:    stack,
		config:   config,
	}
}

func (e *CloudFormationExecutor) Prepare(ctx context.Context, input state.State) error {
	e.inputs = append(e.inputs, input)

	for _, resource := range e.inputs {
		switch r := resource.(type) {
		case state.InputPostgres:
			tmpl := &PostgresTemplater{input: r, config: e.config}
			if err := tmpl.AddToTemplate(e.template); err != nil {
				return errors.Wrap(err, "failed to add postgres template")
			}
		case state.InputMySQL:
			tmpl := &MySQLTemplater{input: r, config: e.config}
			if err := tmpl.AddToTemplate(e.template); err != nil {
				return errors.Wrap(err, "failed to add mysql template")
			}
		default:
			return errors.Errorf("unknown resource type: %T", r)
		}
	}

	if len(e.inputs) == 0 {
		// Stack can not be empty, insert a null resource to keep the stack around
		e.template.Resources["NullResource"] = &cf.WaitConditionHandle{}
	}
	return nil
}

func (e *CloudFormationExecutor) Execute(ctx context.Context) ([]state.State, error) {
	logger := log.FromContext(ctx)

	changeSet := generateChangeSetName(e.stack)
	templateStr, err := e.template.JSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudformation template")
	}
	if err := ensureStackExists(ctx, e.cfn, e.stack); err != nil {
		return nil, errors.Wrap(err, "failed to verify the stack exists")
	}

	logger.Debugf("creating change-set to stack %s", e.stack)
	resp, err := e.cfn.CreateChangeSet(ctx, &cloudformation.CreateChangeSetInput{
		StackName:     &e.stack,
		ChangeSetName: &changeSet,
		TemplateBody:  ptr(string(templateStr)),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create change-set")
	}
	logger.Debugf("waiting for change-set %s to stack %s to become ready", *resp.Id, e.stack)
	hadChanges, err := waitChangeSetReady(ctx, e.cfn, changeSet, e.stack)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wait for change-set to become ready")
	}

	if hadChanges {
		logger.Debugf("executing change-set %s to stack %s", *resp.Id, e.stack)
		_, err = e.cfn.ExecuteChangeSet(ctx, &cloudformation.ExecuteChangeSetInput{
			ChangeSetName: &changeSet,
			StackName:     &e.stack,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to execute change-set")
		}
	} else {
		logger.Debugf("no changes to execute for change-set %s to stack %s", *resp.Id, e.stack)
	}

	logger.Debugf("waiting for stack %s to be ready", e.stack)
	cfOutputs, err := getStackOutputs(ctx, e.stack, e.cfn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to wait for stack to be ready")
	}

	byResourceID, err := outputsByResourceID(cfOutputs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to group outputs by resource ID")
	}

	outputs := make([]state.State, 0, len(e.inputs))
	for _, input := range e.inputs {
		switch r := input.(type) {
		case state.InputPostgres:
			logger.Debugf("finding outputs for postgres resource %s", r.ResourceID)
			res := byResourceID[r.ResourceID]
			outputs = append(outputs, state.RDSInstanceReadyPostgres{
				Module:     r.Module,
				ResourceID: r.ResourceID,

				MasterUserSecretARN: findValue(ctx, res, PropertyPsqlMasterUserARN),
				WriteEndpoint:       fmt.Sprintf("%s:%d", findValue(ctx, res, PropertyPsqlWriteEndpoint), PostgresPort),
				ReadEndpoint:        fmt.Sprintf("%s:%d", findValue(ctx, res, PropertyPsqlReadEndpoint), PostgresPort),
			})
		case state.InputMySQL:
			logger.Debugf("finding outputs for mysql resource %s", r.ResourceID)
			res := byResourceID[r.ResourceID]
			outputs = append(outputs, state.RDSInstanceReadyMySQL{
				Module:     r.Module,
				ResourceID: r.ResourceID,

				MasterUserSecretARN: optional.Some[string](findValue(ctx, res, PropertyMySQLMasterUserARN)),
				WriteEndpoint:       fmt.Sprintf("%s:%d", findValue(ctx, res, PropertyMySQLWriteEndpoint), MySQLPort),
				ReadEndpoint:        fmt.Sprintf("%s:%d", findValue(ctx, res, PropertyMySQLReadEndpoint), MySQLPort),
			})
		}
	}

	logger.Debugf("stack %s is ready", e.stack)

	return outputs, nil
}

func findValue(ctx context.Context, outputs []types.Output, property string) string {
	logger := log.FromContext(ctx)
	for _, output := range outputs {
		decoded, err := decodeOutputKey(output)
		if err != nil {
			logger.Warnf("failed to decode output key: %s", err)
			continue
		}
		if decoded.PropertyName == property {
			return *output.OutputValue
		}
	}
	logger.Warnf("no output found for key %s", property)
	return ""
}
