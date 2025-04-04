//go:build integration

package goruntime_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/common/schema"
	in "github.com/block/ftl/internal/integration"
)

func TestHotReloadDatabaseGo(t *testing.T) {
	in.Run(t,
		in.WithLanguages("go"),
		in.WithDevMode(),
		in.GitInit(),
		in.Exec("rm", "ftl-project.toml"),
		in.Exec("ftl", "init", "test", "."),
		// Create module with database
		in.Exec("ftl", "module", "new", "go", "users"),
		in.EditGoMod("users"),
		in.WaitWithTimeout("users", time.Minute),
		// Initialize MySQL database
		in.Exec("ftl", "mysql", "new", "users.userdb"),
		in.Sleep(time.Second*5),
		// Edit initial migration
		in.EditFiles("users", func(file string, content []byte) (bool, []byte) {
			in.Infof("users.userdb: %s", file)
			if !strings.Contains(file, "init") {
				return false, nil
			}
			return true, []byte(`-- migrate:up
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- migrate:down
DROP TABLE users;`)
		}, "db/mysql/userdb/schema/"),
		in.Sleep(time.Second*5),

		// Add query file
		in.CreateFile("users", `-- name: CreateUser :exec
INSERT INTO users (name) VALUES (?);
`, "db/mysql/userdb/queries/queries.sql"),
		in.Sleep(time.Second*5),
		// Test initial schema works
		in.Call("users", "createUser", map[string]string{"name": "Alice"}, func(t testing.TB, response map[string]interface{}) {

		}),
		in.Call("users", "createUser", map[string]string{"name": "Bob"}, func(t testing.TB, response map[string]interface{}) {

		}),
		// Edit the query file, verify we keep our state
		in.EditFile("users", func(content []byte) []byte {
			return []byte(`-- name: CreateUser :exec
INSERT INTO users (name) VALUES (?);

-- name: GetUsers :many
SELECT id, name FROM users ORDER BY id;`)
		}, "db/mysql/userdb/queries/queries.sql"),
		in.Sleep(time.Second*5),
		in.Call("users", "getUsers", struct{}{}, func(t testing.TB, response []map[string]interface{}) {
			assert.Equal(t, 2, len(response))
			assert.Equal(t, "Alice", response[0]["name"])
			assert.Equal(t, "Bob", response[1]["name"])
		}),
		in.EditFiles("users", func(file string, content []byte) (bool, []byte) {
			if !strings.Contains(file, "init") {
				return false, nil
			}
			return true, []byte(`-- migrate:up
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
	email VARCHAR(255) NOT NULL
);

-- migrate:down
DROP TABLE users;`)
		}, "db/mysql/userdb/schema/"),

		// Update queries
		in.EditFile("users", func(content []byte) []byte {
			return []byte(`-- name: CreateUserNew :exec
INSERT INTO users (name, email) VALUES (?, ?);

-- name: GetUsers :many
SELECT id, name, email FROM users ORDER BY id;`)
		}, "db/mysql/userdb/queries/queries.sql"),
		in.Sleep(time.Second*5),

		// Test updated schema
		// We change the verb name so we can't accidentally call the old verb before the new one is ready
		in.Call("users", "createUserNew", map[string]interface{}{
			"name":  "Charlie",
			"email": "charlie@example.com",
		}, func(t testing.TB, response map[string]interface{}) {

		}),
		in.Call("users", "getUsers", struct{}{}, func(t testing.TB, response []map[string]interface{}) {
			assert.Equal(t, "Charlie", response[0]["name"])
			assert.Equal(t, "charlie@example.com", response[0]["email"])
		}),
	)
}

