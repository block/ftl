package key

import (
	"database/sql"
	"database/sql/driver"
	"encoding"

	"github.com/alecthomas/errors"
)

type Changeset = KeyType[ChangesetPayload, *ChangesetPayload]

var _ interface {
	sql.Scanner
	driver.Valuer
	encoding.TextUnmarshaler
	encoding.TextMarshaler
} = (*Changeset)(nil)

func NewChangesetKey() Changeset { return newKey[ChangesetPayload]() }
func ParseChangesetKey(key string) (Changeset, error) {
	return errors.WithStack2(parseKey[ChangesetPayload](key))
}

type ChangesetPayload struct {
	// Content is just included to make the payload non-empty.
	// Non empty payloads do not work with go2proto
	Content string
}

var _ KeyPayload = (*ChangesetPayload)(nil)

func (c *ChangesetPayload) Kind() string   { return "chs" }
func (c *ChangesetPayload) String() string { return "" }
func (c *ChangesetPayload) Parse(parts []string) error {
	if len(parts) != 0 {
		return errors.WithStack(errors.New("expected no content"))
	}
	return nil
}
func (c *ChangesetPayload) RandomBytes() int { return 10 }
