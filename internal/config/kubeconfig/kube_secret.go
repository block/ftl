package config

import (
	"context"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/block/ftl/internal/config"
	"github.com/block/ftl/internal/kube"
)

const KubeSecretProviderKind config.ProviderKind = "kube-secret"

var _ config.Provider[config.Secrets] = &KubeSecretProvider{}

type KubeSecretProvider struct {
	client *kubernetes.Clientset
	mapper kube.NamespaceMapper
	realm  string
}

// Close implements Provider.
func (k *KubeSecretProvider) Close(ctx context.Context) error {
	return nil
}

// Delete implements Provider.
func (k *KubeSecretProvider) Delete(ctx context.Context, ref config.Ref) error {
	if module, ok := ref.Module.Get(); ok {
		conf, err := k.load(ctx, module)
		if err != nil {
			return errors.Wrapf(err, "set %s", ref)
		}
		delete(conf, ref.Name)
		if err = k.save(ctx, module, conf); err != nil {
			return errors.Wrapf(err, "set %s", ref)
		}
		return nil
	}
	return errors.Errorf("unable to unset global config")
}

// Key implements Provider.
func (k *KubeSecretProvider) Key() config.ProviderKey {
	return config.NewProviderKey(KubeSecretProviderKind)
}

// List implements Provider.
func (k *KubeSecretProvider) List(ctx context.Context, withValues bool, forModule optional.Option[string]) ([]config.Value, error) {
	// Not implemented yet
	return []config.Value{}, nil
}

// Load implements Provider.
func (k *KubeSecretProvider) Load(ctx context.Context, ref config.Ref) ([]byte, error) {
	if module, ok := ref.Module.Get(); ok {
		ns := k.mapper(module, k.realm)
		cm, err := k.client.CoreV1().Secrets(ns).Get(ctx, kube.SecretName(module), v1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, nil
			}
			return nil, errors.Wrap(err, "failed to get Secret")
		}
		val := cm.Data[ref.Name]
		if len(val) == 0 {
			return nil, nil
		}
		return val, nil
	}
	return nil, nil
}

// Role implements Provider.
func (k *KubeSecretProvider) Role() config.Secrets {
	return config.Secrets{}
}

// Store implements Provider.
func (k *KubeSecretProvider) Store(ctx context.Context, ref config.Ref, value []byte) error {

	if module, ok := ref.Module.Get(); ok {
		conf, err := k.load(ctx, module)
		if err != nil {
			return errors.Wrapf(err, "set %s", ref)
		}
		conf[ref.Name] = value
		if err = k.save(ctx, module, conf); err != nil {
			return errors.Wrapf(err, "set %s", ref)
		}
		return nil
	}
	return errors.Errorf("unable to set global config")
}

func NewKubeSecretProvider(client *kubernetes.Clientset, mapper kube.NamespaceMapper, realm string) *KubeSecretProvider {
	return &KubeSecretProvider{client: client, mapper: mapper, realm: realm}
}

func (k *KubeSecretProvider) load(ctx context.Context, module string) (map[string][]byte, error) {
	ns := k.mapper(module, k.realm)
	cm, err := k.client.CoreV1().Secrets(ns).Get(ctx, kube.SecretName(module), v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return map[string][]byte{}, nil
		}
		return nil, errors.Wrap(err, "failed to get Secret")
	}
	serialisable := cm.Data
	if cm.Data == nil {
		return make(map[string][]byte), nil
	}

	out := map[string][]byte{}
	for refStr, keyStr := range serialisable {
		out[refStr] = keyStr
	}
	return out, nil
}

func (k *KubeSecretProvider) save(ctx context.Context, module string, data map[string][]byte) error {
	serialisable := map[string][]byte{}
	for ref, val := range data {
		serialisable[ref] = val
	}
	ns := k.mapper(module, k.realm)
	err := kube.EnsureNamespace(ctx, k.client, k.mapper(module, k.realm), k.realm)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespace")
	}
	cm, err := k.client.CoreV1().Secrets(ns).Get(ctx, kube.SecretName(module), v1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get Secret")
		}
		cm.Name = kube.SecretName(module)
		cm.Namespace = ns
		cm.Labels = map[string]string{"app.kubernetes.io/managed-by": "ftl"}
		cm, err = k.client.CoreV1().Secrets(ns).Create(ctx, cm, v1.CreateOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to create Secret")
		}
	}
	cm.Data = serialisable

	_, err = k.client.CoreV1().Secrets(ns).Update(ctx, cm, v1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to update Secret")
	}
	return nil
}
