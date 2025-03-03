package config

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestKey(t *testing.T) {
	k := NewProviderKey(FileProviderKind)
	assert.Equal(t, FileProviderKind, k.Kind())
	k = NewProviderKey(OnePasswordProviderKind, "realm", "vault")
	assert.Equal(t, OnePasswordProviderKind, k.Kind())
	assert.Equal(t, []string{"realm", "vault"}, k.Payload())
}
