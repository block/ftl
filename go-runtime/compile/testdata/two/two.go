//ftl:module two
package two

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

//ftl:enum
type TwoEnum string

const (
	Red   TwoEnum = "Red"
	Blue  TwoEnum = "Blue"
	Green TwoEnum = "Green"
)

type User struct {
	Name string
}

type Payload[T any] struct {
	Body T
}

//ftl:verb
func Two(ctx context.Context, req Payload[string]) (Payload[string], error) {
	return Payload[string]{}, nil
}

//ftl:verb
func CallsTwo(ctx context.Context, req Payload[string]) (Payload[string], error) {
	return ftl.Call(ctx, Two, req)
}
