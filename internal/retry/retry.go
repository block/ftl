package retry

import (
	"time"

	"github.com/jpillora/backoff"
)

// RetryConfig is a Kong compatible configuration for creating backoff.Backoff instances.
type RetryConfig struct {
	Min    time.Duration `help:"Minimum retry interval" default:"100ms"`
	Max    time.Duration `help:"Maximum retry interval" default:"10s"`
	Factor float64       `help:"Factor to multiply the retry interval by" default:"2"`
	Jitter bool          `help:"Whether to add jitter to the retry interval" default:"true"`
}

func (c *RetryConfig) Backoff() *backoff.Backoff {
	return &backoff.Backoff{
		Min:    c.Min,
		Max:    c.Max,
		Factor: c.Factor,
		Jitter: c.Jitter,
	}
}
