package config

import (
	"context"
	"time"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	"github.com/block/ftl/common/slices"
)

const ASMSyncInterval = time.Second * 10
const ASMProviderKind ProviderKind = "asm"

type ASMClient interface {
	DeleteSecret(ctx context.Context, input *secretsmanager.DeleteSecretInput, options ...func(*secretsmanager.Options)) (*secretsmanager.DeleteSecretOutput, error)
	CreateSecret(ctx context.Context, input *secretsmanager.CreateSecretInput, options ...func(*secretsmanager.Options)) (*secretsmanager.CreateSecretOutput, error)
	UpdateSecret(ctx context.Context, input *secretsmanager.UpdateSecretInput, options ...func(*secretsmanager.Options)) (*secretsmanager.UpdateSecretOutput, error)
	ListSecrets(ctx context.Context, input *secretsmanager.ListSecretsInput, options ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error)
	BatchGetSecretValue(ctx context.Context, input *secretsmanager.BatchGetSecretValueInput, options ...func(*secretsmanager.Options)) (*secretsmanager.BatchGetSecretValueOutput, error)
}

func NewASMFactory(client ASMClient) (ProviderKind, Factory[Secrets]) {
	return ASMProviderKind, func(ctx context.Context, projectRoot string, key ProviderKey) (BaseProvider[Secrets], error) {
		payload := key.Payload()
		if len(payload) != 1 {
			return nil, errors.Errorf("expected asm:<realm> not %q", key)
		}
		return NewASM(payload[0], client), nil
	}
}

func NewASM(realm string, client ASMClient) *ASM {
	return &ASM{
		realm:  realm,
		client: client,
	}
}

type ASMSecret struct {
	Value   []byte
	Created time.Time
	Updated time.Time
}

type ASM struct {
	realm  string
	client ASMClient
}

var _ AsynchronousProvider[Secrets] = (*ASM)(nil)

func (a *ASM) Key() ProviderKey                { return NewProviderKey(ASMProviderKind, a.realm) }
func (a *ASM) Role() Secrets                   { return Secrets{} }
func (a *ASM) Close(ctx context.Context) error { return nil }
func (a *ASM) SyncInterval() time.Duration     { return ASMSyncInterval }

func (a *ASM) Delete(ctx context.Context, ref Ref) error {
	_, err := a.client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(ref.String()),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	})
	if err != nil {
		return errors.Wrap(err, "asm: failed to delete secret")
	}
	return nil
}

func (a *ASM) Store(ctx context.Context, ref Ref, value []byte) error {
	_, err := a.client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(ref.String()),
		SecretString: aws.String(string(value)),
		Tags: []types.Tag{
			{Key: aws.String("ftl-realm"), Value: aws.String(a.realm)},
			{Key: aws.String("ftl-module"), Value: aws.String(ref.Module.Default(""))},
		},
	})
	// https://github.com/aws/aws-sdk-go-v2/issues/1110#issuecomment-1054643716
	ree := new(types.ResourceExistsException)
	if errors.As(err, &ree) {
		_, err = a.client.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
			SecretId:     aws.String(ref.String()),
			SecretString: aws.String(string(value)),
		})
		if err != nil {
			return errors.Wrap(err, "asm: unable to update secret in ASM")
		}

	} else if err != nil {
		return errors.Wrap(err, "asm: failed to create secret")
	}
	return nil
}

func (a *ASM) Sync(ctx context.Context) (map[Ref]SyncedValue, error) {
	refsToLoad := map[Ref]time.Time{}
	nextToken := optional.None[string]()
	for {
		out, err := a.client.ListSecrets(ctx, &secretsmanager.ListSecretsInput{
			MaxResults:             aws.Int32(100),
			NextToken:              nextToken.Ptr(),
			IncludePlannedDeletion: aws.Bool(false),
			Filters: []types.Filter{
				{Key: types.FilterNameStringTypeTagKey, Values: []string{"ftl-realm"}},
				{Key: types.FilterNameStringTypeTagValue, Values: []string{a.realm}},
			},
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get list of secrets from ASM")
		}
		activeSecrets := slices.Filter(out.SecretList, func(s types.SecretListEntry) bool {
			return s.DeletedDate == nil
		})

		for _, s := range activeSecrets {
			ref := ParseRef(*s.Name)
			refsToLoad[ref] = *s.LastChangedDate
		}

		nextToken = optional.Ptr[string](out.NextToken)
		if !nextToken.Ok() {
			break
		}
	}
	values := map[Ref]SyncedValue{}
	// get values for new and updated secrets
	for len(refsToLoad) > 0 {
		// ASM returns an error when there are more than 10 filters
		// A batch size of 9 + 1 tag filter keeps us within this limit
		batchSize := 9
		secretIDs := []string{}
		for ref := range refsToLoad {
			secretIDs = append(secretIDs, ref.String())
			if len(secretIDs) == batchSize {
				break
			}
		}
		out, err := a.client.BatchGetSecretValue(ctx, &secretsmanager.BatchGetSecretValueInput{SecretIdList: secretIDs})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get batch of secret values from ASM")
		}
		before := len(refsToLoad)
		for _, s := range out.SecretValues {
			ref := ParseRef(*s.Name)
			// Expect secrets to be strings, not binary
			if s.SecretBinary != nil {
				return nil, errors.Errorf("secret for %s in ASM is not a string", ref)
			}
			values[ref] = SyncedValue{
				Value:        []byte(aws.ToString(s.SecretString)),
				VersionToken: optional.Some[VersionToken](refsToLoad[ref]),
			}
			delete(refsToLoad, ref)
		}
		if len(refsToLoad) == before {
			return nil, errors.Errorf("asm: new value of secret not loaded")
		}
	}
	return values, nil
}
