Declare a type enum (sum type).

Type enums allow you to define a set of related types that implement a common interface. This provides sum type functionality that Go doesn't natively support.

```go
//ftl:enum
type Animal interface { animal() }

type Cat struct {}
func (Cat) animal() {}

type Dog struct {}
func (Dog) animal() {}
```

See https://block.github.io/ftl/docs/reference/types/
---

//ftl:enum
type ${1:Type} interface { ${2:interface}() }

type ${3:Value} struct {}

func (${3:Value}) ${2:interface}() {}
