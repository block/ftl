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
	v2 "istio.io/client-go/pkg/clientset/versioned/typed/security/v1"
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
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
)

const adminDeploymentName = "ftl-admin"
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

	client    *kubernetes.Clientset
	namespace string
	// Map of known deployments
	knownDeployments *xsync.MapOf[string, bool]
	istioSecurity    optional.Option[istioclient.Clientset]
}

func NewK8sScaling(disableIstio bool, controllerURL string) scaling.RunnerScaling {
	return &k8sScaling{disableIstio: disableIstio, controller: controllerURL}
}

func (r *k8sScaling) UpdateDeployment(ctx context.Context, deploymentKey string, sch *schema.Module) error {
	logger := log.FromContext(ctx)
	module := sch.Name
	logger = logger.Module(module)
	ctx = log.ContextWithLogger(ctx, logger)
	logger.Debugf("Updating deployment for %s", deploymentKey)
	deploymentClient := r.client.AppsV1().Deployments(r.namespace)
	deployment, err := deploymentClient.Get(ctx, deploymentKey, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s: %w", deploymentKey, err)
	}
	return r.handleExistingDeployment(ctx, deployment)
}

func (r *k8sScaling) StartDeployment(ctx context.Context, deploymentKey string, sch *schema.Module, hasCron bool, hasIngress bool) (url.URL, error) {
	logger := log.FromContext(ctx)
	module := sch.Name
	logger = logger.Module(module)
	ctx = log.ContextWithLogger(ctx, logger)
	logger.Debugf("Creating deployment for %s", deploymentKey)
	deploymentClient := r.client.AppsV1().Deployments(r.namespace)
	deployment, err := deploymentClient.Get(ctx, deploymentKey, v1.GetOptions{})
	deploymentExists := true
	if err != nil {
		if errors.IsNotFound(err) {
			deploymentExists = false
		} else {
			return url.URL{}, fmt.Errorf("failed to get deployment %s: %w", deploymentKey, err)
		}
	}

	r.knownDeployments.Store(deploymentKey, true)
	if deploymentExists {
		// This should never really happen, but if it does we need to handle it
		logger.Debugf("Updating deployment %s", deploymentKey)
		err = r.handleExistingDeployment(ctx, deployment)
		return r.GetEndpointForDeployment(deploymentKey), err

	}
	err = r.handleNewDeployment(ctx, module, deploymentKey, sch, hasCron, hasIngress)
	if err != nil {
		return url.URL{}, err
	}
	err = r.waitForDeploymentReady(ctx, deploymentKey, deployTimeout)
	if err != nil {
		err2 := r.TerminateDeployment(ctx, deploymentKey)
		if err2 != nil {
			logger.Errorf(err2, "Failed to terminate deployment %s after failure", deploymentKey)
		}
		return url.URL{}, err
	}

	endpoint := r.GetEndpointForDeployment(deploymentKey)
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

func (r *k8sScaling) TerminateDeployment(ctx context.Context, deploymentKey string) error {
	logger := log.FromContext(ctx)
	delCtx := log.ContextWithLogger(context.Background(), logger)
	deploymentClient := r.client.AppsV1().Deployments(r.namespace)
	err := deploymentClient.Delete(delCtx, deploymentKey, v1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete deployment %s: %w", deploymentKey, err)
	}
	return nil
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
	r.namespace = namespace
	r.knownDeployments = xsync.NewMapOf[string, bool]()
	r.istioSecurity = optional.Ptr(sec)
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

func (r *k8sScaling) GetEndpointForDeployment(deployment string) url.URL {
	// TODO: hard coded port? It's hard to deal with as we might not have the lease
	// I think requiring this port is fine for now
	return url.URL{Scheme: "http",
		Host: fmt.Sprintf("%s:8892", deployment),
	}
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

	context, exists := config.Contexts[currentContext]
	if !exists {
		return "", fmt.Errorf("context %s not found in kubeconfig", currentContext)
	}

	return context.Namespace, nil
}

func (r *k8sScaling) updateDeployment(ctx context.Context, name string, mod func(deployment *kubeapps.Deployment)) error {
	deploymentClient := r.client.AppsV1().Deployments(r.namespace)
	for range 10 {

		get, err := deploymentClient.Get(ctx, name, v1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get deployment %s: %w", name, err)
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
	deploymentClient := r.client.AppsV1().Deployments(r.namespace)
	thisDeployment, err := deploymentClient.Get(ctx, adminDeploymentName, v1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get deployment %s: %w", adminDeploymentName, err)
	}
	return thisDeployment.Spec.Template.Spec.Containers[0].Image, nil
}

func (r *k8sScaling) handleNewDeployment(ctx context.Context, module string, name string, sch *schema.Module, cron bool, ingress bool) error {
	logger := log.FromContext(ctx)

	cm, err := r.client.CoreV1().ConfigMaps(r.namespace).Get(ctx, configMapName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get configMap %s: %w", configMapName, err)
	}
	deploymentClient := r.client.AppsV1().Deployments(r.namespace)
	adminDeployment, err := deploymentClient.Get(ctx, adminDeploymentName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s: %w", adminDeploymentName, err)
	}
	provisionerDeployment, err := deploymentClient.Get(ctx, provisionerDeploymentName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s: %w", provisionerDeploymentName, err)
	}
	// First create a Service, this will be the root owner of all the other resources
	// Only create if it does not exist already
	servicesClient := r.client.CoreV1().Services(r.namespace)
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
	serviceAccountClient := r.client.CoreV1().ServiceAccounts(r.namespace)
	serviceAccount, err := serviceAccountClient.Get(ctx, module, v1.GetOptions{})
	if err != nil {
		//TODO: implement cleanup for Service Accounts of modules that are completly removed
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
		serviceAccount.Labels[moduleLabel] = module
		_, err = serviceAccountClient.Create(ctx, serviceAccount, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create service account%s: %w", name, err)
		}
		logger.Debugf("Created kube service  account%s", name)
	} else {
		logger.Debugf("Service account %s already exists", name)
	}

	// Sync the istio policy if applicable
	if sec, ok := r.istioSecurity.Get(); ok {
		err = r.syncIstioPolicy(ctx, sec, module, name, service, adminDeployment, provisionerDeployment, sch, cron, ingress)
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
		runnerImage = strings.ReplaceAll(ourImage, "ftl-admin", rawRunnerImage[len(`ftl0/`):])
	} else {
		// Images outside of the ftl0 namespace should use the same tag as the controller
		ourImageComponents := strings.Split(ourImage, ":")
		if len(ourImageComponents) != 2 {
			return fmt.Errorf("expected <name>:<tag> for image name %q", ourImage)
		}
		runnerImage = rawRunnerImage + ":" + ourImageComponents[1]
	}

	deployment.Name = name
	deployment.OwnerReferences = []v1.OwnerReference{{APIVersion: "v1", Kind: "service", Name: name, UID: service.UID}}
	deployment.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", runnerImage, ourVersion)
	deployment.Spec.Selector = &v1.LabelSelector{MatchLabels: map[string]string{"app": name}}
	if deployment.Spec.Template.ObjectMeta.Labels == nil {
		deployment.Spec.Template.ObjectMeta.Labels = map[string]string{}
	}

	deployment.Spec.Template.Spec.ServiceAccountName = module
	changes, err := r.syncDeployment(ctx, thisImage, deployment, 1)

	if err != nil {
		return err
	}
	for _, change := range changes {

		change(deployment)
	}

	addLabels(&deployment.ObjectMeta, module, name)
	addLabels(&deployment.Spec.Template.ObjectMeta, module, name)
	deployment, err = deploymentClient.Create(ctx, deployment, v1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create deployment %s: %w", deployment.Name, err)
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

func (r *k8sScaling) handleExistingDeployment(ctx context.Context, deployment *kubeapps.Deployment) error {

	thisContainerImage, err := r.thisContainerImage(ctx)
	if err != nil {
		return err
	}
	changes, err := r.syncDeployment(ctx, thisContainerImage, deployment, 1)
	if err != nil {
		return err
	}

	// If we have queued changes we apply them here. Changes can fail and need to be retried
	// Which is why they are supplied as a list of functions
	if len(changes) > 0 {
		err = r.updateDeployment(ctx, deployment.Name, func(deployment *kubeapps.Deployment) {
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
	changes = r.updateEnvVar(deployment, "FTL_ENDPOINT", r.controller, changes)
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

func (r *k8sScaling) syncIstioPolicy(ctx context.Context, sec istioclient.Clientset, module string, name string, service *kubecore.Service, adminDeployment *kubeapps.Deployment, provisionerDeployment *kubeapps.Deployment, sch *schema.Module, hasCron bool, hasIngress bool) error {
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

	policiesClient := sec.SecurityV1().AuthorizationPolicies(r.namespace)

	// Allow controller ingress
	err := r.createOrUpdateIstioPolicy(ctx, policiesClient, name, func(policy *istiosec.AuthorizationPolicy) {
		policy.Name = name
		policy.Namespace = r.namespace
		addLabels(&policy.ObjectMeta, module, name)
		policy.OwnerReferences = []v1.OwnerReference{{APIVersion: "v1", Kind: "service", Name: name, UID: service.UID}}
		// At present we only allow ingress from the controller
		policy.Spec.Selector = &v1beta1.WorkloadSelector{MatchLabels: map[string]string{"app": name}}
		policy.Spec.Action = istiosecmodel.AuthorizationPolicy_ALLOW
		principals := []string{
			"cluster.local/ns/" + r.namespace + "/sa/" + adminDeployment.Spec.Template.Spec.ServiceAccountName,
			"cluster.local/ns/" + r.namespace + "/sa/" + provisionerDeployment.Spec.Template.Spec.ServiceAccountName,
		}
		// TODO: fix hard coded service account names
		if hasIngress {
			// Allow ingress from the ingress gateway
			principals = append(principals, "cluster.local/ns/"+r.namespace+"/sa/ftl-http-ingress")
		}

		if hasCron {
			// Allow cron invocations
			principals = append(principals, "cluster.local/ns/"+r.namespace+"/sa/ftl-cron")
		}
		principals = append(principals, "cluster.local/ns/"+r.namespace+"/sa/ftl-console")
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
		policyName := module + "-" + callableModule
		err := r.createOrUpdateIstioPolicy(ctx, policiesClient, policyName, func(policy *istiosec.AuthorizationPolicy) {
			if policy.Labels == nil {
				policy.Labels = map[string]string{}
			}
			policy.Labels[moduleLabel] = module
			policy.OwnerReferences = []v1.OwnerReference{{APIVersion: "v1", Kind: "service", Name: name, UID: service.UID}}
			policy.Spec.Selector = &v1beta1.WorkloadSelector{MatchLabels: map[string]string{moduleLabel: callableModule}}
			policy.Spec.Action = istiosecmodel.AuthorizationPolicy_ALLOW
			policy.Spec.Rules = []*istiosecmodel.Rule{
				{
					From: []*istiosecmodel.Rule_From{
						{
							Source: &istiosecmodel.Source{
								Principals: []string{"cluster.local/ns/" + r.namespace + "/sa/" + module},
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

func (r *k8sScaling) createOrUpdateIstioPolicy(ctx context.Context, policiesClient v2.AuthorizationPolicyInterface, name string, controllerIngress func(policy *istiosec.AuthorizationPolicy)) error {
	var update func(policy *istiosec.AuthorizationPolicy) error
	policy, err := policiesClient.Get(ctx, name, v1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get istio policy %s: %w", name, err)
		}
		policy = &istiosec.AuthorizationPolicy{}
		policy.Name = name
		policy.Namespace = r.namespace
		update = func(policy *istiosec.AuthorizationPolicy) error {
			_, err := policiesClient.Create(ctx, policy, v1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create istio policy %s: %w", name, err)
			}
			return nil
		}
	} else {
		update = func(policy *istiosec.AuthorizationPolicy) error {
			_, err := policiesClient.Update(ctx, policy, v1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to update istio policy %s: %w", name, err)
			}
			return nil
		}
	}
	controllerIngress(policy)

	return update(policy)
}

func (r *k8sScaling) waitForDeploymentReady(ctx context.Context, key string, timeout time.Duration) error {
	logger := log.FromContext(ctx)
	deploymentClient := r.client.AppsV1().Deployments(r.namespace)
	podClient := r.client.CoreV1().Pods(r.namespace)
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
				return fmt.Errorf("failed to get deployment %s: %w\n%s", key, err, r.findPodLogs(ctx, key, podClient))
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
