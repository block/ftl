package k8sscaling

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	"github.com/puzpuzpuz/xsync/v3"
	"golang.org/x/exp/maps"
	istiosecmodel "istio.io/api/security/v1"
	"istio.io/api/type/v1beta1"
	istiosec "istio.io/client-go/pkg/apis/security/v1"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	kubeapps "k8s.io/api/apps/v1"
	kubecore "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	v3 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/provisioner/scaling"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
)

const provisionerDeploymentName = "ftl-provisioner"
const configMapName = "ftl-controller-deployment-config"
const deploymentTemplate = "deploymentTemplate"
const serviceTemplate = "serviceTemplate"
const serviceAccountTemplate = "serviceAccountTemplate"
const moduleLabel = "ftl.dev/module"
const deploymentLabel = "ftl.dev/deployment"
const deployTimeout = time.Minute * 5

var _ scaling.RunnerScaling = &k8sScaling{}

type k8sScaling struct {
	disableIstio bool
	controller   string

	client          *kubernetes.Clientset
	systemNamespace string
	// Map of known deployments
	knownDeployments *xsync.MapOf[string, bool]
	istioSecurity    optional.Option[istioclient.Clientset]
	namespaceMapper  NamespaceMapper
	// A unique per cluster identifier for this FTL instance
	instanceName              string
	cronServiceAccount        string
	adminServiceAccount       string
	consoleServiceAccount     string
	httpIngressServiceAccount string
}

type NamespaceMapper func(module string, systemNamespace string) string

func NewK8sScaling(disableIstio bool, controllerURL string, instanceName string, mapper NamespaceMapper, cronServiceAccount string, adminServiceAccount string, consoleServiceAccount string, httpServiceAccount string) scaling.RunnerScaling {
	return &k8sScaling{disableIstio: disableIstio, controller: controllerURL, instanceName: instanceName, namespaceMapper: mapper, consoleServiceAccount: consoleServiceAccount, cronServiceAccount: cronServiceAccount, adminServiceAccount: adminServiceAccount, httpIngressServiceAccount: httpServiceAccount}
}

func (r *k8sScaling) Start(ctx context.Context) error {
	logger := log.FromContext(ctx).Scope("K8sScaling")
	clientset, err := CreateClientSet()
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	namespace, err := GetCurrentNamespace()
	if err != nil {
		// Nothing we can do here, if we don't have a namespace we have no runners
		return fmt.Errorf("failed to get current namespace: %w", err)
	}

	var sec *istioclient.Clientset
	if !r.disableIstio {
		groups, err := clientset.Discovery().ServerGroups()
		if err != nil {
			return fmt.Errorf("failed to get server groups: %w", err)
		}
		// If istio is present and not explicitly disabled we create the client
		for _, group := range groups.Groups {
			if group.Name == "security.istio.io" {
				sec, err = CreateIstioClientSet()
				if err != nil {
					return fmt.Errorf("failed to create istio clientset: %w", err)
				}
				break
			}
		}
	}

	logger.Debugf("Using namespace %s", namespace)
	r.client = clientset
	r.systemNamespace = namespace
	r.knownDeployments = xsync.NewMapOf[string, bool]()
	r.istioSecurity = optional.Ptr(sec)
	return nil
}

func (r *k8sScaling) UpdateDeployment(ctx context.Context, deploymentKey string, sch *schema.Module, subscriptionProcessor bool) error {
	logger := log.FromContext(ctx)
	module := sch.Name
	logger = logger.Module(module)
	ctx = log.ContextWithLogger(ctx, logger)
	logger.Debugf("Updating deployment for %s", deploymentKey)
	deploymentClient := r.client.AppsV1().Deployments(r.namespaceMapper(sch.Name, r.systemNamespace))
	deployment, err := deploymentClient.Get(ctx, deploymentKey, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s for update: %w", deploymentKey, err)
	}
	return r.handleExistingDeployment(ctx, deployment, sch.Runtime.Scaling.MinReplicas)
}

