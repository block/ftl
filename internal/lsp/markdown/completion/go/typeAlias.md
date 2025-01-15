Declare a type alias.

A type alias is an alternate name for an existing type. You can declare it with or without the equals sign.

```go
//ftl:typealias
type UserID string

//ftl:typealias
type UserToken = string
```

See https://block.github.io/ftl/docs/reference/types/
---

//ftl:typealias
type ${1:Alias} ${2:Type}
