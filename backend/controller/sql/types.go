package sql

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/alecthomas/types"
	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

type NullKey = types.Option[Key]

// FromOption converts a types.Option[~ulid.ULID] to a NullKey.
func FromOption[T ~[16]byte](o types.Option[T]) NullKey {
	if v, ok := o.Get(); ok {
		return SomeKey(Key(v))
	}
	return NoneKey()
}
func SomeKey(key Key) NullKey { return types.Some(key) }
func NoneKey() NullKey        { return types.None[Key]() }

// Key is a ULID that can be used as a column in a database.
type Key ulid.ULID

func (u Key) Value() (driver.Value, error) {
	bytes := u[:]
	return bytes, nil
}

func (u *Key) Scan(src any) error {
	switch src := src.(type) {
	case string:
		id, err := uuid.Parse(src)
		if err != nil {
			return err
		}
		*u = Key(id)

	case Key:
		*u = src

	default:
		return fmt.Errorf("invalid key type %T", src)
	}
	return nil
}

func (u *Key) UnmarshalText(text []byte) error {
	id, err := uuid.ParseBytes(text)
	if err != nil {
		return err
	}
	*u = Key(id)
	return nil
}

type NullTime = types.Option[time.Time]
type NullDuration = types.Option[time.Duration]
