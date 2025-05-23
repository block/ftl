package routers

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"sort"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/docker/docker/api/types/mount"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/kube"
)

const (
	FTLConfigmapName = "ftl-deployment"
	FTLSecretName    = "ftl-deployment"
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
	conf, err := f.load()
	if err != nil {
		return nil, errors.Wrap(err, "list")
	}
	out := make([]configuration.Entry, 0, len(conf))
	for ref, key := range conf {
		out = append(out, configuration.Entry{Ref: ref, Accessor: key})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Ref.String() < out[j].Ref.String() })
	return out, nil
}

func (f *KubeConfigRouter) Role() (role configuration.Configuration) { return }

func (f *KubeConfigRouter) Set(ctx context.Context, ref configuration.Ref, key *url.URL) error {
	conf, err := f.load()
	if err != nil {
		return errors.Wrapf(err, "set %s", ref)
	}
	conf[ref] = key
	if err = f.save(conf); err != nil {
		return errors.Wrapf(err, "set %s", ref)
	}
	return nil
}

func (f *KubeConfigRouter) Unset(ctx context.Context, ref configuration.Ref) error {
	conf, err := f.load()
	if err != nil {
		return errors.Wrapf(err, "unset %s", ref)
	}
	delete(conf, ref)
	if err = f.save(conf); err != nil {
		return errors.Wrapf(err, "unset %s", ref)
	}
	return nil
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
	w, err := os.Create(f.path)
	if err != nil {
		return errors.Wrap(err, "failed to create file")
	}
	serialisable := map[string]string{}
	for ref, key := range data {
		serialisable[ref.String()] = key.String()
	}
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
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err = enc.Encode(serialisable); err != nil {
		return errors.Wrapf(err, "failed to encode %s", f.path)
	}
	return nil
}
