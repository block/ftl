//nolint:revive
package key

import (
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"fmt"
	"reflect"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/multiformats/go-base36"
)

// Overridable random source for testing
var randRead = rand.Read

// A constraint that requires itself be a pointer to a T that implements KeyPayload.
//
// This is necessary so that keyType.Payload can be a value rather than a pointer.
type keyPayloadConstraint[T any] interface {
	*T
	KeyPayload
}

// KeyPayload is an interface that all key payloads must implement.
type KeyPayload interface {
	Kind() string
	String() string
	// Parse the hyphen-separated parts of the payload
	Parse(parts []string) error
	// RandomBytes determines the number of random bytes the key should include.
	RandomBytes() int
}

// KeyType is a helper type to avoid having to write a bunch of boilerplate.
type KeyType[T comparable, TP keyPayloadConstraint[T]] struct {
	Payload T
	Suffix  string
}

var _ interface {
	sql.Scanner
	driver.Valuer
	encoding.TextUnmarshaler
	encoding.TextMarshaler
} = (*KeyType[ControllerPayload, *ControllerPayload])(nil)

func (d KeyType[T, TP]) IsZero() bool {
	return d.Equal(KeyType[T, TP]{})
}

func (d KeyType[T, TP]) Equal(other KeyType[T, TP]) bool {
	return reflect.DeepEqual(d, other)
}

func (d KeyType[T, TP]) Value() (driver.Value, error) {
	return d.String(), nil
}

// Scan from DB representation.
func (d *KeyType[T, TP]) Scan(src any) error {
	input, ok := src.(string)
	if !ok {
		return errors.Errorf("expected key to be a string but it's a %T", src)
	}
	if input == "" {
		var zero KeyType[T, TP]
		*d = zero
		return nil
	}
	key, err := parseKey[T, TP](input)
	if err != nil {
		return errors.WithStack(err)
	}
	*d = key
	return nil
}

func (d KeyType[T, TP]) Kind() string {
	var payload TP = &d.Payload
	return payload.Kind()
}

func (d KeyType[T, TP]) GoString() string {
	var t T
	// Assumes the naming convention of:
	// type DeploymentKey = KeyType[DeploymentPayload, *DeploymentPayload]
	return fmt.Sprintf("model.%s(%q)", strings.ReplaceAll(reflect.TypeOf(t).Name(), "Payload", "Key"), d.String())
}

func (d KeyType[T, TP]) String() string {
	if d.IsZero() {
		return ""
	}

	parts := []string{d.Kind()}
	var payload TP = &d.Payload
	if payload := payload.String(); payload != "" {
		parts = append(parts, payload)
	}
	parts = append(parts, d.Suffix)
	return strings.Join(parts, "-")
}

func (d *KeyType[T, TP]) MarshalText() ([]byte, error) {
	if d == nil {
		return nil, nil
	}
	return []byte(d.String()), nil
}
func (d *KeyType[T, TP]) UnmarshalText(bytes []byte) error {
	if len(bytes) == 0 {
		var zero KeyType[T, TP]
		*d = zero
		return nil
	}
	id, err := parseKey[T, TP](string(bytes))
	if err != nil {
		return errors.WithStack(err)
	}
	*d = id
	return nil
}

func suffixFromBytes(bytes []byte) string {
	return base36.EncodeToStringLc(bytes)
}

func bytesFromSuffix(suffix string) ([]byte, error) {
	bytes, err := base36.DecodeString(suffix)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode suffix %q", suffix)
	}
	return bytes, nil
}

// Generate a new key.
//
// If the payload specifies a randomness greater than 0, a random suffix will be generated.
// The payload will be parsed from payloadComponents, which must be a hyphen-separated string.
func newKey[T comparable, TP keyPayloadConstraint[T]](components ...string) (kt KeyType[T, TP]) {
	var payload TP = &kt.Payload
	if err := payload.Parse(components); err != nil {
		panic(errors.Wrapf(err, "failed to parse payload %q", strings.Join(components, "-")))
	}
	if randomness := payload.RandomBytes(); randomness > 0 {
		bytes := make([]byte, randomness)
		if _, err := randRead(bytes); err != nil {
			panic(errors.Wrap(err, "failed to generate random suffix"))
		}
		kt.Suffix = suffixFromBytes(bytes)
	}
	return kt
}

// Parse a key in the form <kind>[-<payload>][-<suffix>]
//
// Suffix will be parsed if the payload specifies a randomness greater than 0.
func parseKey[T comparable, TP keyPayloadConstraint[T]](key string) (kt KeyType[T, TP], err error) {
	components := strings.Split(key, "-")
	if len(components) == 0 {
		return kt, errors.Errorf("expected a prefix for key %q", key)
	}

	// Validate and strip kind.
	var payload TP = &kt.Payload
	if components[0] != payload.Kind() {
		return kt, errors.Errorf("expected prefix %q for key %q", payload.Kind(), key)
	}
	components = components[1:]

	// Optionally parse and strip random suffix.
	randomness := payload.RandomBytes()
	if randomness > 0 {
		if len(components) == 0 {
			return kt, errors.Errorf("expected a suffix for key %q", key)
		}
		kt.Suffix = components[len(components)-1]
		bytes, err := bytesFromSuffix(kt.Suffix)
		if err != nil {
			return kt, errors.Wrapf(err, "expected a base36 suffix for key %q", key)
		}
		if len(bytes) != randomness {
			return kt, errors.Errorf("expected a suffix of %d bytes for key %q, not %d", randomness, key, len(kt.Suffix))
		}
		components = components[:len(components)-1]
	}

	if err := payload.Parse(components); err != nil {
		return kt, errors.Wrapf(err, "failed to parse payload for key %q", key)
	}

	return kt, nil
}
