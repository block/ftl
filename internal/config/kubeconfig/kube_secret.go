package kubeconfig

import (
	"context"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	kubecore "k8s.io/api/core/v1"
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
	module, ok := ref.Module.Get()
	if !ok {
		return errors.Errorf("unable to unset global config")
	}
	conf, sec, err := k.load(ctx, module)
	if err != nil {
		return errors.Wrapf(err, "set %s", ref)
	}
	delete(conf, ref.Name)
	if err = k.save(ctx, module, conf, sec); err != nil {
		return errors.Wrapf(err, "set %s", ref)
	}
	return nil
}

// Key implements Provider.
func (k *KubeSecretProvider) Key() config.ProviderKey {
	return config.NewProviderKey(KubeSecretProviderKind)
}

// List implements Provider.
func (k *KubeSecretProvider) List(ctx context.Context, withValues bool, forModule optional.Option[string]) ([]config.Value, error) {
	ns := ""
	mod := ""
	ok := false
	if mod, ok = forModule.Get(); ok {
		ns = k.mapper(mod, k.realm)
	}
	secrets, err := k.client.CoreV1().Secrets(ns).List(ctx, v1.ListOptions{LabelSelector: kube.RealmLabel + "=" + k.realm})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get Secrets")
	}
	ret := []config.Value{}
	for _, secret := range secrets.Items {
		module := secret.Labels[kube.ModuleLabel]
		if module == "" || (mod != "" && mod != module) {
			continue
		}
		for k, v := range secret.Data {
			ret = append(ret, config.Value{Ref: config.NewRef(optional.Some(module), k), Value: optional.Some(v)})
		}
	}
	return ret, nil
}

// Load implements Provider.
func (k *KubeSecretProvider) Load(ctx context.Context, ref config.Ref) ([]byte, error) {
	module, ok := ref.Module.Get()
	if !ok {
		return nil, config.ErrNotFound
	}
	ns := k.mapper(module, k.realm)
	cm, err := k.client.CoreV1().Secrets(ns).Get(ctx, kube.SecretName(module), v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, config.ErrNotFound
		}
		return nil, errors.Wrap(err, "failed to get Secret")
	}
	val, ok := cm.Data[ref.Name]
	if !ok {
		return nil, config.ErrNotFound
	}
	return val, nil

}

// Role implements Provider.
func (k *KubeSecretProvider) Role() config.Secrets {
	return config.Secrets{}
}

// Store implements Provider.
func (k *KubeSecretProvider) Store(ctx context.Context, ref config.Ref, value []byte) error {
	module, ok := ref.Module.Get()
	if !ok {
		return errors.Errorf("unable to set global config")
	}
	conf, sec, err := k.load(ctx, module)
	if err != nil {
		return errors.Wrapf(err, "set %s", ref)
	}
	conf[ref.Name] = value
	if err = k.save(ctx, module, conf, sec); err != nil {
		return errors.Wrapf(err, "set %s", ref)
	}
	return nil
}

func NewKubeSecretProvider(client *kubernetes.Clientset, mapper kube.NamespaceMapper, realm string) *KubeSecretProvider {
	return &KubeSecretProvider{client: client, mapper: mapper, realm: realm}
}

func (k *KubeSecretProvider) load(ctx context.Context, module string) (map[string][]byte, *kubecore.Secret, error) {
	ns := k.mapper(module, k.realm)
	cm, err := k.client.CoreV1().Secrets(ns).Get(ctx, kube.SecretName(module), v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return map[string][]byte{}, nil, nil
		}
		return nil, nil, errors.Wrap(err, "failed to get Secret")
	}
	serialisable := cm.Data
	if cm.Data == nil {
		cm.Data = map[string][]byte{}
	}

	out := map[string][]byte{}
	for refStr, keyStr := range serialisable {
		out[refStr] = keyStr
	}
	// Note that we return the Secret so if it is changed in the background our update will fail
	// This prevents any possible data loss
	return out, cm, nil
}

func (k *KubeSecretProvider) save(ctx context.Context, module string, data map[string][]byte, secret *kubecore.Secret) error {

	serialisable := map[string][]byte{}
	for ref, val := range data {
		serialisable[ref] = val
	}
	ns := k.mapper(module, k.realm)
	err := kube.EnsureNamespace(ctx, k.client, k.mapper(module, k.realm), k.realm)
	if err != nil {
		return errors.Wrapf(err, "unable to create namespace")
	}
	if secret == nil {
		secret = &kubecore.Secret{}
		kube.AddLabels(&secret.ObjectMeta, k.realm, module)
		secret.Name = kube.SecretName(module)
		secret.Namespace = ns
		secret, err = k.client.CoreV1().Secrets(ns).Create(ctx, secret, v1.CreateOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to create Secret")
		}
	}
	secret.Data = serialisable

	// This may fail if the secret has been updated by another user
	// This is fine, the user can just retry
	_, err = k.client.CoreV1().Secrets(ns).Update(ctx, secret, v1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to update Secret")
	}
	return nil
}
