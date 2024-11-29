// Code generated by "enumer -type AliasKind -trimprefix AliasKind -transform=lower -json -text"; DO NOT EDIT.

package schema

import (
	"encoding/json"
	"fmt"
	"strings"
)

const _AliasKindName = "unspecifiedjson"

var _AliasKindIndex = [...]uint8{0, 11, 15}

const _AliasKindLowerName = "unspecifiedjson"

func (i AliasKind) String() string {
	if i < 0 || i >= AliasKind(len(_AliasKindIndex)-1) {
		return fmt.Sprintf("AliasKind(%d)", i)
	}
	return _AliasKindName[_AliasKindIndex[i]:_AliasKindIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _AliasKindNoOp() {
	var x [1]struct{}
	_ = x[AliasKindUnspecified-(0)]
	_ = x[AliasKindJson-(1)]
}

var _AliasKindValues = []AliasKind{AliasKindUnspecified, AliasKindJson}

var _AliasKindNameToValueMap = map[string]AliasKind{
	_AliasKindName[0:11]:       AliasKindUnspecified,
	_AliasKindLowerName[0:11]:  AliasKindUnspecified,
	_AliasKindName[11:15]:      AliasKindJson,
	_AliasKindLowerName[11:15]: AliasKindJson,
}

var _AliasKindNames = []string{
	_AliasKindName[0:11],
	_AliasKindName[11:15],
}

// AliasKindString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func AliasKindString(s string) (AliasKind, error) {
	if val, ok := _AliasKindNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _AliasKindNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to AliasKind values", s)
}

// AliasKindValues returns all values of the enum
func AliasKindValues() []AliasKind {
	return _AliasKindValues
}

// AliasKindStrings returns a slice of all String values of the enum
func AliasKindStrings() []string {
	strs := make([]string, len(_AliasKindNames))
	copy(strs, _AliasKindNames)
	return strs
}

// IsAAliasKind returns "true" if the value is listed in the enum definition. "false" otherwise
func (i AliasKind) IsAAliasKind() bool {
	for _, v := range _AliasKindValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for AliasKind
func (i AliasKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for AliasKind
func (i *AliasKind) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("AliasKind should be a string, got %s", data)
	}

	var err error
	*i, err = AliasKindString(s)
	return err
}

// MarshalText implements the encoding.TextMarshaler interface for AliasKind
func (i AliasKind) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for AliasKind
func (i *AliasKind) UnmarshalText(text []byte) error {
	var err error
	*i, err = AliasKindString(string(text))
	return err
}
