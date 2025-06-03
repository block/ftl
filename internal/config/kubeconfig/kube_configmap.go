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

const KubeConfigMapProviderKind config.ProviderKind = "kube-config-map"

var _ config.Provider[config.Configuration] = &KubeConfigMapProvider{}

type KubeConfigMapProvider struct {
	client *kubernetes.Clientset
	mapper kube.NamespaceMapper
	realm  string
}

// Close implements Provider.
func (k *KubeConfigMapProvider) Close(ctx context.Context) error {
	return nil
}

// Delete implements Provider.
func (k *KubeConfigMapProvider) Delete(ctx context.Context, ref config.Ref) error {
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
func (k *KubeConfigMapProvider) Key() config.ProviderKey {
	return config.NewProviderKey(KubeConfigMapProviderKind)
}

// List implements Provider.
func (k *KubeConfigMapProvider) List(ctx context.Context, withValues bool, forModule optional.Option[string]) ([]config.Value, error) {
	// Not implemented yet
	return []config.Value{}, nil
}

// Load implements Provider.
func (k *KubeConfigMapProvider) Load(ctx context.Context, ref config.Ref) ([]byte, error) {
	if module, ok := ref.Module.Get(); ok {
		ns := k.mapper(module, k.realm)
		cm, err := k.client.CoreV1().ConfigMaps(ns).Get(ctx, kube.ConfigMapName(module), v1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, nil
			}
			return nil, errors.Wrap(err, "failed to get ConfigMap")
		}
		val := cm.Data[ref.Name]
		if val == "" {
			return nil, nil
		}
		return []byte(val), nil
	}
	return nil, nil
}

// Role implements Provider.
func (k *KubeConfigMapProvider) Role() config.Configuration {
	return config.Configuration{}
}

// Store implements Provider.
func (k *KubeConfigMapProvider) Store(ctx context.Context, ref config.Ref, value []byte) error {

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

func NewKubeConfigProvider(client *kubernetes.Clientset, mapper kube.NamespaceMapper, realm string) *KubeConfigMapProvider {
	return &KubeConfigMapProvider{client: client, mapper: mapper, realm: realm}
}

func (k *KubeConfigMapProvider) load(ctx context.Context, module string) (map[string][]byte, error) {
	ns := k.mapper(module, k.realm)
	cm, err := k.client.CoreV1().ConfigMaps(ns).Get(ctx, kube.ConfigMapName(module), v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return map[string][]byte{}, nil
		}
		return nil, errors.Wrap(err, "failed to get ConfigMap")
	}
	serialisable := cm.Data
	if cm.Data == nil {
		return make(map[string][]byte), nil
	}

	out := map[string][]byte{}
	for refStr, keyStr := range serialisable {
		out[refStr] = []byte(keyStr)
	}
	return out, nil
}

func (k *KubeConfigMapProvider) save(ctx context.Context, module string, data map[string][]byte) error {
	serialisable := map[string]string{}
	for ref, key := range data {
		serialisable[ref] = string(key)
	}
	ns := k.mapper(module, k.realm)
	err := kube.EnsureNamespace(ctx, k.client, k.mapper(module, k.realm), k.realm)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespace")
	}
	cm, err := k.client.CoreV1().ConfigMaps(ns).Get(ctx, kube.ConfigMapName(module), v1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get ConfigMap")
		}
		cm.Name = kube.ConfigMapName(module)
		cm.Namespace = ns
		cm.Labels = map[string]string{"app.kubernetes.io/managed-by": "ftl"}
		cm, err = k.client.CoreV1().ConfigMaps(ns).Create(ctx, cm, v1.CreateOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to create ConfigMap")
		}
	}
	cm.Data = serialisable

	_, err = k.client.CoreV1().ConfigMaps(ns).Update(ctx, cm, v1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to update ConfigMap")
	}
	return nil
}
