//go:build infrastructure

package scaling_test

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	in "github.com/block/ftl/internal/integration"
)

func TestKubeScalingUserNamespace(t *testing.T) {
	runTest(t, func(dep string) string {
		return "demo"
	})
}
func TestKubeScalingDeploymentPerNamespace(t *testing.T) {
	runTest(t, func(dep string) string {
		return dep + "-ftl"
	}, "--set", "ftl.provisioner.userNamespace=null")
}

func runTest(t *testing.T, namespace func(dep string) string, helmArgs ...string) {
	//failure := atomic.Value[error]{}
	//done := atomic.Value[bool]{}
	//done.Store(false)
	//routineStopped := sync.WaitGroup{}
	//routineStopped.Add(1)
	echoDeployment := map[string]string{}
	in.Run(t,
		in.WithKubernetes(helmArgs...),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.CopyModule("proxy"),
		in.Deploy("proxy"),
		in.CopyModule("naughty"),
		in.Deploy("naughty"),
		in.CopyModule("types"),
		in.Deploy("types"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Hello, Bob!!!", response)
		}),
		in.Call("proxy", "proxy", "Bob", func(t testing.TB, response string) {
			// Verify peer to peer communication
			assert.Equal(t, "Hello, Bob!!!", response)
		}),
		in.Call("proxy", "proxyCallerIdentity", map[string]string{}, func(t testing.TB, response string) {
			// Verify istio caller identity
			assert.Equal(t, "spiffe://cluster.local/ns/"+namespace("proxy")+"/sa/proxy", response)
		}),
		in.VerifyKubeState(func(ctx context.Context, t testing.TB, client kubernetes.Clientset) {
			deps, err := client.AppsV1().Deployments(namespace("echo")).List(ctx, v1.ListOptions{})
			assert.NoError(t, err)
			typesDeployed := false
			for _, dep := range deps.Items {
				if strings.HasPrefix(dep.Name, "dpl-ftl-echo") {
					echoDeployment["name"] = dep.Name
				} else if strings.HasPrefix(dep.Name, "dpl-ftl-types") {
					typesDeployed = true
				}
				assert.Equal(t, 1, *dep.Spec.Replicas)
			}
			assert.False(t, typesDeployed, "no kube deployment should have been created for types")
			assert.NotEqual(t, "", echoDeployment["name"])
		}),
		in.Exec("ftl", "update", "echo", "-n", "2"),
		in.VerifyKubeState(func(ctx context.Context, t testing.TB, client kubernetes.Clientset) {
			deps, err := client.AppsV1().Deployments(namespace("echo")).List(ctx, v1.ListOptions{})
			assert.NoError(t, err)
			for _, dep := range deps.Items {
				if strings.HasPrefix(dep.Name, "dpl-ftl-echo") {
					assert.Equal(t, 2, *dep.Spec.Replicas)
				}
			}
		}),
		in.Call("naughty", "beNaughty", echoDeployment, func(t testing.TB, response string) {
			// If istio is not present we should be able to ping the echo service directly.
			// Istio should prevent this
			assert.Equal(t, strconv.FormatBool(false), response)
		}),
		//func(t testing.TB, ic in.TestContext) {
		//	// Hit the verb constantly to test rolling updates.
		//	go func() {
		//		defer func() {
		//			if r := recover(); r != nil {
		//				failure.Store(errors.Errorf("panic calling verb: %v at %v", r, time.Now()))
		//			}
		//			routineStopped.Done()
		//		}()
		//		for !done.Load() {
		//			in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
		//				if !strings.Contains(response, "Bob") {
		//					failure.Store(errors.Errorf("unexpected response: %s", response))
		//					return
		//				}
		//			})(t, ic)
		//		}
		//	}()
		//},
		in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "Hello", "Bye"))
		}, "echo.go"),
		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Bye, Bob!!!", response)
		}),
		in.VerifyKubeState(func(ctx context.Context, t testing.TB, client kubernetes.Clientset) {
			deps, err := client.AppsV1().Deployments(namespace("echo")).List(ctx, v1.ListOptions{})
			assert.NoError(t, err)
			for _, dep := range deps.Items {
				if strings.HasPrefix(dep.Name, "dpl-ftl-echo") {
					// We should have inherited the scaling from the previous deployment
					assert.Equal(t, 2, *dep.Spec.Replicas)
				}
			}
		}),
		func(t testing.TB, ic in.TestContext) {
			//err := failure.Load()
			//assert.NoError(t, err)
		},
		in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "Bye", "Bonjour"))
		}, "echo.go"),
		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Bonjour, Bob!!!", response)
		}),
		in.Deploy("echo"),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Bonjour, Bob!!!", response)
		}),
		in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "os.Getenv(\"BOGUS\")", "os.Exit(1)"))
		}, "echo.go"),
		in.DeployBroken("echo"),
		func(t testing.TB, ic in.TestContext) {

			// Disabled until after the refactor

			//t.Logf("Checking for no failure during redeploys")
			//done.Store(true)
			//routineStopped.Wait()
			//err := failure.Load()
			//assert.NoError(t, err)
		},
		in.VerifyKubeState(func(ctx context.Context, t testing.TB, client kubernetes.Clientset) {
			deps, err := client.AppsV1().Deployments(namespace("echo")).List(ctx, v1.ListOptions{})
			assert.NoError(t, err)
			depCount := 0
			for _, dep := range deps.Items {
				if strings.HasPrefix(dep.Name, "dpl-ftl-echo") {
					t.Logf("Found deployment %s", dep.Name)
					depCount++
				}
			}
			assert.Equal(t, 1, depCount, "Expected 1 deployment, found %d", depCount)
		}),
		in.Exec("ftl", "kill", "proxy"),
		in.Exec("ftl", "kill", "echo"),
		in.Sleep(time.Second*6), // The drain delay
		in.VerifyKubeState(func(ctx context.Context, t testing.TB, client kubernetes.Clientset) {
			deps, err := client.AppsV1().Deployments(namespace("echo")).List(ctx, v1.ListOptions{})
			assert.NoError(t, err)
			depCount := 0
			for _, dep := range deps.Items {
				if strings.HasPrefix(dep.Name, "dpl-ftl-echo") {
					t.Logf("Found deployment %s", dep.Name)
					depCount++
				}
			}
			assert.Equal(t, 0, depCount, "Expected 0 deployments, found %d", depCount)
		}),
	)
}
