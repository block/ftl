//go:build infrastructure

package deploymentcontext_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	in "github.com/block/ftl/internal/integration"
)

func TestKubeDeploymentContext(t *testing.T) {
	secretsPath, err := filepath.Abs("testdata/secrets.json")
	secrets2Path, err := filepath.Abs("testdata/secrets2.json")
	assert.NoError(t, err)
	in.Run(t,
		in.WithKubernetes(),
		in.Exec("ftl", "config", "set", "echo.greeting", "Bonjour"),
		in.Exec("ftl", "secret", "import", secretsPath),
		in.CopyModule("echo"),
		in.Deploy("echo"),
		in.Call("echo", "hello", map[string]string{"name": "Bob"}, func(t testing.TB, response map[string]string) {
			assert.Equal(t, "Bonjour, Bob!!! (authenticated with test-api-key-123)", response["message"])
		}),
		// Now update the secret
		in.Exec("ftl", "config", "set", "echo.greeting", "Hola"),
		in.Exec("ftl", "secret", "import", secrets2Path),
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
