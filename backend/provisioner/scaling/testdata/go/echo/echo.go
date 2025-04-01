// This is the echo module.
package echo

import (
	"context"
	"fmt"
	"os"

	"github.com/block/ftl/go-runtime/ftl"
)

// Echo returns a greeting with the current time.
//
//ftl:verb export
func Echo(ctx context.Context, req string) (string, error) {
	return fmt.Sprintf("Hello, %s!!!", req), nil
}

//ftl:verb export
func CallerIdentity(ctx context.Context) (string, error) {
	identity := ftl.WorkloadIdentity(ctx)
	id, err := identity.SpiffeID()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func init() {
	os.Getenv("BOGUS")
}