func (r *k8sScaling) StartDeployment(ctx context.Context, deploymentKey string, sch *schema.Module, hasCron bool, hasIngress bool, subscriptionProcessor bool) (url.URL, error) {
	logger := log.FromContext(ctx)
	module := sch.Name
	logger = logger.Module(module)
	ctx = log.ContextWithLogger(ctx, logger)
	dk, err := key.ParseDeploymentKey(deploymentKey)
	if err != nil {
		return url.URL{}, fmt.Errorf("failed to parse deployment key: %w", err)
	}
	deploymentName := deploymentName(subscriptionProcessor, dk)
	logger.Debugf("Creating deployment for %s", deploymentName)
	namespace, err := r.ensureNamespace(ctx, sch)
	if err != nil {
		return url.URL{}, fmt.Errorf("failed to ensure namespace: %w", err)
	}

	deploymentClient := r.client.AppsV1().Deployments(namespace)
	deployment, err := deploymentClient.Get(ctx, deploymentName, v1.GetOptions{})
	deploymentExists := true
	if err != nil {
		if errors.IsNotFound(err) {
			deploymentExists = false
		} else {
			return url.URL{}, fmt.Errorf("failed to check for existence of deployment %s: %w", deploymentName, err)
		}
	}

	r.knownDeployments.Store(deploymentName, true)
	if deploymentExists {
		logger.Debugf("Updating deployment %s", deploymentName)
		err = r.handleExistingDeployment(ctx, deployment, sch.Runtime.Scaling.MinReplicas)
		return r.GetEndpointForDeployment(dk), err

	}
	err = r.handleNewDeployment(ctx, module, deploymentName, sch, hasCron, hasIngress, subscriptionProcessor)
	if err != nil {
		return url.URL{}, err
	}
	err = r.waitForDeploymentReady(ctx, namespace, deploymentName, deployTimeout)
	if err != nil {
		err2 := r.TerminateDeployment(ctx, deploymentName, subscriptionProcessor)
		if err2 != nil {
			logger.Errorf(err2, "Failed to terminate deployment %s after failure", deploymentName)
		}
		return url.URL{}, err
	}

	endpoint := r.GetEndpointForDeployment(dk)
	client := rpc.Dial(ftlv1connect.NewVerbServiceClient, endpoint.String(), log.Error)
	timeout := time.After(1 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return url.URL{}, fmt.Errorf("context cancelled: %w", ctx.Err())
		case <-timeout:
			return url.URL{}, fmt.Errorf("timed out waiting for runner to be ready")
		case <-time.After(time.Millisecond * 100):
			_, err := client.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
			if err == nil {
				return endpoint, nil
			}
		}
	}
}

func deploymentName(subscriptionProcessor bool, dk key.Deployment) string {
	if subscriptionProcessor {
		return "dpl-sub-" + dk.Payload.Module + "-" + dk.Suffix
	}
	return dk.String()
}

func (r *k8sScaling) TerminateDeployment(ctx context.Context, deploymentKey string, subscriptionProcessor bool) error {
	logger := log.FromContext(ctx)
	delCtx := log.ContextWithLogger(context.Background(), logger)
	dk, err := key.ParseDeploymentKey(deploymentKey)
	if err != nil {
		return fmt.Errorf("failed to parse deployment key %s: %w", deploymentKey, err)
	}
	deploymentName := deploymentName(subscriptionProcessor, dk)
	serviceClient := r.client.CoreV1().Services(r.namespaceMapper(dk.Payload.Module, r.systemNamespace))
	err = serviceClient.Delete(delCtx, deploymentName, v1.DeleteOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete service %s: %w", deploymentName, err)
		}
	}
	return nil
}

func CreateClientSet() (*kubernetes.Clientset, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client set: %w", err)
	}
	return clientset, nil
}

func CreateIstioClientSet() (*istioclient.Clientset, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	clientset, err := istioclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client set: %w", err)
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
			return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
		}
	}
	return config, nil
}

