package config

import (
	"strings"
)

// ProviderKind is the kind of provider, eg. "asm", "file", etc.
type ProviderKind string

func (p ProviderKind) String() string { return string(p) }

// ProviderKey is the key of a provider, eg. "op:realm:vault".
type ProviderKey string

func NewProviderKey(kind ProviderKind, payload ...string) ProviderKey {
	if len(payload) > 0 {
		return ProviderKey(kind.String() + ":" + strings.Join(payload, ":"))
	}
	return ProviderKey(kind)
}

func (p ProviderKey) String() string { return string(p) }
func (p ProviderKey) Kind() ProviderKind {
	parts := strings.Split(string(p), ":")
	return ProviderKind(parts[0])
}
func (p ProviderKey) Payload() []string {
	return strings.Split(string(p), ":")[1:]
}
