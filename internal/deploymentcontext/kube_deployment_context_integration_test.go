//go:build infrastructure

package deploymentcontext_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	in "github.com/block/ftl/internal/integration"
	"github.com/block/ftl/internal/kube"
)

func TestKubeDeploymentContext(t *testing.T) {
	secretsPath, err := filepath.Abs("testdata/secrets.json")
	assert.NoError(t, err)
	secretsCmName := kube.SecretName("echo")
	configsCmName := kube.ConfigMapName("echo")
	in.Run(t,
		in.WithKubernetes(),
		in.Exec("ftl", "config", "set", "echo.greeting", "Bonjour"),
		in.Exec("ftl", "secret", "import", secretsPath),
		in.VerifyKubeState(func(ctx context.Context, t testing.TB, client kubernetes.Clientset) {
			// setup config / secrets
			s := &corev1.Secret{}
			s.Name = secretsCmName
			s.Data = map[string][]byte{
				"apiKey": []byte("\"test-api-key-123\""),
			}
			_, err := client.CoreV1().Secrets("demo").Create(ctx, s, v1.CreateOptions{})
			assert.NoError(t, err, "Failed to create secret %s", secretsCmName)

		}),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.Call("echo", "hello", map[string]string{"name": "Bob"}, func(t testing.TB, response map[string]string) {
			assert.Equal(t, "Bonjour, Bob!!! (authenticated with test-api-key-123)", response["message"])
		}),
		// Now update the secret
		in.VerifyKubeState(func(ctx context.Context, t testing.TB, client kubernetes.Clientset) {

			// Update secrets and configs.
			secretsCm, err := client.CoreV1().Secrets("demo").Get(ctx, secretsCmName, v1.GetOptions{})
			assert.NoError(t, err, "Failed to get Secrets %s", secretsCmName)
			secretsCm.Data["apiKey"] = []byte("\"prod-api-key-123\"")
			_, err = client.CoreV1().Secrets("demo").Update(ctx, secretsCm, v1.UpdateOptions{})
			assert.NoError(t, err, "Failed to update Secrets %s", secretsCmName)

			configsCm, err := client.CoreV1().ConfigMaps("demo").Get(ctx, configsCmName, v1.GetOptions{})
			assert.NoError(t, err, "Failed to get ConfigMap %s", configsCmName)
			configsCm.Data["greeting"] = "\"Hola\""
			_, err = client.CoreV1().ConfigMaps("demo").Update(ctx, configsCm, v1.UpdateOptions{})
			assert.NoError(t, err, "Failed to update ConfigMap %s", configsCmName)

		}),
		func(t testing.TB, ic in.TestContext) {
			// Allow time for the change to propagate and be detected by the watcher.
			// For k8s, this involves ConfigMap update, kubelet syncing to pod, and then disk provider polling.
			time.Sleep(15 * time.Second) // Increased sleep for k8s propagation and watcher delay (10s poll + buffer).
		},

		// Verify the changes are detected by calling the verb again.
		in.Call("echo", "hello", map[string]string{"name": "Bob"}, func(t testing.TB, response map[string]string) {
			assert.Equal(t, "Hola, Bob!!! (authenticated with prod-api-key-123)", response["message"])
		}),
	)
}
