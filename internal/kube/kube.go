package kube

import (
	"context"
	"os"

	errors "github.com/alecthomas/errors"
	"github.com/block/ftl/common/schema"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

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

func NewNamespaceMapper(userNamespace string) NamespaceMapper {
	if userNamespace != "" {
		return func(module string, realm string) string {
			return userNamespace
		}
	} else {
		return func(module string, realm string) string {
			return module + "-" + realm
		}
	}
}
