//go:build infrastructure

package deploymentcontext_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	in "github.com/block/ftl/internal/integration"
)

func TestKubeDeploymentContext(t *testing.T) {
	secretsPath, err := filepath.Abs("testdata/secrets.json")
	assert.NoError(t, err)
	in.Run(t,
		in.WithKubernetes(),
		in.CopyModule("echo"),
		in.Exec("ftl", "config", "set", "echo.greeting", "Bonjour", "--inline"),
		in.Exec("ftl", "secret", "import", secretsPath),
		in.Deploy("echo"),
		in.Call("echo", "hello", map[string]string{"name": "Bob"}, func(t testing.TB, response map[string]string) {
			assert.Equal(t, "Bonjour, Bob!!! (authenticated with test-api-key-123)", response["message"])
		}),
		in.VerifyKubeState(func(ctx context.Context, t testing.TB, client kubernetes.Clientset) {
			pods, err := client.CoreV1().Pods("echo-ftl").List(ctx, v1.ListOptions{
				LabelSelector: "ftl.dev/module=echo",
			})
			assert.NoError(t, err)
			assert.NotEqual(t, 0, len(pods.Items), "No pods found for echo module")

			// Verify pod environment variables and volume mounts for secrets/config.
			for _, pod := range pods.Items {
				for _, container := range pod.Spec.Containers {
					if container.Name == "ftl-runner" {
						foundSecretsPathVar := false
						foundConfigsPathVar := false
						foundSchemaLocationVar := false

						for _, env := range container.Env {
							switch env.Name {
							case "FTL_SECRETS_PATH":
								foundSecretsPathVar = true
								assert.Equal(t, "/etc/ftl/secrets", env.Value, "FTL_SECRETS_PATH in pod")
							case "FTL_CONFIGS_PATH":
								foundConfigsPathVar = true
								assert.Equal(t, "/etc/ftl/configs", env.Value, "FTL_CONFIGS_PATH in pod")
							case "FTL_SCHEMA_LOCATION":
								foundSchemaLocationVar = true
								assert.Equal(t, "/etc/ftl/schema.ftl", env.Value, "FTL_SCHEMA_LOCATION in pod")
							}
						}
						assert.True(t, foundSecretsPathVar, "FTL_SECRETS_PATH env var not found in pod")
						assert.True(t, foundConfigsPathVar, "FTL_CONFIGS_PATH env var not found in pod")
						assert.True(t, foundSchemaLocationVar, "FTL_SCHEMA_LOCATION env var not found in pod")

						foundSecretsMount := false
						foundConfigsMount := false
						foundSchemaMount := false

						for _, mount := range container.VolumeMounts {
							if mount.MountPath == "/etc/ftl/secrets" {
								foundSecretsMount = true
							}
							if mount.MountPath == "/etc/ftl/configs" {
								foundConfigsMount = true
							}
							if mount.MountPath == "/etc/ftl/schema.ftl" && mount.SubPath == "schema.ftl" {
								foundSchemaMount = true
							}
						}
						assert.True(t, foundSecretsMount, "Secrets volume mount not found or path incorrect")
						assert.True(t, foundConfigsMount, "Configs volume mount not found or path incorrect")
						assert.True(t, foundSchemaMount, "Schema volume mount not found or path incorrect")
					}
				}
			}

			// Verify ConfigMaps for secrets and configs.
			secretsCmName := "ftl-module-echo-secrets"
			secretsCm, err := client.CoreV1().ConfigMaps("echo-ftl").Get(ctx, secretsCmName, v1.GetOptions{})
			assert.NoError(t, err, "Failed to get secrets ConfigMap %s", secretsCmName)
			assert.NotZero(t, secretsCm, "%s ConfigMap should exist", secretsCmName)
			assert.Equal(t, "test-api-key-123", secretsCm.Data["echo.apiKey"], "Secret echo.apiKey mismatch in ConfigMap %s", secretsCmName)

			configsCmName := "ftl-module-echo-configs"
			configsCm, err := client.CoreV1().ConfigMaps("echo-ftl").Get(ctx, configsCmName, v1.GetOptions{})
			assert.NoError(t, err, "Failed to get configs ConfigMap %s", configsCmName)
			assert.NotZero(t, configsCm, "%s ConfigMap should exist", configsCmName)
			assert.Equal(t, "Bonjour", configsCm.Data["echo.greeting"], "Config echo.greeting mismatch in ConfigMap %s", configsCmName)
		}),

		// Update the config file.
		in.Exec("ftl", "config", "set", "echo.greeting", "Hola", "--inline"),
		func(t testing.TB, ic in.TestContext) {
			// Allow time for the change to propagate and be detected by the watcher.
			// For k8s, this involves ConfigMap update, kubelet syncing to pod, and then disk provider polling.
			time.Sleep(15 * time.Second) // Increased sleep for k8s propagation and watcher delay (10s poll + buffer).
		},

		// Verify the changes are detected by calling the verb again.
		in.Call("echo", "hello", map[string]string{"name": "Bob"}, func(t testing.TB, response map[string]string) {
			assert.Equal(t, "Hola, Bob!!! (authenticated with test-api-key-123)", response["message"])
		}),
	)
}