func (r *k8sScaling) GetEndpointForDeployment(deployment key.Deployment) url.URL {

	return url.URL{Scheme: "http",
		Host: fmt.Sprintf("%s.%s:8892", deployment.String(), r.namespaceMapper(deployment.Payload.Module, r.systemNamespace))}

}

func GetCurrentNamespace() (string, error) {
	namespaceFile := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	namespace, err := os.ReadFile(namespaceFile)
	if err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to read namespace file: %w", err)
	} else if err == nil {
		return string(namespace), nil
	}

	// If not running in a cluster, get the namespace from the kubeconfig
	configAccess := clientcmd.NewDefaultPathOptions()
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return "", fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	currentContext := config.CurrentContext
	if currentContext == "" {
		return "", fmt.Errorf("no current context found in kubeconfig")
	}

	c, exists := config.Contexts[currentContext]
	if !exists {
		return "", fmt.Errorf("context %s not found in kubeconfig", currentContext)
	}

	return c.Namespace, nil
}

func (r *k8sScaling) updateDeployment(ctx context.Context, namespace string, name string, mod func(deployment *kubeapps.Deployment)) error {
	deploymentClient := r.client.AppsV1().Deployments(namespace)
	for range 10 {

		get, err := deploymentClient.Get(ctx, name, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get deployment %s to apply update: %w", name, err)
		}
		mod(get)
		_, err = deploymentClient.Update(ctx, get, v1.UpdateOptions{})
		if err != nil {
			if errors.IsConflict(err) {
				time.Sleep(time.Second)
				continue
			}
			return fmt.Errorf("failed to update deployment %s: %w", name, err)
		}
		return nil
	}
	return fmt.Errorf("failed to update deployment %s, 10 clonflicts in a row", name)
}

func (r *k8sScaling) thisContainerImage(ctx context.Context) (string, error) {
	deploymentClient := r.client.AppsV1().Deployments(r.systemNamespace)
	thisDeployment, err := deploymentClient.Get(ctx, provisionerDeploymentName, v1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get admin deployment %s: %w", provisionerDeploymentName, err)
	}
	return thisDeployment.Spec.Template.Spec.Containers[0].Image, nil
}

