package config_test

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/errors"
	. "github.com/alecthomas/types/optional"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/config"
)

func TestASM(t *testing.T) {
	t.Parallel()
	ctx := log.ContextWithNewDefaultLogger(t.Context())
	client := NewFakeASM()
	asmProvider := OverrideSyncInterval(config.NewASM("test", client), time.Second)
	provider := config.NewCacheDecorator(ctx, asmProvider)
	testConfig(t, ctx, provider)

	// Do an OOB update to ensure the provider picks it up via background sync
	_, err := client.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
		SecretId:     aws.String("name"),
		SecretString: aws.String(`"Eve"`),
	})
	assert.NoError(t, err)
	time.Sleep(time.Second * 2)
	secrets, err := provider.List(ctx, true, None[string]())
	assert.NoError(t, err)
	assert.Equal(t, []config.Value{
		{
			Ref:   config.ParseRef("name"),
			Value: Some([]byte(`"Eve"`)),
		},
	}, secrets)
}

type ASMSecret struct {
	ID        string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewFakeASM() *FakeASM {
	return &FakeASM{secrets: map[string]ASMSecret{}}
}

type FakeASM struct {
	secrets map[string]ASMSecret
}

func (a *FakeASM) BatchGetSecretValue(ctx context.Context, input *secretsmanager.BatchGetSecretValueInput, options ...func(*secretsmanager.Options)) (*secretsmanager.BatchGetSecretValueOutput, error) {
	out := &secretsmanager.BatchGetSecretValueOutput{}
	for _, id := range input.SecretIdList {
		if secret, ok := a.secrets[id]; ok {
			out.SecretValues = append(out.SecretValues, types.SecretValueEntry{
				Name:         &id,
				SecretString: &secret.Value,
				CreatedDate:  &secret.CreatedAt,
			})
		}
	}
	return out, nil
}

func (a *FakeASM) CreateSecret(ctx context.Context, input *secretsmanager.CreateSecretInput, options ...func(*secretsmanager.Options)) (*secretsmanager.CreateSecretOutput, error) {
	_, ok := a.secrets[*input.Name]
	if ok {
		return nil, errors.WithStack(&types.ResourceExistsException{Message: aws.String("secret already exists")})
	}
	a.secrets[*input.Name] = ASMSecret{
		ID:        *input.Name,
		Value:     *input.SecretString,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return &secretsmanager.CreateSecretOutput{Name: input.Name}, nil
}

func (a *FakeASM) DeleteSecret(ctx context.Context, input *secretsmanager.DeleteSecretInput, options ...func(*secretsmanager.Options)) (*secretsmanager.DeleteSecretOutput, error) {
	_, ok := a.secrets[*input.SecretId]
	if !ok {
		return nil, errors.WithStack(&types.ResourceNotFoundException{Message: aws.String("secret not found")})
	}
	delete(a.secrets, *input.SecretId)
	return &secretsmanager.DeleteSecretOutput{Name: input.SecretId}, nil
}

func (a *FakeASM) ListSecrets(ctx context.Context, input *secretsmanager.ListSecretsInput, options ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error) {
	var secrets []types.SecretListEntry
	for _, secret := range a.secrets {
		secrets = append(secrets, types.SecretListEntry{
			Name:            aws.String(secret.ID),
			CreatedDate:     aws.Time(secret.CreatedAt),
			LastChangedDate: aws.Time(secret.UpdatedAt),
		})
	}
	return &secretsmanager.ListSecretsOutput{SecretList: secrets}, nil
}

func (a *FakeASM) UpdateSecret(ctx context.Context, input *secretsmanager.UpdateSecretInput, options ...func(*secretsmanager.Options)) (*secretsmanager.UpdateSecretOutput, error) {
	existing, ok := a.secrets[*input.SecretId]
	if !ok {
		return nil, errors.WithStack(&types.ResourceNotFoundException{Message: aws.String("secret not found")})
	}
	a.secrets[*input.SecretId] = ASMSecret{
		ID:        *input.SecretId,
		Value:     *input.SecretString,
		CreatedAt: existing.CreatedAt,
		UpdatedAt: time.Now(),
	}
	return &secretsmanager.UpdateSecretOutput{Name: input.SecretId}, nil
}
