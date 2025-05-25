package routers

import (
	"context"
	"net/url"

	errors "github.com/alecthomas/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/kube"
)

const (
	FTLConfigmapName = "ftl-deployment"
)

var _ configuration.Router[configuration.Configuration] = (*KubeConfigRouter)(nil)

// KubeConfigRouter is a simple kube based router
type KubeConfigRouter struct {
	client *kubernetes.Clientset
	mapper kube.NamespaceMapper
	realm  string
}

func NewKubeConfigRouter(client *kubernetes.Clientset, mapper kube.NamespaceMapper, realm string) *KubeConfigRouter {
	return &KubeConfigRouter{client: client, mapper: mapper, realm: realm}
}

func (f *KubeConfigRouter) Get(ctx context.Context, ref configuration.Ref) (key *url.URL, err error) {
	if module, ok := ref.Module.Get(); ok {
		ns := f.mapper(module, f.realm)
		cm, err := f.client.CoreV1().ConfigMaps(ns).Get(ctx, FTLConfigmapName, v1.GetOptions{})
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
		u, err := url.Parse(val)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse URL")
		}
		return u, nil
	}
	return nil, nil
}

func (f *KubeConfigRouter) List(ctx context.Context) ([]configuration.Entry, error) {
	// We don't support this currently
	return nil, nil
}

func (f *KubeConfigRouter) Role() (role configuration.Configuration) { return }

func (f *KubeConfigRouter) Set(ctx context.Context, ref configuration.Ref, key *url.URL) error {
	if module, ok := ref.Module.Get(); ok {
		conf, err := f.load(ctx, module)
		if err != nil {
			return errors.Wrapf(err, "set %s", ref)
		}
		conf[ref] = key
		if err = f.save(ctx, module, conf); err != nil {
			return errors.Wrapf(err, "set %s", ref)
		}
		return nil
	}
	return errors.Errorf("unable to set global config")
}

func (f *KubeConfigRouter) Unset(ctx context.Context, ref configuration.Ref) error {
	if module, ok := ref.Module.Get(); ok {
		conf, err := f.load(ctx, module)
		if err != nil {
			return errors.Wrapf(err, "set %s", ref)
		}
		delete(conf, ref)
		if err = f.save(ctx, module, conf); err != nil {
			return errors.Wrapf(err, "set %s", ref)
		}
		return nil
	}
	return errors.Errorf("unable to unset global config")
}

func (f *KubeConfigRouter) load(ctx context.Context, module string) (map[configuration.Ref]*url.URL, error) {
	ns := f.mapper(module, f.realm)
	cm, err := f.client.CoreV1().ConfigMaps(ns).Get(ctx, FTLConfigmapName, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to get ConfigMap")
	}
	serialisable := cm.Data
	if cm.Data == nil {
		return make(map[configuration.Ref]*url.URL), nil
	}

	out := map[configuration.Ref]*url.URL{}
	for refStr, keyStr := range serialisable {
		ref, err := configuration.ParseRef(refStr)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse ref %s", refStr)
		}
		key, err := url.Parse(keyStr)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse key %s", keyStr)
		}
		out[ref] = key
	}
	return out, nil
}

func (f *KubeConfigRouter) save(ctx context.Context, module string, data map[configuration.Ref]*url.URL) error {
	serialisable := map[string]string{}
	for ref, key := range data {
		serialisable[ref.String()] = key.String()
	}
	ns := f.mapper(module, f.realm)
	cm, err := f.client.CoreV1().ConfigMaps(ns).Get(ctx, FTLConfigmapName, v1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to get ConfigMap")
		}
		cm.Name = FTLConfigmapName
		cm.Namespace = ns
		cm.Labels = map[string]string{"app.kubernetes.io/managed-by": "ftl"}
		cm, err = f.client.CoreV1().ConfigMaps(ns).Create(ctx, cm, v1.CreateOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to create ConfigMap")
		}
	}
	cm.Data = serialisable

	cm, err = f.client.CoreV1().ConfigMaps(ns).Update(ctx, cm, v1.UpdateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to update ConfigMap")
	}
	return nil
}