func (r *k8sScaling) handleNewDeployment(ctx context.Context, module string, name string, sch *schema.Module, cron bool, ingress bool, subscriptionProcessor bool) error {
	logger := log.FromContext(ctx)
	userNamespace := r.namespaceMapper(module, r.systemNamespace)
	cm, err := r.client.CoreV1().ConfigMaps(r.systemNamespace).Get(ctx, configMapName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get configMap %s: %w", configMapName, err)
	}
	systemDeploymentClient := r.client.AppsV1().Deployments(r.systemNamespace)
	userDeploymentClient := r.client.AppsV1().Deployments(userNamespace)
	provisionerDeployment, err := systemDeploymentClient.Get(ctx, provisionerDeploymentName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get this provisioner deployment %s: %w", provisionerDeploymentName, err)
	}
	// First create a Service, this will be the root owner of all the other resources
	// Only create if it does not exist already
	servicesClient := r.client.CoreV1().Services(userNamespace)
	service, err := servicesClient.Get(ctx, name, v1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get service %s: %w", name, err)
		}
		logger.Debugf("Creating new kube service %s", name)
		err = decodeBytesToObject([]byte(cm.Data[serviceTemplate]), service)
		if err != nil {
			return fmt.Errorf("failed to decode service from configMap %s: %w", configMapName, err)
		}
		service.Name = name
		service.Spec.Selector = map[string]string{"app": name}
		addLabels(&service.ObjectMeta, module, name)
		service, err = servicesClient.Create(ctx, service, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create service %s: %w", name, err)
		}
		logger.Debugf("Created kube service %s", name)
	} else {
		logger.Debugf("Service %s already exists", name)
	}

	// Now create a ServiceAccount, we mostly need this for Istio but we create it for all deployments
	// To keep things consistent
	serviceAccountClient := r.client.CoreV1().ServiceAccounts(userNamespace)
	serviceAccount, err := serviceAccountClient.Get(ctx, module, v1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get service account %s: %w", name, err)
		}
		logger.Debugf("Creating new kube service account %s", name)
		err = decodeBytesToObject([]byte(cm.Data[serviceAccountTemplate]), serviceAccount)
		if err != nil {
			return fmt.Errorf("failed to decode service account from configMap %s: %w", configMapName, err)
		}
		serviceAccount.Name = module
		if serviceAccount.Labels == nil {
			serviceAccount.Labels = map[string]string{}
		}
		serviceAccount.OwnerReferences = []v1.OwnerReference{{APIVersion: "v1", Kind: "service", Name: name, UID: service.UID}}
		serviceAccount.Labels[moduleLabel] = module
		_, err = serviceAccountClient.Create(ctx, serviceAccount, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create service account%s: %w", name, err)
		}
		logger.Debugf("Created kube service  account%s", name)
	} else {
		logger.Debugf("Service account %s already exists", name)
		serviceAccount.OwnerReferences = append(serviceAccount.OwnerReferences, v1.OwnerReference{APIVersion: "v1", Kind: "service", Name: name, UID: service.UID})
		_, err = serviceAccountClient.Update(ctx, serviceAccount, v1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update service account %s: %w", name, err)
		}
	}

	// Sync the istio policy if applicable
	if sec, ok := r.istioSecurity.Get(); ok {
		err = r.syncIstioPolicy(ctx, sec, userNamespace, module, name, service, provisionerDeployment, sch, cron, ingress)
		if err != nil {
			return err
		}
	}

	// Now create the deployment

	logger.Debugf("Creating new kube deployment %s", name)
	thisImage, err := r.thisContainerImage(ctx)
	if err != nil {
		return fmt.Errorf("failed to get container image: %w", err)
	}
	data := cm.Data[deploymentTemplate]
	deployment := &kubeapps.Deployment{}
	err = decodeBytesToObject([]byte(data), deployment)
	if err != nil {
		return fmt.Errorf("failed to decode deployment from configMap %s: %w", configMapName, err)
	}
	ourVersion, err := extractTag(thisImage)
	if err != nil {
		return err
	}
	ourImage, err := extractBase(thisImage)
	if err != nil {
		return err
	}

	// runner images use the same tag as the controller
	rawRunnerImage := sch.Runtime.Base.Image
	if rawRunnerImage == "" {
		rawRunnerImage = "ftl0/ftl-runner"
	}
	var runnerImage string
	if len(strings.Split(rawRunnerImage, ":")) != 1 {
		return fmt.Errorf("module runtime's image should not contain a tag: %s", rawRunnerImage)
	}
	if strings.HasPrefix(rawRunnerImage, "ftl0/") {
		// Images in the ftl0 namespace should use the same tag as the controller and use the same namespace as ourImage
		runnerImage = strings.ReplaceAll(ourImage, "ftl-provisioner", rawRunnerImage[len(`ftl0/`):])
	} else {
		// Images outside of the ftl0 namespace should use the same tag as the controller
		ourImageComponents := strings.Split(ourImage, ":")
		if len(ourImageComponents) != 2 {
			return fmt.Errorf("expected <name>:<tag> for image name %q", ourImage)
		}
		runnerImage = rawRunnerImage + ":" + ourImageComponents[1]
	}

	deployment.Name = name
	deployment.Namespace = userNamespace
	deployment.OwnerReferences = []v1.OwnerReference{{APIVersion: "v1", Kind: "service", Name: name, UID: service.UID}}
	deployment.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", runnerImage, ourVersion)
	deployment.Spec.Selector = &v1.LabelSelector{MatchLabels: map[string]string{"app": name}}
	if deployment.Spec.Template.ObjectMeta.Labels == nil {
		deployment.Spec.Template.ObjectMeta.Labels = map[string]string{}
	}

	deployment.Spec.Template.Spec.ServiceAccountName = module
	changes, err := r.syncDeployment(ctx, thisImage, deployment, sch.Runtime.Scaling.MinReplicas)

	if err != nil {
		return err
	}
	for _, change := range changes {

		change(deployment)
	}
	envVar := kubecore.EnvVar{Name: "SUBSCRIPTION_PROCESSOR", Value: fmt.Sprintf("%v", subscriptionProcessor)}
	deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, envVar)

	addLabels(&deployment.ObjectMeta, module, name)
	addLabels(&deployment.Spec.Template.ObjectMeta, module, name)
	_, err = userDeploymentClient.Create(ctx, deployment, v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create deployment %s: %w", name, err)
	}
	logger.Debugf("Created kube deployment %s", name)

	return nil
}

