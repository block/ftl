package parent

import (
	"context"

	"ftl/parent/child"
)

//ftl:verb export
func Verb(ctx context.Context) (child.ChildStruct, error) {
	return child.ChildStruct{}, nil
}