func TestHotReloadMultiModuleGo(t *testing.T) {
	var serviceDeployment, clientDeployment string
	in.Run(t,
		in.WithLanguages("go"),
		in.WithDevMode(),
		in.GitInit(),
		in.Exec("rm", "ftl-project.toml"),
		in.Exec("ftl", "init", "test", "."),
		// Create two modules that will call each other
		in.Exec("ftl", "module", "new", "go", "service"),
		in.Exec("ftl", "module", "new", "go", "client"),
		in.WaitWithTimeout("service", time.Minute),
		in.WaitWithTimeout("client", time.Minute),

		// Edit service module to add a verb
		// This also tests fixtures by initializing the greeting in a fixture
		// Fixtures are only applicable in dev mode which is why it is tested here
		in.EditFile("service", func(content []byte) []byte {
			return []byte(`package service

import (
	"context"
)

var greeting string

type GreetRequest struct {
	Name string
}

type GreetResponse struct {
	Message string
}

//ftl:fixture
func Fixture(ctx context.Context)  error {
    greeting = "Hello, "
	return nil
}
//ftl:verb export
func Greet(ctx context.Context, req GreetRequest) (GreetResponse, error) {
	return GreetResponse{Message: greeting + req.Name + "!"}, nil
}
`)
		}, "service.go"),

		// Edit client module to call service
		in.EditFile("client", func(content []byte) []byte {
			return []byte(`package client

import (
	"context"
	"ftl/service"
)

type EchoRequest struct {
	Name string
}

type EchoResponse struct {
	Message string
}

//ftl:verb export
func Echo(ctx context.Context, req EchoRequest, greet service.GreetClient) (EchoResponse, error) {
	resp, err := greet(ctx, service.GreetRequest{Name: req.Name})
	if err != nil {
		return EchoResponse{}, err
	}
	return EchoResponse{Message: resp.Message}, nil
}
`)
		}, "client.go"),

		// Verify initial schema
		in.VerifySchema(func(ctx context.Context, t testing.TB, schema *schema.Schema) {
			assert.Equal(t, 3, len(schema.InternalModules())) // builtin + service + client
			for _, m := range schema.InternalModules() {
				if !m.Builtin {
					if m.Name == "service" {
						serviceDeployment = m.Runtime.Deployment.DeploymentKey.String()
					} else if m.Name == "client" {
						clientDeployment = m.Runtime.Deployment.DeploymentKey.String()
					}
				}
			}
		}),

		// Test initial call
		in.Call("client", "echo", map[string]string{"name": "Bob"}, func(t testing.TB, response map[string]string) {
			assert.Equal(t, "Hello, Bob!", response["message"])
		}),

		// Modify service response
		in.EditFile("service", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "Hello", "Hi"))
		}, "service.go"),

		// Test modified service response propagates through client
		in.Call("client", "echo", map[string]string{"name": "Bob"}, func(t testing.TB, response map[string]string) {
			assert.Equal(t, "Hi, Bob!", response["message"])
		}),

		// Verify deployment change
		in.VerifySchema(func(ctx context.Context, t testing.TB, sch *schema.Schema) {
			assert.Equal(t, 3, len(sch.InternalModules()))
			for _, m := range sch.InternalModules() {
				if !m.Builtin {
					if m.Name == "service" {
						assert.NotEqual(t, serviceDeployment, m.Runtime.Deployment.DeploymentKey.String())
					} else if m.Name == "client" {
						assert.NotEqual(t, clientDeployment, m.Runtime.Deployment.DeploymentKey.String())
					}
				}
			}
		}),

		// Break service compilation
		in.EditFile("service", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "//ftl:verb export", "broken"))
		}, "service.go"),
		in.Sleep(time.Second*2),

		// Fix service compilation with structural change
		in.EditFile("service", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content), "broken", "//ftl:verb export"))
		}, "service.go"),

		// Test service recovers and client still works
		in.Call("client", "echo", map[string]string{"name": "Bob"}, func(t testing.TB, response map[string]string) {
			assert.Equal(t, "Hi, Bob!", response["message"])
		}),

		// Verify new deployment for structural change
		in.VerifySchema(func(ctx context.Context, t testing.TB, sch *schema.Schema) {
			assert.Equal(t, 3, len(sch.InternalModules()))
			for _, m := range sch.InternalModules() {
				if !m.Builtin {
					if m.Name == "service" {
						assert.NotEqual(t, serviceDeployment, m.Runtime.Deployment.DeploymentKey.String())
					}
				}
			}
		}),

		// Add a new parameter to service
		in.EditFile("service", func(content []byte) []byte {
			resp := strings.ReplaceAll(string(content),
				"Name string",
				"Name string\n\tTitle string")
			resp = strings.ReplaceAll(resp, "greeting + req.Name", "\"Hi, \" + req.Title + \" \" + req.Name")
			return []byte(resp)
		}, "service.go"),

		// Update client to use new parametei
		in.EditFile("client", func(content []byte) []byte {
			return []byte(strings.ReplaceAll(string(content),
				"Name: req.Name",
				"Name: req.Name, Title: \"Mr.\""))
		}, "client.go"),

		// Test with structural changes in both modules
		in.Call("client", "echo", map[string]string{"name": "Bob"}, func(t testing.TB, response map[string]string) {
			assert.Equal(t, "Hi, Mr. Bob!", response["message"])
		}),

		// Verify both modules got new deployments
		in.VerifySchema(func(ctx context.Context, t testing.TB, sch *schema.Schema) {
			assert.Equal(t, 3, len(sch.InternalModules()))
			for _, m := range sch.InternalModules() {
				if !m.Builtin {
					if m.Name == "service" {
						assert.NotEqual(t, serviceDeployment, m.Runtime.Deployment.DeploymentKey.String())
					} else if m.Name == "client" {
						assert.NotEqual(t, clientDeployment, m.Runtime.Deployment.DeploymentKey.String())
					}
				}
			}
		}),
	)
}
