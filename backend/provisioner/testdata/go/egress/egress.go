// This is the egress module.
package egress

import (
	"context"

	"github.com/block/ftl/go-runtime/ftl"
)

// TODO: this should be created automatically
type Url = ftl.Config[string]

//ftl:verb export
//ftl:egress url="${url}"
func Egress(ctx context.Context, url ftl.EgressTarget) (string, error) {
	return url.GetString(ctx), nil
}
