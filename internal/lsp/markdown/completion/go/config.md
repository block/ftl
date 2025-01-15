Declare a config variable.

Configuration values are named, typed values. They are managed by the `ftl config` command-line.

```go
// Will create a config value called "myConfig" in the FTL schema
type MyConfig = ftl.Config[string]
```

See https://block.github.io/ftl/docs/reference/secretsconfig/
---

type ${1:Name} = ftl.Config[${2:Type}]
