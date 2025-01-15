Declare a secret.

Secrets are encrypted, named, typed values. They are managed by the `ftl secret` command-line.

```go
// Will create a secret value called "mySecret" in the FTL schema
type MySecret = ftl.Secret[string]
```

See https://block.github.io/ftl/docs/reference/secretsconfig/
---

type ${1:Name} = ftl.Secret[${2:Type}]
