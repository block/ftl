package key

import (
	"strconv"
)

type Controller = KeyType[ControllerPayload, *ControllerPayload]

func NewControllerKey(hostname, port string) Controller {
	return newKey[ControllerPayload](hostname, port)
}

func NewLocalControllerKey(suffix int) Controller {
	return newKey[ControllerPayload]("", strconv.Itoa(suffix))
}

func ParseControllerKey(key string) (Controller, error) { return parseKey[ControllerPayload](key) }

var _ KeyPayload = (*ControllerPayload)(nil)

type ControllerPayload struct {
	HostPortMixin
}

func (c *ControllerPayload) Kind() string { return "ctr" }
