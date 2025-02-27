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
		in.WithLanguages("kotlin"),
		in.WithDevMode(),
		in.GitInit(),
		in.Exec("rm", "ftl-project.toml"),
		in.Exec("ftl", "init", "test", "."),
		in.Exec("ftl", "module", "new", "kotlin", "echo"),
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
		in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "Hello", "Bye"))
		}, "src/main/kotlin/ftl/echo/Echo.kt"),
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
		in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "@Export", "broken"))
		}, "src/main/kotlin/ftl/echo/Echo.kt"),
		in.Sleep(time.Second*2),
		// Structural change should result in a new deployment
		in.EditFile("echo", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "broken", ""))
		}, "src/main/kotlin/ftl/echo/Echo.kt"),
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
		in.Sleep(time.Second*2),
		in.QueryRow("echo_testdb", "SELECT stock from StockPrice", "FOO"),
	)
}

func TestMultiModuleHotReloadJVM(t *testing.T) {
	javaDeployment := ""
	kotlinDeployment := ""

	in.Run(t,
		in.WithLanguages("java"),
		in.WithDevMode(),
		in.GitInit(),
		in.Exec("rm", "ftl-project.toml"),
		in.Exec("ftl", "init", "test", "."),
		// Create Java and Kotlin modules
		in.Exec("ftl", "module", "new", "java", "calculator"),
		in.Exec("ftl", "module", "new", "kotlin", "display"),
		in.WaitWithTimeout("calculator", time.Minute*3),
		in.WaitWithTimeout("display", time.Minute*3),

		// Verify initial schema
		in.VerifySchema(func(ctx context.Context, t testing.TB, schema *schema.Schema) {
			assert.Equal(t, 3, len(schema.Modules)) // Including builtin
			for _, m := range schema.Modules {
				if !m.Builtin {
					if m.Runtime.Base.Language == "java" {
						javaDeployment = m.Runtime.Deployment.DeploymentKey.String()
					} else if m.Runtime.Base.Language == "kotlin" {
						kotlinDeployment = m.Runtime.Deployment.DeploymentKey.String()
					}
				}
			}
		}),

		// Edit Java module - add calculator functionality
		in.EditFile("calculator", func(content []byte) []byte {
			return []byte(`
package ftl.calculator;

import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class Calculator {
    @Export
    public static class CalcRequest {
        public int a;
        public int b;
        public CalcRequest() {}
        public CalcRequest(int a, int b) {
            this.a = a;
            this.b = b;
        }
    }

    @Export
    public static class CalcResponse {
        public int result;
        public CalcResponse() {}
        public CalcResponse(int result) {
            this.result = result;
        }
    }

    @Export
    @Verb
    public CalcResponse add(CalcRequest req) {
        return new CalcResponse(req.a + req.b);
    }
}`)
		}, "src/main/java/ftl/calculator/Calculator.java"),

		// Edit Kotlin module - add display functionality that uses calculator
		in.EditFile("display", func(content []byte) []byte {
			return []byte(`
package ftl.display

import xyz.block.ftl.Export
import xyz.block.ftl.Verb
import ftl.calculator.Calculator.CalcRequest
import ftl.calculator.Calculator.CalcResponse

@Export
data class DisplayRequest(val a: Int, val b: Int)

@Export
data class DisplayResponse(val message: String)

class Display {
    @Export
    @Verb
    fun display(req: DisplayRequest, calculator: Calculator): DisplayResponse {
        val result = calculator.add(CalcRequest(req.a, req.b))
        return DisplayResponse("Result is: ${result.result}")
    }
}`)
		}, "src/main/kotlin/ftl/display/Display.kt"),

		// Test initial functionality
		in.Call("display", "display", map[string]interface{}{"a": 5, "b": 3}, func(t testing.TB, response map[string]interface{}) {
			assert.Equal(t, "Result is: 8", response["message"])
		}),

		// Modify Java module - change addition to multiplication
		in.EditFile("calculator", func(content []byte) []byte {
			return []byte(`
package ftl.calculator;

import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class Calculator {
    @Export
    public static class CalcRequest {
        public int a;
        public int b;
        public CalcRequest() {}
        public CalcRequest(int a, int b) {
            this.a = a;
            this.b = b;
        }
    }

    @Export
    public static class CalcResponse {
        public int result;
        public CalcResponse() {}
        public CalcResponse(int result) {
            this.result = result;
        }
    }

    @Export
    @Verb
    public CalcResponse add(CalcRequest req) {
        return new CalcResponse(req.a * req.b); // Changed to multiplication
    }
}`)
		}, "src/main/java/ftl/calculator/Calculator.java"),

		// Verify Java module change is reflected in Kotlin module
		in.Call("display", "display", map[string]interface{}{"a": 5, "b": 3}, func(t testing.TB, response map[string]interface{}) {
			assert.Equal(t, "Result is: 15", response["message"])
		}),

		// Verify deployments haven't changed (non-structural change)
		in.VerifySchema(func(ctx context.Context, t testing.TB, sch *schema.Schema) {
			for _, m := range sch.Modules {
				if !m.Builtin {
					if m.Runtime.Base.Language == "java" {
						assert.Equal(t, javaDeployment, m.Runtime.Deployment.DeploymentKey.String())
					} else if m.Runtime.Base.Language == "kotlin" {
						assert.Equal(t, kotlinDeployment, m.Runtime.Deployment.DeploymentKey.String())
					}
				}
			}
		}),

		// Break Java module compilation
		in.EditFile("calculator", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "@Export", "broken"))
		}, "src/main/java/ftl/calculator/Calculator.java"),
		in.Sleep(time.Second*2),

		// Fix Java module compilation with structural change
		in.EditFile("calculator", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "broken", "@Export\n    @Deprecated")) // Added @Deprecated
		}, "src/main/java/ftl/calculator/Calculator.java"),

		// Verify functionality still works
		in.Call("display", "display", map[string]interface{}{"a": 5, "b": 3}, func(t testing.TB, response map[string]interface{}) {
			assert.Equal(t, "Result is: 15", response["message"])
		}),

		// Verify Java deployment changed (structural change) but Kotlin didn't
		in.VerifySchema(func(ctx context.Context, t testing.TB, sch *schema.Schema) {
			for _, m := range sch.Modules {
				if !m.Builtin {
					if m.Runtime.Base.Language == "java" {
						assert.NotEqual(t, javaDeployment, m.Runtime.Deployment.DeploymentKey.String())
					} else if m.Runtime.Base.Language == "kotlin" {
						assert.Equal(t, kotlinDeployment, m.Runtime.Deployment.DeploymentKey.String())
					}
				}
			}
		}),

		// Modify Kotlin module - change display format
		in.EditFile("display", func(content []byte) []byte {
			return []byte(`
package ftl.display

import xyz.block.ftl.Export
import xyz.block.ftl.Verb
import ftl.calculator.Calculator.CalcRequest
import ftl.calculator.Calculator.CalcResponse

@Export
data class DisplayRequest(val a: Int, val b: Int)

@Export
data class DisplayResponse(val message: String)

class Display {
    @Export
    @Verb
    fun display(req: DisplayRequest, calculator: Calculator): DisplayResponse {
        val result = calculator.add(CalcRequest(req.a, req.b))
        return DisplayResponse("The calculation result is: ${result.result}") // Changed message format
    }
}`)
		}, "src/main/kotlin/ftl/display/Display.kt"),

		// Verify Kotlin changes work and use Java module correctly
		in.Call("display", "display", map[string]interface{}{"a": 5, "b": 3}, func(t testing.TB, response map[string]interface{}) {
			assert.Equal(t, "The calculation result is: 15", response["message"])
		}),

		// Break Kotlin module compilation
		in.EditFile("display", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "@Export", "broken"))
		}, "src/main/kotlin/ftl/display/Display.kt"),
		in.Sleep(time.Second*2),

		// Fix Kotlin module compilation with structural change
		in.EditFile("display", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "broken", "@Export\n    @Deprecated")) // Added @Deprecated
		}, "src/main/kotlin/ftl/display/Display.kt"),

		// Verify functionality still works
		in.Call("display", "display", map[string]interface{}{"a": 5, "b": 3}, func(t testing.TB, response map[string]interface{}) {
			assert.Equal(t, "The calculation result is: 15", response["message"])
		}),

		// Verify Kotlin deployment changed but Java didn't
		in.VerifySchema(func(ctx context.Context, t testing.TB, sch *schema.Schema) {
			for _, m := range sch.Modules {
				if !m.Builtin {
					if m.Runtime.Base.Language == "java" {
						assert.NotEqual(t, javaDeployment, m.Runtime.Deployment.DeploymentKey.String())
					} else if m.Runtime.Base.Language == "kotlin" {
						assert.NotEqual(t, kotlinDeployment, m.Runtime.Deployment.DeploymentKey.String())
					}
				}
			}
		}),
	)
}
