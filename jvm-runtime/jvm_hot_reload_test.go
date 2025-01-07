//go:build integration

package ftl_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	in "github.com/block/ftl/internal/integration"
)

func TestLifecycleJVM(t *testing.T) {
	deployment := ""
	in.Run(t,
		in.WithLanguages("java", "kotlin"),
		in.WithDevMode(),
		in.GitInit(),
		in.Exec("rm", "ftl-project.toml"),
		in.Exec("ftl", "init", "test", "."),
		in.IfLanguage("java", in.Exec("ftl", "new", "java", "echo")),
		in.IfLanguage("kotlin", in.Exec("ftl", "new", "kotlin", "echo")),
		in.WaitWithTimeout("echo", time.Minute),
		in.VerifyControllerStatus(func(ctx context.Context, t testing.TB, status *ftlv1.StatusResponse) {
			assert.Equal(t, 1, len(status.Deployments))
			deployment = status.Deployments[0].Key
		}),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Hello, Bob!", response)
		}),
		// Now test hot reload
		in.IfLanguage("java", in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "Hello", "Bye"))
		}, "src/main/java/com/example/EchoVerb.java")),
		in.IfLanguage("kotlin", in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "Hello", "Bye"))
		}, "src/main/kotlin/com/example/EchoVerb.kt")),
		in.Sleep(time.Second*2), // Annoyingly quarkus limits to one restart check every 2s, which is fine for normal dev, but a pain for these tests
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Bye, Bob!", response)
		}),
		in.VerifyControllerStatus(func(ctx context.Context, t testing.TB, status *ftlv1.StatusResponse) {
			// Non structurally changing edits should not trigger a new deployment.
			t.Logf("status %v", status)
			assert.Equal(t, 1, len(status.Deployments))
			assert.Equal(t, deployment, status.Deployments[0].Key)
		}),
		// Structural change should result in a new deployment
		in.IfLanguage("java", in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "@Export", ""))
		}, "src/main/java/com/example/EchoVerb.java")),
		in.IfLanguage("kotlin", in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "@Export", ""))
		}, "src/main/kotlin/com/example/EchoVerb.kt")),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Bye, Bob!", response)
		}),
		in.VerifyControllerStatus(func(ctx context.Context, t testing.TB, status *ftlv1.StatusResponse) {
			assert.Equal(t, 1, len(status.Deployments))
			assert.NotEqual(t, deployment, status.Deployments[0].Key)
		}),

		// Now lets add a database, add the ftl config
		in.EditFile("echo", func(content []byte) []byte {
			return []byte(`
quarkus.datasource.testdb.db-kind=postgresql
quarkus.hibernate-orm.datasource=testdb
`)
		}, "src/main/resources/application.properties"),

		// Create a migration
		in.Exec("ftl", "new-sql-migration", "echo.testdb", "initdb"),

		// Add contents to the migration
		in.EditFiles("echo", func(file string, content []byte) (bool, []byte) {
			if strings.Contains(file, "initdb") {
				return true, []byte(`
-- migrate:up
create sequence StockPrice_SEQ start with 1 increment by 50;

create table StockPrice (
    id bigint not null,
    price float(53) not null,
    stock varchar(255) not null,
    primary key (id)
);
INSERT INTO StockPrice VALUES (0, 100.0, 'FOO');

-- migrate:down

`)
			}
			return false, nil
		}, "src/main/resources/db/"),
		in.Sleep(time.Second*2),
		in.QueryRow("echo_testdb", "SELECT stock from StockPrice", "FOO"),
	)
}
