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

// TestLifecycleJVM tests the lifecycle of creating a JVM module and editing it
func TestLifecycleJVM(t *testing.T) {
	deployment := ""
	in.Run(t,
		in.WithLanguages("java", "kotlin"),
		in.WithDevMode(),
		in.WithDebugLogging(),
		in.GitInit(),
		in.Exec("rm", "ftl-project.toml"),
		in.Exec("ftl", "init", "test", "."),
		in.IfLanguage("java", in.Exec("ftl", "module", "new", "java", "echo")),
		in.IfLanguage("kotlin", in.Exec("ftl", "module", "new", "kotlin", "echo")),
		in.WaitWithTimeout("echo", time.Minute*3),
		in.VerifySchema(func(ctx context.Context, t testing.TB, schema *schema.Schema) {
			assert.Equal(t, 2, len(schema.InternalModules()))
			for _, m := range schema.InternalModules() {
				if !m.Builtin {
					deployment = m.Runtime.Deployment.DeploymentKey.String()
				}
			}
		}),
		in.Call("echo", "hello", map[string]string{"name": "Bob"}, func(t testing.TB, response map[string]string) {
			assert.Equal(t, "Hello, Bob!", response["message"])
		}),
		// Now test hot reload
		// Deliberate compile error, we need to check that we can recover from this
		in.IfLanguage("java", in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "\"Hello", "\"Bye"))
		}, "src/main/java/ftl/echo/Echo.java")),
		in.IfLanguage("kotlin", in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "\"Hello", "\"Bye"))
		}, "src/main/kotlin/ftl/echo/Echo.kt")),
		in.Call("echo", "hello", map[string]string{"name": "Bob"}, func(t testing.TB, response map[string]string) {
			assert.Equal(t, "Bye, Bob!", response["message"])
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
		in.Call("echo", "hello", map[string]string{"name": "Bob"}, func(t testing.TB, response map[string]string) {
			assert.Equal(t, "Bye, Bob!", response["message"])
		}),
		in.VerifySchema(func(ctx context.Context, t testing.TB, sch *schema.Schema) {
			assert.Equal(t, 2, len(sch.InternalModules()))
			for _, m := range sch.InternalModules() {
				if !m.Builtin {
					assert.NotEqual(t, deployment, m.Runtime.Deployment.DeploymentKey.String())
				}
			}
		}),
		in.Call("echo", "hello", map[string]string{"name": "Bob"}, func(t testing.TB, response map[string]string) {
			assert.Equal(t, "Bye, Bob!", response["message"])
		}),
		// Create a new datasource
		in.Exec("ftl", "postgres", "new", "echo.testdb"),
		// Add contents to the migration
		in.EditFiles("echo", func(file string, content []byte) (bool, []byte) {
			if strings.Contains(file, "init") {
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
		}, "src/main/resources/db/postgres/testdb/schema/"),
		in.CreateFile("echo", "-- name: GetStockPrice :many\nSELECT stock from StockPrice;", "src/main/resources/db/postgres/testdb/queries/queries.sql"),
		in.Sleep(time.Second*6),
		in.Call("echo", "getStockPrice", map[string]string{}, func(t testing.TB, response []string) {
			assert.Equal(t, 1, len(response))
			assert.Equal(t, "FOO", response[0])
		}))
}

func TestMultiModuleJVMHotReload(t *testing.T) {
	in.Run(t,
		in.WithLanguages("kotlin"),
		in.WithDevMode(),
		in.GitInit(),
		in.Exec("rm", "ftl-project.toml"),
		in.Exec("ftl", "init", "test", "."),
		in.Exec("ftl", "module", "new", "kotlin", "echo"),
		in.Exec("ftl", "module", "new", "kotlin", "greeter"),
		in.WaitWithTimeout("echo", time.Minute*3),
		in.WaitWithTimeout("greeter", time.Minute*3),
		in.Call("echo", "hello", map[string]string{"name": "Bob"}, func(t testing.TB, response map[string]string) {
			assert.Equal(t, "Hello, Bob!", response["message"])
		}),
		in.EditNamedFile("echo", "Echo", func(content []byte) []byte {
			return []byte(`
package ftl.echo

import xyz.block.ftl.Export
import xyz.block.ftl.Verb

data class EchoResponse(val message: String)

@Export
@Verb
fun echo(req: String): EchoResponse {
  return EchoResponse(message = "${req}!")
}
`)
		}),
		in.Call("echo", "echo", "Bob", func(t testing.TB, response map[string]string) {
			assert.Equal(t, "Bob!", response["message"])
		}),
		in.EditNamedFile("greeter", "Greeter", func(content []byte) []byte {
			return []byte(`
package ftl.greeter

import ftl.echo.*
import xyz.block.ftl.Export
import xyz.block.ftl.Verb

data class GreetingRequest(val name: String)
data class GreetingResponse(val message: String)

@Export
@Verb
fun greet(req: GreetingRequest, echo: EchoClient): GreetingResponse {
  val response = echo.call(req.name)
  return GreetingResponse(message = "Greetings, ${response.message}")
}
`)
		}),
		in.Call("greeter", "greet", map[string]string{"name": "Bob"}, func(t testing.TB, response map[string]string) {
			assert.Equal(t, "Greetings, Bob!", response["message"])
		}))
}
