---
sidebar_position: 1
title: Quick Start
description: One page summary of how to start a new FTL project
---

# Quick Start

One page summary of how to start a new FTL project.

## Requirements

### Install the FTL CLI

Install the FTL CLI on Mac or Linux via [Homebrew](https://brew.sh/), [Hermit](https://cashapp.github.io/hermit), or manually.

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

<Tabs groupId="package-manager">
  <TabItem value="homebrew" label="Homebrew" default>

```bash
brew tap block/ftl && brew install ftl
```

  </TabItem>
  <TabItem value="hermit" label="Hermit">

FTL can be installed from the main Hermit package repository by simply:

```bash
hermit install ftl
```

Alternatively you can add [hermit-ftl](https://github.com/block/hermit-ftl) to your sources by adding the following to your Hermit environment's `bin/hermit.hcl` file:

```hcl
sources = ["https://github.com/block/hermit-ftl.git", "https://github.com/cashapp/hermit-packages.git"]
```

  </TabItem>
  <TabItem value="manual" label="Manually">

Download binaries from the [latest release page](https://github.com/block/ftl/releases/latest) and place them in your `$PATH`.

  </TabItem>
</Tabs>

### Install the VSCode extension

The [FTL VSCode extension](https://marketplace.visualstudio.com/items?itemName=FTL.ftl) provides error and problem reporting through the language server and includes code snippets for common FTL patterns.

## Development

### Initialize an FTL project

Once FTL is installed, initialize an FTL project:

```bash
ftl init myproject
cd myproject
```

This will create a new `myproject` directory containing an `ftl-project.toml` file, a git repository, and a `bin/` directory with Hermit tooling. The Hermit tooling includes the current version of FTL, and language support for go and JVM based languages.

### Create a new module

Now that you have an FTL project, create a new module:

<Tabs groupId="languages">
  <TabItem value="go" label="Go" default>

```bash
ftl module new go alice
```

This will place the code for the new module `alice` in `myproject/alice/alice.go`:

```go
package alice

import (
    "context"
    "fmt"

    "github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

//ftl:verb
func Echo(ctx context.Context, name ftl.Option[string]) (string, error) {
    return fmt.Sprintf("Hello, %s!", name.Default("anonymous")), nil
}
```

Each module is its own Go module.

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

```bash
ftl module new kotlin alice
```

This will create a new Maven `pom.xml` based project in the directory `alice` and create new example code in `alice/src/main/kotlin/ftl/alice/Alice.kt`:

```kotlin
package com.example

import xyz.block.ftl.Export
import xyz.block.ftl.Verb

data class HelloRequest(val name: String)

@Export
@Verb
fun hello(req: HelloRequest): String {
  return "Hello, ${req.name}!"
}
```

  </TabItem>
  <TabItem value="java" label="Java">

```bash
ftl module new java alice
```

This will create a new Maven `pom.xml` based project in the directory `alice` and create new example code in `alice/src/main/java/ftl/alice/Alice.java`:

```java
package com.example;

import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;

public class Alice {

    @Export
    @Verb
    public String hello(String request) {
        return "Hello, " + request + "!";
    }
}
```

  </TabItem>
  <TabItem value="schema" label="Schema">

When you create a new module, FTL generates a schema that represents your code. For the examples above, the schema would look like:

```schema
module alice {
  verb echo(String?) String
}
```

The schema is automatically generated from your code and represents the structure of your FTL module, including data types, verbs, and their relationships.

  </TabItem>
</Tabs>

Any number of modules can be added to your project, adjacent to each other.

### Start the FTL cluster

Start the local FTL development cluster from the command-line:

![ftl dev](/img/quick-start/ftldev.png)

This will build and deploy all local modules. Modifying the code will cause `ftl dev` to rebuild and redeploy the module.

### Open the console

FTL has a console that allows interaction with the cluster topology, logs, traces, and more. Open a browser window at [http://localhost:8899](http://localhost:8899) to view it:

![FTL Console](/img/quick-start/console.png)

### Call your verb

You can call verbs from the console:

![console call](/img/quick-start/consolecall.png)

Or from a terminal use `ftl call` to call your verb:

![ftl call](/img/quick-start/ftlcall.png)

And view your trace in the console:

![console trace](/img/quick-start/consoletrace.png)

### Create another module

Create another module and call `alice.echo` from it with by importing the `alice` module and adding the verb client, `alice.EchoClient`, to the signature of the calling verb. It can be invoked as a function:

<Tabs groupId="languages">
  <TabItem value="go" label="Go" default>

```go
//ftl:verb
import "ftl/alice"

//ftl:verb
func Other(ctx context.Context, in string, ec alice.EchoClient) (string, error) {
    out, err := ec(ctx, in)
    ...
}
```

  </TabItem>
  <TabItem value="kotlin" label="Kotlin">

```kotlin
package com.example

import xyz.block.ftl.Export
import xyz.block.ftl.Verb
import ftl.alice.EchoClient


@Export
@Verb
fun other(req: String, echo: EchoClient): String = "Hello from Other , ${echo.call(req)}!"
```

Note that the `EchoClient` is generated by FTL and must be imported. Unfortunately at the moment JVM based languages have a bit of a chicken-and-egg problem with the generated clients. To force a dependency between the modules you need to add an import on a class that does not exist yet, and then FTL will generate the client for you. This will be fixed in the future.

  </TabItem>
  <TabItem value="java" label="Java">

```java
package com.example.client;

import xyz.block.ftl.Export;
import xyz.block.ftl.Verb;
import ftl.alice.EchoClient;

public class OtherVerb {

    @Export
    @Verb
    public String other(String request, EchoClient echoClient) {
        return "Hello, " + echoClient.call(request) + "!";
    }
}
```

Note that the `EchoClient` is generated by FTL and must be imported. Unfortunately at the moment JVM based languages have a bit of a chicken-and-egg problem with the generated clients. To force a dependency between the modules you need to add an import on a class that does not exist yet, and then FTL will generate the client for you. This will be fixed in the future.

  </TabItem>
  <TabItem value="schema" label="Schema">

When you create a second module that calls the first one, the schema would look like:

```schema
module alice {
  export verb echo(String?) String
}

module other {
  export verb other(String) String
    +calls alice.echo
}
```

The `+calls` annotation in the schema indicates that the `other` verb calls the `echo` verb from the `alice` module.

  </TabItem>
</Tabs>
