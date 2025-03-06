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
		in.WaitWithTimeout("users", time.Minute),
		// Initialize MySQL database
		in.Exec("ftl", "mysql", "new", "users.userdb"),
		
		// Edit initial migration
		in.EditFile("users/db/mysql/userdb/schema/20240101000000_init.sql", func(content []byte) []byte {
			return []byte(`-- migrate:up
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- migrate:down
DROP TABLE users;`)
		}),

		// Add query file
		in.EditFile("users/db/mysql/userdb/queries/users.sql", func(content []byte) []byte {
			return []byte(`-- name: CreateUserRow :exec
INSERT INTO users (name) VALUES (?);

-- name: GetUserRows :many
SELECT id, name FROM users ORDER BY id;`)
		}),

		// Create module code
		in.EditFile("users", func(content []byte) []byte {
			return []byte(`package users

import (
	"context"
)

type User struct {
	ID   int    ` + "`" + `json:"id"` + "`" + `
	Name string ` + "`" + `json:"name"` + "`" + `
}

type CreateUserRequest struct {
	Name string ` + "`" + `json:"name"` + "`" + `
}

//ftl:verb export
func CreateUser(ctx context.Context, req CreateUserRequest, create CreateUserRowClient) error {
	return create(ctx, req.Name)
}

//ftl:verb export
func ListUsers(ctx context.Context, _ struct{}, list GetUserRowsClient) ([]User, error) {
	rows, err := list(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]User, len(rows))
	for i, row := range rows {
		users[i] = User{
			ID:   int(row.ID),
			Name: row.Name,
		}
	}
	return users, nil
}`)
		}, "users.go"),

		// Test initial schema works
		in.Call("users", "createUser", map[string]string{"name": "Alice"}, nil),
		in.Call("users", "createUser", map[string]string{"name": "Bob"}, nil),
		in.Call("users", "listUsers", struct{}{}, func(t testing.TB, response []map[string]interface{}) {
			assert.Equal(t, 2, len(response))
			assert.Equal(t, "Alice", response[0]["name"])
			assert.Equal(t, "Bob", response[1]["name"])
		}),

		// Create new migration to add email column
		in.Exec("ftl", "mysql", "new", "migration", "users.userdb", "add_email"),
		in.EditFile("users/db/mysql/userdb/schema/", func(content []byte) []byte {
			return []byte(`-- migrate:up
ALTER TABLE users ADD COLUMN email VARCHAR(255);

-- migrate:down
ALTER TABLE users DROP COLUMN email;`)
		}, "20240101000001_add_email.sql"),

		// Update queries
		in.EditFile("users/db/mysql/userdb/queries/users.sql", func(content []byte) []byte {
			return []byte(`-- name: CreateUserRow :exec
INSERT INTO users (name, email) VALUES (?, ?);

-- name: GetUserRows :many
SELECT id, name, email FROM users ORDER BY id;`)
		}),

		// Update module code
		in.EditFile("users", func(content []byte) []byte {
			return []byte(`package users

import (
	"context"
)

type User struct {
	ID    int     ` + "`" + `json:"id"` + "`" + `
	Name  string  ` + "`" + `json:"name"` + "`" + `
	Email *string ` + "`" + `json:"email,omitempty"` + "`" + `
}

type CreateUserRequest struct {
	Name  string  ` + "`" + `json:"name"` + "`" + `
	Email *string ` + "`" + `json:"email,omitempty"` + "`" + `
}

//ftl:verb export
func CreateUser(ctx context.Context, req CreateUserRequest, create CreateUserRowClient) error {
	return create(ctx, req.Name, req.Email)
}

//ftl:verb export
func ListUsers(ctx context.Context, _ struct{}, list GetUserRowsClient) ([]User, error) {
	rows, err := list(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]User, len(rows))
	for i, row := range rows {
		users[i] = User{
			ID:    int(row.ID),
			Name:  row.Name,
			Email: row.Email,
		}
	}
	return users, nil
}`)
		}, "users.go"),

		// Test updated schema
		in.Call("users", "createUser", map[string]interface{}{
			"name": "Charlie",
			"email": "charlie@example.com",
		}, nil),
		in.Call("users", "listUsers", struct{}{}, func(t testing.TB, response []map[string]interface{}) {
			assert.Equal(t, 3, len(response))
			assert.Equal(t, "Alice", response[0]["name"])
			assert.Equal(t, nil, response[0]["email"])
			assert.Equal(t, "Bob", response[1]["name"])
			assert.Equal(t, nil, response[1]["email"])
			assert.Equal(t, "Charlie", response[2]["name"])
			assert.Equal(t, "charlie@example.com", response[2]["email"])
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
		in.EditFile("service", func(content []byte) []byte {
			return []byte(`package service

import (
	"context"
)

type GreetRequest struct {
	Name string
}

type GreetResponse struct {
	Message string
}

//ftl:verb export
func Greet(ctx context.Context, req GreetRequest) (GreetResponse, error) {
	return GreetResponse{Message: "Hello, " + req.Name + "!"}, nil
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
			assert.Equal(t, 3, len(schema.Modules)) // builtin + service + client
			for _, m := range schema.Modules {
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
			assert.Equal(t, 3, len(sch.Modules))
			for _, m := range sch.Modules {
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
			assert.Equal(t, 3, len(sch.Modules))
			for _, m := range sch.Modules {
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
			resp = strings.ReplaceAll(resp, "\"Hi, \" + req.Name", "\"Hi, \" + req.Title + \" \" + req.Name")
			return []byte(resp)
		}, "service.go"),

		// Update client to use new parameter
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
			assert.Equal(t, 3, len(sch.Modules))
			for _, m := range sch.Modules {
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
