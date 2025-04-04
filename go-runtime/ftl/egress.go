package ftl

import (
	"context"
	"fmt"

	"github.com/block/ftl/go-runtime/internal"
)

// Secret declares a typed secret for the current module.
type EgressTarget struct {
	Name string
}

func (s EgressTarget) String() string { return fmt.Sprintf("egress \"%s\"", s.Name) }

func (s EgressTarget) GoString() string {
	return "ftl.Egress"
}

// GetString returns the value of the egress from FTL.
func (s EgressTarget) GetString(ctx context.Context) string {
	val, err := internal.FromContext(ctx).GetEgress(ctx, s.Name)
	if err != nil {
		panic(fmt.Errorf("failed to get %s: %w", s, err))
	}
	return val
}
