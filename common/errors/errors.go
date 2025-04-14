package errors

import (
	"strings"

	"github.com/alecthomas/errors"
	"golang.org/x/exp/maps"
)

// Deduplicate de-duplicates textually equivalent errors.
func Deduplicate(merr []error) []error {
	set := map[string]error{}
	for _, err := range merr {
		for _, subErr := range errors.UnwrapAllInnermost(err) {
			key := strings.TrimSpace(subErr.Error())
			set[key] = subErr
		}
	}
	return maps.Values(set)
}
