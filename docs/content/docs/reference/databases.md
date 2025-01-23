+++
title = "Databases"
description = "Databases"
date = 2024-12-27T08:20:00+00:00
updated = 2024-12-27T08:20:00+00:00
draft = false
weight = 60
sort_by = "weight"
template = "docs/page.html"

[extra]
toc = true
top = false
+++

FTL has support for Postgresql and MySQL databases, including support for automatic provisioning and migrations.

The process for declaring a database differs by language.

{% code_selector() %}
<!-- go -->

To use a database in go you must create a struct that implements either the `ftl.MySQLDatabaseConfig` or 
`ftl.PostgresDatabaseConfig` interface. Generally this will involve creating a struct that embeds the
`ftl.DefaultMySQLDatabaseConfig` or `ftl.DefaultPostgresDatabaseConfig` struct and then implementing the `Name() string` method.


You can then use the `ftl.DatabaseHandle` type to access the database by injecting it into an FTL verb. 
An example for MySQL is shown below:

```go
package mysql

import (
	"context"
	"database/sql"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type TestDatasourceConfig struct {
	ftl.DefaultMySQLDatabaseConfig
}

func (TestDatasourceConfig) Name() string { return "testdb" }

//ftl:verb export
func Query(ctx context.Context, db ftl.DatabaseHandle[TestDatasourceConfig]) ([]string, error) {
	var database *sql.DB = db.Get(ctx) // Get the database connection.
	// The following code is standard golang SQL code, it has nothing FTL specific.
	rows, err := database.QueryContext(ctx, "SELECT data FROM requests")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var i string
		if err := rows.Scan(
			&i,
		); err != nil {
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
<!-- kotlin -->
To declare a datasource in Kotlin you must use the `@SQLDatasource` annotation. This annotations is used to define
the database name and type.

```kotlin
@SQLDatasource(name = "testdb", type = SQLDatabaseType.POSTGRESQL)
```

You must also include the appropriate depdencies in your `pom.xml` for the database you are using:

```xml
<dependency>
    <groupId>io.quarkus</groupId>
    <artifactId>quarkus-jdbc-postgresql</artifactId>
</dependency>
<dependency>
    <groupId>io.quarkus</groupId>
    <artifactId>quarkus-jdbc-mysql</artifactId>
</dependency>
```

You can also use [Hibernate directly](https://quarkus.io/guides/hibernate-orm) or using [Panache](https://quarkus.io/guides/hibernate-orm-panache).

This will require adding one of the following dependencies:

```xml
<dependency>
    <groupId>io.quarkus</groupId>
    <artifactId>quarkus-hibernate-orm</artifactId>
</dependency>
<dependency>
    <groupId>io.quarkus</groupId>
    <artifactId>quarkus-hibernate-orm-panache</artifactId>
</dependency>
```

Note that this will likely change significantly in future once FTL has SQL Verbs.

<!-- java -->
To declare a datasource in Java you must use the `@SQLDatasource` annotation. This annotations is used to define
the database name and type.

```java
@SQLDatasource(name = "testdb", type = SQLDatabaseType.POSTGRESQL)
```

You must also include the appropriate depdencies in your `pom.xml` for the database you are using:

```xml
<dependency>
    <groupId>io.quarkus</groupId>
    <artifactId>quarkus-jdbc-postgresql</artifactId>
</dependency>
<dependency>
    <groupId>io.quarkus</groupId>
    <artifactId>quarkus-jdbc-mysql</artifactId>
</dependency>
```

You can also use [Hibernate directly](https://quarkus.io/guides/hibernate-orm) or using [Panache](https://quarkus.io/guides/hibernate-orm-panache).

This will require adding one of the following dependencies:

```xml
<dependency>
    <groupId>io.quarkus</groupId>
    <artifactId>quarkus-hibernate-orm</artifactId>
</dependency>
<dependency>
    <groupId>io.quarkus</groupId>
    <artifactId>quarkus-hibernate-orm-panache</artifactId>
</dependency>
```

Note that this will likely change significantly in future once FTL has SQL Verbs.

```
{% end %}

# Provisioning

FTL includes support for automatically provisioning databases. The actual backing implementation is
extensible, and presently we include support for both local development provisioning using docker,
and cloud formations based provisioning for AWS deployments. When using `ftl dev` a docker container
will automatically be spun up for each datasource that has been defined, and FTL will automatically
handle configuration. The same applies when deploying to an AWS cluster with cloud formations
provisioning setup.

# Migrations

FTL includes support for automatically running migrations on databases. This is provided by [dbmate](https://github.com/amacneil/dbmate). 

To create migrations you can use the `ftl new-sql-migration` command. This will create new migration files, and initialize the required
directory structure if it does not exist. The format of the command is `ftl new-sql-migration <module>.<datasource> <migration-name>`.

The module name can be omitted if the current working directory only contains a single module.

E.g. to create a new migration called `init` for the `testdb` datasource in the `mysql` module you would run `ftl new-sql-migration mysql.testdb init`.

When the modules are provisioned FTL will automatically run these migrations for you.
