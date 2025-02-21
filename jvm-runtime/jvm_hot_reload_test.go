//go:build integration

package ftl_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/common/schema"
	in "github.com/block/ftl/internal/integration"
)

func TestLifecycleJVM(t *testing.T) {
	deployment := ""
	in.Run(t,
		in.WithLanguages("java"),
		in.WithDevMode(),
		in.GitInit(),
		in.Exec("rm", "ftl-project.toml"),
		in.Exec("ftl", "init", "test", "."),
		in.IfLanguage("java", in.Exec("ftl", "new", "java", "echo")),
		in.IfLanguage("kotlin", in.Exec("ftl", "new", "kotlin", "echo")),
		in.WaitWithTimeout("echo", time.Minute*3),
		in.VerifySchema(func(ctx context.Context, t testing.TB, schema *schema.Schema) {
			assert.Equal(t, 2, len(schema.Modules))
			for _, m := range schema.Modules {
				if !m.Builtin {
					deployment = m.Runtime.Deployment.DeploymentKey.String()
				}
			}
		}),
		in.Call("echo", "hello", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Hello, Bob!", response)
		}),
		// Now test hot reload
		// Deliberate compile error, we need to check that we can recover from this
		in.IfLanguage("java", in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "Hello", "Bye"))
		}, "src/main/java/ftl/echo/Echo.java")),
		in.IfLanguage("kotlin", in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "Hello", "Bye"))
		}, "src/main/kotlin/ftl/echo/Echo.kt")),
		in.Call("echo", "hello", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Bye, Bob!", response)
		}),
		in.VerifySchema(func(ctx context.Context, t testing.TB, sch *schema.Schema) {
			// Non structurally changing edits should not trigger a new deployment.
			assert.Equal(t, 2, len(sch.Modules))
			for _, m := range sch.Modules {
				if !m.Builtin {
					assert.Equal(t, deployment, m.Runtime.Deployment.DeploymentKey.String())
				}
			}
		}),
		//now break compilation
		in.IfLanguage("java", in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "@Export", "broken"))
		}, "src/main/java/ftl/echo/Echo.java")),
		in.IfLanguage("kotlin", in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "@Export", "broken"))
		}, "src/main/kotlin/ftl/echo/Echo.kt")),
		in.Sleep(time.Second*2),
		// Structural change should result in a new deployment
		in.IfLanguage("java", in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "broken", ""))
		}, "src/main/java/ftl/echo/Echo.java")),
		in.IfLanguage("kotlin", in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "broken", ""))
		}, "src/main/kotlin/ftl/echo/Echo.kt")),
		in.Call("echo", "hello", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Bye, Bob!", response)
		}),
		in.VerifySchema(func(ctx context.Context, t testing.TB, sch *schema.Schema) {
			// Non structurally changing edits should not trigger a new deployment.
			assert.Equal(t, 2, len(sch.Modules))
			for _, m := range sch.Modules {
				if !m.Builtin {
					assert.NotEqual(t, deployment, m.Runtime.Deployment.DeploymentKey.String())
				}
			}
		}),
		in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.Replace(string(content), "</parent>", `
</parent>
<dependencies>
	<dependency>
        <groupId>io.quarkus</groupId>
        <artifactId>quarkus-jdbc-postgresql</artifactId>
    </dependency>
</dependencies>`, 1))
		}, "pom.xml"),

		in.Call("echo", "hello", "Bob", func(t testing.TB, response string) {
			assert.Equal(t, "Bye, Bob!", response)
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
