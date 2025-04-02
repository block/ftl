package echo

import (
	"context"

	"github.com/block/ftl/go-runtime/ftl"
)

//ftl:verb export
func Echo(ctx context.Context, req string) (string, error) {
	identity := ftl.WorkloadIdentity(ctx)
	id, err := identity.SpiffeID()
	if err != nil {
		return "", err
	}
	return req + " " + id.String(), nil
}
