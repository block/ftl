package deploymentcontext

import (
	"context" //nolint:depguard
	"testing"

	"github.com/alecthomas/assert/v2"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/internal/log"
)

func TestFromSecrets(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	secrets := map[string][]byte{
		"FTL_DSN_ECHO_ECHO": []byte(`"postgres://echo:echo@localhost:5432/echo?sslmode=disable\u0026user=echo\u0026password=echo"
`),
	}
	databases, err := DatabasesFromSecrets(ctx, "echo", secrets, nil)
	assert.NoError(t, err)

	response := NewBuilder("echo").AddDatabases(databases).Build().ToProto()
	assert.Equal(t, &ftlv1.GetDeploymentContextResponse{
		Module:  "echo",
		Configs: map[string][]byte{},
		Secrets: map[string][]byte{},
		Databases: []*ftlv1.GetDeploymentContextResponse_DSN{
			{Name: "echo", Dsn: `postgres://echo:echo@localhost:5432/echo?sslmode=disable&user=echo&password=echo`},
		},
	}, response)
}
