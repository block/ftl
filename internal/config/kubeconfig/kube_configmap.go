package kubeconfig

import (
	"context"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	kubecore "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/block/ftl/common/log"
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
		conf, cm, err := k.load(ctx, module)
		if err != nil {
			return errors.Wrapf(err, "set %s", ref)
		}
		delete(conf, ref.Name)
		if err = k.save(ctx, module, conf, cm); err != nil {
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
	ns := ""
	mod := ""
	ok := false
	if mod, ok = forModule.Get(); ok {
		ns = k.mapper(mod, k.realm)
	}
	maps, err := k.client.CoreV1().ConfigMaps(ns).List(ctx, v1.ListOptions{LabelSelector: kube.RealmLabel + "=" + k.realm})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ConfigMaps")
	}
	ret := []config.Value{}
	for _, cm := range maps.Items {
		module := cm.Labels[kube.ModuleLabel]
		if module == "" || (mod != "" && mod != module) {
			continue
		}
		for k, v := range cm.Data {
			ret = append(ret, config.Value{Ref: config.NewRef(optional.Some(module), k), Value: optional.Some([]byte(v))})
		}
	}
	return ret, nil
}

// Load implements Provider.
func (k *KubeConfigMapProvider) Load(ctx context.Context, ref config.Ref) ([]byte, error) {
	module, ok := ref.Module.Get()
	if !ok {
		return nil, config.ErrNotFound
	}
	ns := k.mapper(module, k.realm)
	cm, err := k.client.CoreV1().ConfigMaps(ns).Get(ctx, kube.ConfigMapName(module), v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, config.ErrNotFound
		}
		return nil, errors.Wrap(err, "failed to get ConfigMap")
	}
	val, ok := cm.Data[ref.Name]
	if !ok {
		return nil, config.ErrNotFound
	}
	return []byte(val), nil
}

// Role implements Provider.
func (k *KubeConfigMapProvider) Role() config.Configuration {
	return config.Configuration{}
}

// Store implements Provider.
func (k *KubeConfigMapProvider) Store(ctx context.Context, ref config.Ref, value []byte) error {
	module, ok := ref.Module.Get()
	if !ok {
		return errors.Errorf("unable to set global config")
	}
	conf, cm, err := k.load(ctx, module)
	if err != nil {
		return errors.Wrapf(err, "set %s", ref)
	}
	conf[ref.Name] = value
	if err = k.save(ctx, module, conf, cm); err != nil {
		return errors.Wrapf(err, "set %s", ref)
	}
	return nil
}

func NewKubeConfigProvider(client *kubernetes.Clientset, mapper kube.NamespaceMapper, realm string) *KubeConfigMapProvider {
	return &KubeConfigMapProvider{client: client, mapper: mapper, realm: realm}
}

func (k *KubeConfigMapProvider) load(ctx context.Context, module string) (map[string][]byte, *kubecore.ConfigMap, error) {
	ns := k.mapper(module, k.realm)
	cm, err := k.client.CoreV1().ConfigMaps(ns).Get(ctx, kube.ConfigMapName(module), v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return map[string][]byte{}, nil, nil
		}
		return nil, nil, errors.Wrap(err, "failed to get ConfigMap")
	}
	serialisable := cm.Data
	if cm.Data == nil {
		cm.Data = map[string]string{}
	}

	out := map[string][]byte{}
	for refStr, keyStr := range serialisable {
		out[refStr] = []byte(keyStr)
	}
	// Note that we return the ConfigMap so if it is changed in the background our update will fail
	// This prevents any possible data loss
	return out, cm, nil
}

func (k *KubeConfigMapProvider) save(ctx context.Context, module string, data map[string][]byte, cm *kubecore.ConfigMap) error {
	logger := log.FromContext(ctx)
	serialisable := map[string]string{}
	for ref, key := range data {
		serialisable[ref] = string(key)
	}
	ns := k.mapper(module, k.realm)
	err := kube.EnsureNamespace(ctx, k.client, k.mapper(module, k.realm), k.realm)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespace")
	}
	if cm == nil {
		cm = &kubecore.ConfigMap{}
		kube.AddLabels(&cm.ObjectMeta, k.realm, module)
		cm.Name = kube.ConfigMapName(module)
		cm.Namespace = ns
		logger.Debugf("Creating configmap %s/%s", ns, cm.Name)
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
