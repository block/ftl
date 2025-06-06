package key

import (
	"strconv"

	"github.com/alecthomas/errors"
)

type Runner = KeyType[RunnerPayload, *RunnerPayload]

func NewRunnerKey(hostname, port string) Runner {
	return newKey[RunnerPayload](hostname, port)
}

func NewLocalRunnerKey(suffix int) Runner {
	return newKey[RunnerPayload]("", strconv.Itoa(suffix))
}

func ParseRunnerKey(key string) (Runner, error) {
	return errors.WithStack2(parseKey[RunnerPayload](key))
}

type RunnerPayload struct {
	HostPortMixin
}

var _ KeyPayload = (*RunnerPayload)(nil)

func (r *RunnerPayload) Kind() string { return "rnr" }