func addLabels(obj *v1.ObjectMeta, module string, deployment string) {
	if obj.Labels == nil {
		obj.Labels = map[string]string{}
	}
	obj.Labels["app"] = deployment
	obj.Labels[deploymentLabel] = deployment
	obj.Labels[moduleLabel] = module
}

func decodeBytesToObject(bytes []byte, deployment runtime.Object) error {
	decodingScheme := runtime.NewScheme()
	decoderCodecFactory := serializer.NewCodecFactory(decodingScheme)
	decoder := decoderCodecFactory.UniversalDecoder()
	err := runtime.DecodeInto(decoder, bytes, deployment)
	if err != nil {
		return fmt.Errorf("failed to decode deployment: %w", err)
	}
	return nil
}

func (r *k8sScaling) handleExistingDeployment(ctx context.Context, deployment *kubeapps.Deployment, replicas int32) error {

	thisContainerImage, err := r.thisContainerImage(ctx)
	if err != nil {
		return err
	}
	changes, err := r.syncDeployment(ctx, thisContainerImage, deployment, replicas)
	if err != nil {
		return err
	}

	// If we have queued changes we apply them here. Changes can fail and need to be retried
	// Which is why they are supplied as a list of functions
	if len(changes) > 0 {
		err = r.updateDeployment(ctx, deployment.Namespace, deployment.Name, func(deployment *kubeapps.Deployment) {
			for _, change := range changes {
				change(deployment)
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *k8sScaling) syncDeployment(ctx context.Context, thisImage string, deployment *kubeapps.Deployment, replicas int32) ([]func(*kubeapps.Deployment), error) {
	logger := log.FromContext(ctx)
	changes := []func(*kubeapps.Deployment){}
	ourVersion, err := extractTag(thisImage)
	if err != nil {
		return nil, err
	}
	deploymentVersion, err := extractTag(deployment.Spec.Template.Spec.Containers[0].Image)
	if err != nil {
		return nil, err
	}
	if ourVersion != deploymentVersion {
		// This means there has been an FTL upgrade
		// We are assuming the runner and provisioner run the same version
		// If they are different it means the provisioner has been upgraded and we need
		// to upgrade the deployments

		base, err := extractBase(deployment.Spec.Template.Spec.Containers[0].Image)
		if err != nil {
			logger.Errorf(err, "Could not determine base image for FTL deployment")
		} else {
			changes = append(changes, func(deployment *kubeapps.Deployment) {
				deployment.Spec.Template.Spec.Containers[0].Image = base + ":" + ourVersion
			})
		}
	}

	// For now we just make sure the number of replicas match
	if deployment.Spec.Replicas == nil || *deployment.Spec.Replicas != replicas {
		changes = append(changes, func(deployment *kubeapps.Deployment) {
			deployment.Spec.Replicas = &replicas
		})
	}
	changes = r.updateEnvVar(deployment, "FTL_DEPLOYMENT", deployment.Name, changes)
	return changes, nil
}

func (r *k8sScaling) updateEnvVar(deployment *kubeapps.Deployment, envVerName string, envVarValue string, changes []func(*kubeapps.Deployment)) []func(*kubeapps.Deployment) {
	found := false
	for pos, env := range deployment.Spec.Template.Spec.Containers[0].Env {
		if env.Name == envVerName {
			found = true
			if env.Value != envVarValue {
				changes = append(changes, func(deployment *kubeapps.Deployment) {
					deployment.Spec.Template.Spec.Containers[0].Env[pos] = kubecore.EnvVar{
						Name:  envVerName,
						Value: envVarValue,
					}
				})
			}
			break
		}
	}
	if !found {
		changes = append(changes, func(deployment *kubeapps.Deployment) {
			deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, kubecore.EnvVar{
				Name:  envVerName,
				Value: envVarValue,
			})
		})
	}
	return changes
}

func (r *k8sScaling) syncIstioPolicy(ctx context.Context, sec istioclient.Clientset, namespace string, module string, name string, service *kubecore.Service, provisionerDeployment *kubeapps.Deployment, sch *schema.Module, hasCron bool, hasIngress bool) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Creating new istio policy for %s", name)

	var callableModuleNames []string
	callableModules := map[string]bool{}
	for _, decl := range sch.Decls {
		if verb, ok := decl.(*schema.Verb); ok {
			for _, md := range verb.Metadata {
				if calls, ok := md.(*schema.MetadataCalls); ok {
					for _, call := range calls.Calls {
						callableModules[call.Module] = true
					}
				}
			}

		}
	}
	callableModuleNames = maps.Keys(callableModules)
	callableModuleNames = slices.Sort(callableModuleNames)

	// Allow controller ingress
	err := r.createOrUpdateIstioPolicy(ctx, sec, namespace, name, func(policy *istiosec.AuthorizationPolicy) {
		addLabels(&policy.ObjectMeta, module, name)
		policy.OwnerReferences = []v1.OwnerReference{{APIVersion: "v1", Kind: "service", Name: name, UID: service.UID}}
		// At present we only allow ingress from the controller
		policy.Spec.Selector = &v1beta1.WorkloadSelector{MatchLabels: map[string]string{"app": name}}
		policy.Spec.Action = istiosecmodel.AuthorizationPolicy_ALLOW
		principals := []string{
			"cluster.local/ns/" + r.systemNamespace + "/sa/" + provisionerDeployment.Spec.Template.Spec.ServiceAccountName,
			"cluster.local/ns/" + r.systemNamespace + "/sa/" + r.adminServiceAccount,
			"cluster.local/ns/" + r.systemNamespace + "/sa/" + r.consoleServiceAccount,
		}
		// TODO: fix hard coded service account names
		if hasIngress {
			// Allow ingress from the ingress gateway
			principals = append(principals, "cluster.local/ns/"+r.systemNamespace+"/sa/"+r.httpIngressServiceAccount)
		}

		if hasCron {
			// Allow cron invocations
			principals = append(principals, "cluster.local/ns/"+r.systemNamespace+"/sa/"+r.cronServiceAccount)
		}
		policy.Spec.Rules = []*istiosecmodel.Rule{
			{
				From: []*istiosecmodel.Rule_From{
					{
						Source: &istiosecmodel.Source{
							Principals: principals,
						},
					},
				},
			},
		}
	})
	if err != nil {
		return err
	}

	// Setup policies for the modules we call
	// This feels like the wrong way around but given the way the provisioner works there is not much we can do about this at this stage
	for _, callableModule := range callableModuleNames {
		if callableModule == module {
			continue
		}
		logger.Debugf("Processing callable module %s", callableModule)
		policyName := module + "-" + callableModule
		callableModuleNamespace := r.namespaceMapper(callableModule, r.systemNamespace)
		err := r.createOrUpdateIstioPolicy(ctx, sec, callableModuleNamespace, policyName, func(policy *istiosec.AuthorizationPolicy) {

			targetServiceAccount, err := r.client.CoreV1().ServiceAccounts(callableModuleNamespace).Get(ctx, callableModule, v1.GetOptions{})
			if err != nil {
				logger.Errorf(err, "Failed to get service account %s for module %s", callableModule, callableModule)
			} else {
				policy.OwnerReferences = []v1.OwnerReference{{APIVersion: "v1", Kind: "serviceAccount", Name: targetServiceAccount.Namespace, UID: targetServiceAccount.UID}}
			}

			if policy.Labels == nil {
				policy.Labels = map[string]string{}
			}
			policy.Labels[moduleLabel] = module
			policy.Spec.Selector = &v1beta1.WorkloadSelector{MatchLabels: map[string]string{moduleLabel: callableModule}}
			policy.Spec.Action = istiosecmodel.AuthorizationPolicy_ALLOW
			policy.Spec.Rules = []*istiosecmodel.Rule{
				{
					From: []*istiosecmodel.Rule_From{
						{
							Source: &istiosecmodel.Source{
								Principals: []string{"cluster.local/ns/" + r.namespaceMapper(module, r.systemNamespace) + "/sa/" + module},
							},
						},
					},
				},
			}
		})
		if err != nil {
			return err
		}
	}
	return err
}

func (r *k8sScaling) createOrUpdateIstioPolicy(ctx context.Context, sec istioclient.Clientset, namespace string, name string, modify func(policy *istiosec.AuthorizationPolicy)) error {
	logger := log.FromContext(ctx)
	var update func(policy *istiosec.AuthorizationPolicy) error
	policiesClient := sec.SecurityV1().AuthorizationPolicies(namespace)
	policy, err := policiesClient.Get(ctx, name, v1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get istio policy %s: %w", name, err)
		}
		logger.Debugf("Creating Istio policy for %s/%s", namespace, name)
		policy = &istiosec.AuthorizationPolicy{}
		policy.Name = name
		policy.Namespace = namespace
		update = func(policy *istiosec.AuthorizationPolicy) error {
			_, err := policiesClient.Create(ctx, policy, v1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create istio policy %s: %w", name, err)
			}
			return nil
		}
	} else {
		logger.Debugf("Updating Istio policy for %s/%s", name, namespace)
		update = func(policy *istiosec.AuthorizationPolicy) error {
			_, err := policiesClient.Update(ctx, policy, v1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to update istio policy %s: %w", name, err)
			}
			return nil
		}
	}
	modify(policy)

	return update(policy)
}

func (r *k8sScaling) waitForDeploymentReady(ctx context.Context, namespace string, key string, timeout time.Duration) error {
	logger := log.FromContext(ctx)
	deploymentClient := r.client.AppsV1().Deployments(namespace)
	podClient := r.client.CoreV1().Pods(namespace)
	watch, err := deploymentClient.Watch(ctx, v1.ListOptions{LabelSelector: deploymentLabel + "=" + key})
	podWatch, err := podClient.Watch(ctx, v1.ListOptions{LabelSelector: deploymentLabel + "=" + key})
	if err != nil {
		return fmt.Errorf("failed to watch deployment %s: %w", key, err)
	}
	end := time.After(timeout)
	for {
		select {
		case <-end:
			return fmt.Errorf("deployment %s did not become ready in time \n%s", key, r.findPodLogs(ctx, key, podClient))
		case <-watch.ResultChan():
			deployment, err := deploymentClient.Get(ctx, key, v1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get deployment %s to check readiness: %w\n%s", key, err, r.findPodLogs(ctx, key, podClient))
			}
			if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
				logger.Debugf("Deployment %s is ready", key)
				return nil
			}
			for _, condition := range deployment.Status.Conditions {
				if condition.Type == kubeapps.DeploymentReplicaFailure && condition.Status == kubecore.ConditionTrue {
					return fmt.Errorf("deployment %s is in error state: %s \n%s", deployment, condition.Message, r.findPodLogs(ctx, key, podClient))
				}
			}
		case <-podWatch.ResultChan():
			pods, err := podClient.List(ctx, v1.ListOptions{LabelSelector: deploymentLabel + "=" + key})
			if err != nil {
				return fmt.Errorf("failed to get pods for deployment %s: %w", key, err)
			}
			for _, p := range pods.Items {
				if p.Status.Phase == kubecore.PodFailed {
					return fmt.Errorf("pod %s failed: %s", p.Name, p.Status.Message)
				}
				for _, container := range p.Status.ContainerStatuses {
					if container.State.Waiting != nil {
						if container.State.Waiting.Reason == "ImagePullBackOff" {
							return fmt.Errorf("pod %s is in ImagePullBackOff state", p.Name)
						}
						if container.State.Waiting.Reason == "CrashLoopBackOff" {
							logs, err := readPodLogs(ctx, podClient, &p)
							if err != nil {
								return fmt.Errorf("pod %s is in CrashLoopBackOff state and reading logs failed %w", p.Name, err)
							}
							return fmt.Errorf("pod %s is in CrashLoopBackOff state, logs:\n%s", p.Name, logs)
						}
					}
				}
			}
		}
	}
}

