//go:build infrastructure

package deploymentcontext_test

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	in "github.com/block/ftl/internal/integration"
)

func TestKubeDeploymentContext(t *testing.T) {
	in.Run(t,
		in.WithKubernetes(),
		in.VerifyKubeState(func(ctx context.Context, t testing.TB, client kubernetes.Clientset) {
			// setup config / secrets
			secretsCmName := "ftl-module-echo-secrets"
			s := &corev1.Secret{}
			s.Name = secretsCmName
			s.Data = map[string][]byte{
				"apiKey": []byte("\"test-api-key-123\""),
			}
			_, err := client.CoreV1().Secrets("demo").Create(ctx, s, v1.CreateOptions{})
			assert.NoError(t, err, "Failed to create secret %s", secretsCmName)

			configsCmName := "ftl-module-echo-configs"
			cfg := &corev1.ConfigMap{}
			cfg.Name = configsCmName
			cfg.Data = map[string]string{
				"greeting": "\"Bonjour\"",
			}
			_, err = client.CoreV1().ConfigMaps("demo").Create(ctx, cfg, v1.CreateOptions{})
			assert.NoError(t, err, "Failed to create ConfigMap %s", configsCmName)
		}),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.Call("echo", "hello", map[string]string{"name": "Bob"}, func(t testing.TB, response map[string]string) {
			assert.Equal(t, "Bonjour, Bob!!! (authenticated with test-api-key-123)", response["message"])
		}),
		// Now update the secret
		in.VerifyKubeState(func(ctx context.Context, t testing.TB, client kubernetes.Clientset) {

			// Update secrets and configs.
			secretsCmName := "ftl-module-echo-secrets"
			secretsCm, err := client.CoreV1().Secrets("demo").Get(ctx, secretsCmName, v1.GetOptions{})
			assert.NoError(t, err, "Failed to get Secrets %s", secretsCmName)
			secretsCm.Data["echo.apiKey"] = []byte("prod-api-key-123")
			_, err = client.CoreV1().Secrets("demo").Update(ctx, secretsCm, v1.UpdateOptions{})
			assert.NoError(t, err, "Failed to update Secrets %s", secretsCmName)

			configsCmName := "ftl-module-echo-configs"
			configsCm, err := client.CoreV1().ConfigMaps("demo").Get(ctx, configsCmName, v1.GetOptions{})
			assert.NoError(t, err, "Failed to get ConfigMap %s", configsCmName)
			configsCm.Data["echo.greeting"] = "Hola"
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
