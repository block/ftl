Declare a value enum.

A value enum is an enumerated set of string or integer values.

```go
//ftl:enum
type Color string

const (
  Red   Color = "red"
  Green Color = "green"
  Blue  Color = "blue"
)
```

See https://block.github.io/ftl/docs/reference/types/
---

//ftl:enum
type ${1:Enum} string

const (
	${2:Value1} ${1:Enum} = "${2:Value1}"
	${3:Value2} ${1:Enum} = "${3:Value2}"
)