func (r *k8sScaling) findPodLogs(ctx context.Context, key string, podClient v3.PodInterface) string {
	pods, err := podClient.List(ctx, v1.ListOptions{LabelSelector: deploymentLabel + "=" + key})
	if err != nil {
		log.FromContext(ctx).Errorf(err, "Failed to read logs for deployment %s", key)
	}
	ret := ""
	for _, p := range pods.Items {
		for _, container := range p.Status.ContainerStatuses {
			logs, err := readPodLogs(ctx, podClient, &p)
			if err != nil {
				log.FromContext(ctx).Errorf(err, "Failed to read logs for pod %s", p.Name)
			}
			ret += fmt.Sprintf("-- pod %s, container: %s\n%s", p.Name, container.Name, logs)
		}
	}
	return ret
}

func (r *k8sScaling) ensureNamespace(ctx context.Context, sch *schema.Module) (string, error) {
	namespace := r.namespaceMapper(sch.Name, r.systemNamespace)
	if namespace == r.systemNamespace {
		return namespace, nil
	}
	ns, err := r.client.CoreV1().Namespaces().Get(ctx, namespace, v1.GetOptions{})
	if err == nil {
		if ns.Labels != nil {
			// We can deploy into non managed namespaces
			// But if they are managed we check that they are managed by this instance
			if ns.Labels["app.kubernetes.io/managed-by"] == "ftl" {
				if part, ok := ns.Labels["app.kubernetes.io/part-of"]; ok {
					if part != r.instanceName {
						return "", fmt.Errorf("namespace %s is managed by a different ftl instance: %s, this instance is %s", namespace, part, r.instanceName)
					}
				}
			}
		}
		return namespace, nil
	}
	if !errors.IsNotFound(err) {
		return "", fmt.Errorf("failed to get namespace %s: %w", namespace, err)
	}
	ns = &kubecore.Namespace{
		Spec: kubecore.NamespaceSpec{},
		ObjectMeta: v1.ObjectMeta{
			Name:   namespace,
			Labels: map[string]string{"app.kubernetes.io/managed-by": "ftl", "app.kubernetes.io/part-of": r.instanceName, "istio-injection": "enabled"},
		},
	}
	_, err = r.client.CoreV1().Namespaces().Create(ctx, ns, v1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to create namespace %s: %w", namespace, err)
	}
	return namespace, nil
}

func readPodLogs(ctx context.Context, client v3.PodInterface, pod *kubecore.Pod) (string, error) {
	logs := ""
	for _, container := range pod.Spec.Containers {
		if container.Name == "istio-proxy" {
			continue
		}
		req := client.GetLogs(pod.Name, &kubecore.PodLogOptions{Container: container.Name, Previous: false})
		podLogs, err := req.Stream(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to read logs for pod %s: %w", pod.Name, err)
		}
		defer func() {
			_ = podLogs.Close()
		}()
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			return "", fmt.Errorf("failed to read logs for pod %s: %w", pod.Name, err)
		}
		logs += buf.String()
	}
	return logs, nil
}

func extractTag(image string) (string, error) {
	idx := strings.LastIndex(image, ":")
	if idx == -1 {
		return "", fmt.Errorf("no tag found in image %s", image)
	}
	ret := image[idx+1:]
	at := strings.LastIndex(ret, "@")
	if at != -1 {
		ret = image[:at]
	}
	return ret, nil
}

func extractBase(image string) (string, error) {
	idx := strings.LastIndex(image, ":")
	if idx == -1 {
		return "", fmt.Errorf("no tag found in image %s", image)
	}
	return image[:idx], nil
}
