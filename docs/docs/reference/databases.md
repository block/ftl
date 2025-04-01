---
sidebar_position: 17
title: Databases
description: Working with databases in FTL
---

# Databases

FTL has support for Postgresql and MySQL databases, including support for automatic provisioning and migrations. 

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs groupId="languages">
  <TabItem value="go" label="Go" default>

Your database is automatically declared by following a specific directory structure for your SQL files. No additional configuration is needed - just create the directory structure and FTL will handle the rest. See the [Creating a New Database](#creating-a-new-database) section for CLI shortcuts.

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

Your database is automatically declared by following a specific directory structure for your SQL files. No additional configuration is needed - just create the directory structure and FTL will handle the rest. See the [Creating a New Database](#creating-a-new-database) section for CLI shortcuts.

  </TabItem>
  <TabItem value="java" label="Java">

Your database is automatically declared by following a specific directory structure for your SQL files. No additional configuration is needed - just create the directory structure and FTL will handle the rest. See the [Creating a New Database](#creating-a-new-database) section for CLI shortcuts.

  </TabItem>
  <TabItem value="schema" label="Schema">

In the FTL schema, databases are represented using the `database` keyword with the engine type and name:

```schema
module example {
  // Database declaration
  database postgres testdb  
    +migration sha256:59b989063b6de57a1b6867e8ad7915109c9b8632616118c6ef23e4439cf17f8e
  
  // Data structures for database operations
  data CreateUserParams {
    name String
    email String
  }
  
  data UserResult {
    id Int +sql column "users"."id"
    name String +sql column "users"."name"
    email String +sql column "users"."email"
  }
  
  // Query that returns a single row
  verb getUser(Int) example.UserResult
    +database calls example.testdb
    +sql query one "SELECT id, name, email FROM users WHERE id = ?"
  
  // Query that returns multiple rows
  verb listUsers(Unit) [example.UserResult]
    +database calls example.testdb
    +sql query many "SELECT id, name, email FROM users ORDER BY name"
  
  // Query that performs an action but doesn't return data
  verb createUser(example.CreateUserParams) Unit
    +database calls example.testdb
    +sql query exec "INSERT INTO users (name, email) VALUES (?, ?)"
  
  // Custom verb that uses a database query
  export verb getUserEmail(Int) String
}
```

The schema representation includes:
1. A `database` declaration with the engine type (`postgres` or `mysql`) and database name
2. The `+migration` annotation with a SHA256 hash of the migration files
3. Data structures with `+sql column` annotations mapping to database columns
4. Verb declarations with `+database calls` and `+sql query` annotations specifying the query type and SQL statement

  </TabItem>
</Tabs>

## Creating a New Database

To create a new database with the required directory structure, you can use the `ftl postgres new` or `ftl mysql new` command. The format of the command is:

```bash
ftl <engine> new <module>.<datasource>
```

Where:
- `<engine>` is either `mysql` or `postgres`
- `<module>.<datasource>` is the qualified name of the datasource (module name can be omitted if in a single module directory)

For example:
```bash
ftl mysql new mymodule.mydb    # Create a MySQL database named "mydb" in module "mymodule"
ftl postgres new mydb          # Create a PostgreSQL database named "mydb" in the current module
```

This command will:
1. Create the appropriate directory structure
2. Create an initial migration file in the `schema` directory

## SQL File Structure

In order to be discoverable by FTL, the SQL files in your project must follow a specific directory structure. FTL supports two database engines, declared via the directory hierarchy as either `mysql` or `postgres`:

<Tabs groupId="languages">
  <TabItem value="go" label="Go" default>

For Go projects, SQL files must be located in:
```
db/
  ├── mysql/           # must be exactly "mysql" or "postgres"
  │   └── mydb/        # database name
  │       ├── schema/  # contains migration files
  │       └── queries/ # contains query files
```

The presence of a `schema` directory under your database name automatically declares the database in FTL.

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

For Kotlin projects, SQL files must be located in:
```
src/main/resources/
  └── db/
      ├── mysql/           # must be exactly "mysql" or "postgres"
      │   └── mydb/        # database name
      │       ├── schema/  # contains migration files
      │       └── queries/ # contains query files
```

The presence of a `schema` directory under your database name automatically declares the database in FTL.

  </TabItem>
  <TabItem value="java" label="Java">

For Java projects, SQL files must be located in:
```
src/main/resources/
  └── db/
      ├── mysql/           # must be exactly "mysql" or "postgres"
      │   └── mydb/        # database name
      │       ├── schema/  # contains migration files
      │       └── queries/ # contains query files
```

The presence of a `schema` directory under your database name automatically declares the database in FTL.

  </TabItem>
</Tabs>

### Schema Directory

The `schema` directory contains all your database migration `.sql` files. These files are used to create and modify your database schema.

### Queries Directory

The `queries` directory contains `.sql` files with any SQL queries you would like generated as FTL verbs for use in your module. These queries must be annotated with [SQLC annotation syntax](https://docs.sqlc.dev/). FTL will automatically lift these queries into the module schema and provide a type-safe client to execute each query.

Find more information in the [Using Generated Query Clients](#using-generated-query-clients) section below.

## Provisioning

FTL includes support for automatically provisioning databases. The actual backing implementation is
extensible, and presently we include support for both local development provisioning using docker,
and cloud formations based provisioning for AWS deployments. When using `ftl dev` a docker container
will automatically be spun up for each datasource that has been defined, and FTL will automatically
handle configuration. The same applies when deploying to an AWS cluster with cloud formations
provisioning setup.

## Migrations

FTL includes support for automatically running migrations on databases. This is provided by [dbmate](https://github.com/amacneil/dbmate). 

To create additional migrations you can use the `ftl postgres new migration` or `ftl mysql new migration` command. The format of the command is `ftl <engine> new migration <module>.<datasource> <migration-name>`.

The module name can be omitted if the current working directory only contains a single module.

E.g. to create a new migration called `init` for the `testdb` datasource in the `mysql` module you would run `ftl mysql new migration mysql.testdb init`.

When the modules are provisioned FTL will automatically run these migrations for you. 

## Connecting with your DB

There are two supported ways to interact with your database in FTL: using the generated database handle to perform raw queries, or using generated query clients.

### Using the Generated Database Handle

<Tabs groupId="languages">
  <TabItem value="go" label="Go" default>

Once you've declared a database, FTL automatically generates a database handle that provides direct access to the underlying connection. You can use this to execute raw SQL queries (where `MydbHandle` is the generated handle type for the `mydb` datasource):

```go
//ftl:verb export
func Query(ctx context.Context, db MydbHandle) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT data FROM requests")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var i string
		if err := rows.Scan(&i); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
```

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

	TBD

  </TabItem>
  <TabItem value="java" label="Java">
	
	TBD

  </TabItem>
  <TabItem value="schema" label="Schema">

In the FTL schema, the database handle is represented by the `+database calls` annotation on verbs:

```schema
module example {
  // Database declaration
  database postgres mydb
    +migration sha256:59b989063b6de57a1b6867e8ad7915109c9b8632616118c6ef23e4439cf17f8e
  
  // Verb that uses the database handle directly
  export verb query(Unit) [String]
    +database calls example.mydb
}
```

When you use a database handle in your code, you're directly accessing the underlying database connection. The FTL compiler automatically generates the appropriate handle type based on the database declaration.

  </TabItem>
</Tabs>

### Using Generated Query Clients

For better type safety and maintainability, FTL can automatically generate type-safe query clients from SQL files in your `queries` directory. Your SQL files must be annotated with [SQLC annotation syntax](https://docs.sqlc.dev/) to specify the type of query and its parameters. For example:

```sql
-- name: GetUser :one
SELECT id, name, email
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, name, email
FROM users
ORDER BY name;

-- name: CreateUser :exec
INSERT INTO users (name, email)
VALUES ($1, $2);
```

These queries will be automatically converted into FTL verbs with corresponding generated clients that you can inject into your verbs just like any other verb client. For example:

<Tabs groupId="languages">
  <TabItem value="go" label="Go" default>

```go
//ftl:verb export
func GetEmail(ctx context.Context, id int, query GetUserClient) (string, error) {
	result, err := query(ctx, id)
	if err != nil {
		return "", err
	}
	return result.Email, nil
}
```

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

```kotlin
@Verb
fun getEmail(id: Int, query: GetUserClient): String {
    val result = query.getUser(id)
    return result.email
}
```

  </TabItem>
  <TabItem value="java" label="Java">
	
```java
@Verb
public String getEmail(int id, GetUserClient query) {
    UserResult result = query.getUser(id);
    return result.getEmail();
}
```

  </TabItem>
  <TabItem value="schema" label="Schema">

In the FTL schema, the generated query clients are represented as verbs with the `+database calls` and `+sql query` annotations:

```schema
module example {
  // Database declaration
  database postgres testdb
    +migration sha256:59b989063b6de57a1b6867e8ad7915109c9b8632616118c6ef23e4439cf17f8e
  
  // Data structures for query results and parameters
  data UserResult {
    id Int +sql column "users"."id"
    name String +sql column "users"."name"
    email String +sql column "users"."email"
  }
  
  data CreateUserParams {
    name String
    email String
  }
  
  // Query that returns a single row
  verb getUser(Int) example.UserResult
    +database calls example.testdb
    +sql query one "SELECT id, name, email FROM users WHERE id = ?"
  
  // Query that returns multiple rows
  verb listUsers(Unit) [example.UserResult]
    +database calls example.testdb
    +sql query many "SELECT id, name, email FROM users ORDER BY name"
  
  // Query that performs an action but doesn't return data
  verb createUser(example.CreateUserParams) Unit
    +database calls example.testdb
    +sql query exec "INSERT INTO users (name, email) VALUES (?, ?)"
  
  // Custom verb that uses the generated query client
  export verb getUserEmail(Int) String
    +calls example.getUser
}
```

When you use a generated query client in your code, you're calling a verb that has been automatically generated from your SQL query. FTL handles the mapping between your SQL queries and the generated verbs.

  </TabItem>
</Tabs>

## Transactions

FTL provides transaction support through verb annotations. When a verb is marked as transactional, FTL will automatically manage the transaction boundaries - beginning the transaction when the verb is called, committing if the verb completes successfully, and rolling back if any errors occur during execution.

<Tabs groupId="languages">
  <TabItem value="go" label="Go" default>

To mark a verb as transactional, use the `//ftl:transaction` comment directive:

```go
type CreateUserRequest struct {
    Name  string
    Email string
}

//ftl:transaction
func CreateUser(ctx context.Context, req CreateUserRequest, db MydbHandle, createAddress CreateAddressClient) error {
    // All database operations within this verb will be executed in a single transaction
    result, err := db.Get(ctx).Exec("INSERT INTO users (name, email) VALUES ($1, $2)", req.Name, req.Email)
    if err != nil {
        return err // Transaction will be rolled back
    }
    
    userId, err := result.LastInsertId()
    if err != nil {
        return err // Transaction will be rolled back
    }
    
    // createAddress is a generated query client from a SQL file in queries/
    err = createAddress(ctx, AddressRequest{UserId: userId})
    if err != nil {
        return err // Transaction will be rolled back
    }
    
    return nil // Transaction will be committed
}
```

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

To mark a verb as transactional, use the `@Transactional` annotation:

```kotlin
data class CreateUserRequest(
    val name: String,
    val email: String
)

@Transactional
fun createUser(
    request: CreateUserRequest,
    createAddress: CreateAddressClient // Generated query client
): Unit {    
    // createAddress is a generated query client from a SQL file in queries/
    createAddress.createAddress(AddressRequest(userId = userId))
    // Transaction commits if we reach here without errors
    // Any errors will cause automatic rollback
}
```

  </TabItem>
  <TabItem value="java" label="Java">

To mark a verb as transactional, use the `@Transactional` annotation:

```java
public class CreateUserRequest {
    private String name;
    private String email;
}

public class UserService {
    @Transactional
    public void createUser(
        CreateUserRequest request,
        CreateAddressClient createAddress  // Generated query client
    ) { 
        // createAddress is a generated query client from a SQL file in queries/
        createAddress.createAddress(new AddressRequest(userId));
        // Transaction commits if we reach here without errors
        // Any errors will cause automatic rollback
    }
}
```

  </TabItem>
  <TabItem value="schema" label="Schema">

In the FTL schema, transaction verbs are represented with the `+transaction` annotation:

```schema
module example {
  // Database declaration
  database postgres mydb
    +migration sha256:59b989063b6de57a1b6867e8ad7915109c9b8632616118c6ef23e4439cf17f8e
  
  data CreateUserRequest {
    name String
    email String
  }
  
  data AddressRequest {
    userId Int +sql column "addresses"."id"
    street String +sql column "addresses"."street"
    city String +sql column "addresses"."city"
  }
  
  // Generated query client from SQL file
  verb createAddress(example.AddressRequest) Unit
    +database calls example.mydb
    +sql query exec "INSERT INTO addresses (user_id, street, city) VALUES (?, ?, ?)"
  
  // Transaction verb that uses both direct database access and a generated query client
  verb createUser(example.CreateUserRequest) Unit
    +transaction
    +database calls example.mydb
    +calls example.createAddress
}
```

  </TabItem>
</Tabs>

### Transaction Rules and Limitations

When using transactions, FTL enforces the following restrictions:

1. **Database Usage**: Transaction verbs must use at least one database resource (either through a database handle or SQL query clients). FTL infers which database to open the transaction against based on these usages.

2. **Single Database**: FTL only supports transactions against a single database. Two-phase commits across multiple databases are not supported.

3. **No Nesting**: Transaction verbs cannot be nested. This means you cannot inject another transaction verb's client as a resource within a transaction verb.

4. **Module Scope**: Transaction verbs can only call other verbs within the same module. Calls to external verbs from other modules are not allowed within a transaction.

5. **Automatic Management**: Transaction boundaries are automatically managed by FTL:
   - The transaction begins when the verb is called
   - The transaction commits if the verb completes successfully
   - The transaction rolls back if any errors occur during execution
