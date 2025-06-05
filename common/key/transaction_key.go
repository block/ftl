package key

import (
	"fmt"
	"regexp"
	"strings"

	errors "github.com/alecthomas/errors"
)

type TransactionKey = KeyType[TransactionPayload, *TransactionPayload]

func NewTransactionKey(database, id string) TransactionKey {
	return newKey[TransactionPayload](database, id)
}

func ParseTransactionKey(key string) (TransactionKey, error) {
	return errors.WithStack2(parseKey[TransactionPayload](key))
}

var transactionKeyNormaliserRe = regexp.MustCompile("[^a-zA-Z0-9]+")

type TransactionPayload struct {
	Database string
	ID       string
}

var _ KeyPayload = (*SubscriberPayload)(nil)

func (s *TransactionPayload) Kind() string   { return "txn" }
func (s *TransactionPayload) String() string { return fmt.Sprintf("%s-%s", s.Database, s.ID) }
func (s *TransactionPayload) Parse(parts []string) error {
	if len(parts) < 2 {
		return errors.Errorf("expected <database>-<id> but got %q", strings.Join(parts, "-"))
	}
	s.Database = parts[0]
	s.ID = transactionKeyNormaliserRe.ReplaceAllString(strings.Join(parts[1:], "-"), "-")
	return nil
}
func (s *TransactionPayload) RandomBytes() int { return 10 }
