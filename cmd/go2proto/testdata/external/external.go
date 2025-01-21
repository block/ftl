package external

import (
	"encoding"
	"errors"
	"strings"
)

type Root struct {
	Prefix string
	Suffix string
}

var _ encoding.TextMarshaler = (*Root)(nil)
var _ encoding.TextUnmarshaler = (*Root)(nil)

func (r *Root) MarshalText() ([]byte, error) {
	return []byte(r.Prefix + "." + r.Suffix), nil
}
func (r *Root) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), ".")
	if len(parts) != 2 {
		return errors.New("expected exactly one '.'")
	}
	r.Prefix = parts[0]
	r.Suffix = parts[1]
	return nil
}
