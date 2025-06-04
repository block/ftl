package kube

import (
	"context"
	"fmt"
	"os"

	"github.com/alecthomas/errors"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	kubecore "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeConfig struct {
	UserNamespace string `help:"Namespace to use for kube user resources." env:"FTL_USER_NAMESPACE"`
}

const ModuleLabel = "ftl.dev/module"

const RealmLabel = "ftl.dev/realm"

type NamespaceMapper func(module string, realm string) string

func CreateClientSet() (*kubernetes.Clientset, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client set")
	}
	return clientset, nil
}

func CreateIstioClientSet() (*istioclient.Clientset, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// creates the clientset
	clientset, err := istioclient.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client set")
	}
	return clientset, nil
}

func getKubeConfig() (*rest.Config, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		// if we're not in a cluster, use the kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get kubeconfig")
		}
	}
	return config, nil
}

func GetCurrentNamespace() (string, error) {
	namespaceFile := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	namespace, err := os.ReadFile(namespaceFile)
	if err != nil && !os.IsNotExist(err) {
		return "", errors.Wrap(err, "failed to read namespace file")
	} else if err == nil {
		return string(namespace), nil
	}

	// If not running in a cluster, get the namespace from the kubeconfig
	configAccess := clientcmd.NewDefaultPathOptions()
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return "", errors.Wrap(err, "failed to get kubeconfig")
	}

	currentContext := config.CurrentContext
	if currentContext == "" {
		return "", errors.Errorf("no current context found in kubeconfig")
	}

	c, exists := config.Contexts[currentContext]
	if !exists {
		return "", errors.Errorf("context %s not found in kubeconfig", currentContext)
	}

	return c.Namespace, nil
}

func (k *KubeConfig) NamespaceMapper() NamespaceMapper {
	if k.UserNamespace != "" {
		return func(module string, realm string) string {
			return k.UserNamespace
		}
	}
	return func(module string, realm string) string {
		return module + "-" + realm
	}
}

func (k *KubeConfig) RouteTemplate() string {
	if k.UserNamespace != "" {
		return fmt.Sprintf("http://${module}.%s:8892", k.UserNamespace)
	}
	return "http://${module}.${module}-${realm}:8892"
}

func ConfigMapName(module string) string {
	return fmt.Sprintf("ftl-module-%s-configs", module)
}

func SecretName(module string) string {
	return fmt.Sprintf("ftl-module-%s-secrets", module)
}

func EnsureNamespace(ctx context.Context, client *kubernetes.Clientset, namespace string, instanceName string) error {
	ns, err := client.CoreV1().Namespaces().Get(ctx, namespace, v1.GetOptions{})
	if err == nil {
		if ns.Labels != nil {
			// We can deploy into non managed namespaces
			// But if they are managed we check that they are managed by this instance
			if ns.Labels["app.kubernetes.io/managed-by"] == "ftl" {
				if part, ok := ns.Labels["app.kubernetes.io/part-of"]; ok {
					if part != instanceName {
						return errors.Errorf("namespace %s is managed by a different ftl instance: %s, this instance is %s", namespace, part, instanceName)
					}
				}
			}
		}
		return nil
	}
	if !k8serrors.IsNotFound(err) {
		return errors.Wrapf(err, "failed to get namespace %s", namespace)
	}
	ns = &kubecore.Namespace{
		Spec: kubecore.NamespaceSpec{},
		ObjectMeta: v1.ObjectMeta{
			Name:   namespace,
			Labels: map[string]string{"app.kubernetes.io/managed-by": "ftl", "app.kubernetes.io/part-of": instanceName, "istio-injection": "enabled"},
		},
	}
	_, err = client.CoreV1().Namespaces().Create(ctx, ns, v1.CreateOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to create namespace %s", namespace)
	}
	return nil
}

func AddLabels(obj *v1.ObjectMeta, realm string, module string) {
	if obj.Labels == nil {
		obj.Labels = map[string]string{}
	}
	obj.Labels["app.kubernetes.io/managed-by"] = "ftl"
	obj.Labels[ModuleLabel] = module
	obj.Labels[RealmLabel] = realm
}
