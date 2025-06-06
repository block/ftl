// Code generated by "enumer -type=Level -text -transform=lower -output log_level_string.go"; DO NOT EDIT.

package log

import (
	"fmt"
	"strings"

	"github.com/alecthomas/errors"
)

const (
	_LevelName_0      = "defaulttrace"
	_LevelLowerName_0 = "defaulttrace"
	_LevelName_1      = "debug"
	_LevelLowerName_1 = "debug"
	_LevelName_2      = "info"
	_LevelLowerName_2 = "info"
	_LevelName_3      = "warn"
	_LevelLowerName_3 = "warn"
	_LevelName_4      = "error"
	_LevelLowerName_4 = "error"
)

var (
	_LevelIndex_0 = [...]uint8{0, 7, 12}
	_LevelIndex_1 = [...]uint8{0, 5}
	_LevelIndex_2 = [...]uint8{0, 4}
	_LevelIndex_3 = [...]uint8{0, 4}
	_LevelIndex_4 = [...]uint8{0, 5}
)

func (i Level) String() string {
	switch {
	case 0 <= i && i <= 1:
		return _LevelName_0[_LevelIndex_0[i]:_LevelIndex_0[i+1]]
	case i == 5:
		return _LevelName_1
	case i == 9:
		return _LevelName_2
	case i == 13:
		return _LevelName_3
	case i == 17:
		return _LevelName_4
	default:
		return fmt.Sprintf("Level(%d)", i)
	}
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _LevelNoOp() {
	var x [1]struct{}
	_ = x[Default-(0)]
	_ = x[Trace-(1)]
	_ = x[Debug-(5)]
	_ = x[Info-(9)]
	_ = x[Warn-(13)]
	_ = x[Error-(17)]
}

var _LevelValues = []Level{Default, Trace, Debug, Info, Warn, Error}

var _LevelNameToValueMap = map[string]Level{
	_LevelName_0[0:7]:       Default,
	_LevelLowerName_0[0:7]:  Default,
	_LevelName_0[7:12]:      Trace,
	_LevelLowerName_0[7:12]: Trace,
	_LevelName_1[0:5]:       Debug,
	_LevelLowerName_1[0:5]:  Debug,
	_LevelName_2[0:4]:       Info,
	_LevelLowerName_2[0:4]:  Info,
	_LevelName_3[0:4]:       Warn,
	_LevelLowerName_3[0:4]:  Warn,
	_LevelName_4[0:5]:       Error,
	_LevelLowerName_4[0:5]:  Error,
}

var _LevelNames = []string{
	_LevelName_0[0:7],
	_LevelName_0[7:12],
	_LevelName_1[0:5],
	_LevelName_2[0:4],
	_LevelName_3[0:4],
	_LevelName_4[0:5],
}

// LevelString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func LevelString(s string) (Level, error) {
	if val, ok := _LevelNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _LevelNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, errors.Errorf("%s does not belong to Level values", s)
}

// LevelValues returns all values of the enum
func LevelValues() []Level {
	return _LevelValues
}

// LevelStrings returns a slice of all String values of the enum
func LevelStrings() []string {
	strs := make([]string, len(_LevelNames))
	copy(strs, _LevelNames)
	return strs
}

// IsALevel returns "true" if the value is listed in the enum definition. "false" otherwise
func (i Level) IsALevel() bool {
	for _, v := range _LevelValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalText implements the encoding.TextMarshaler interface for Level
func (i Level) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for Level
func (i *Level) UnmarshalText(text []byte) error {
	var err error
	*i, err = LevelString(string(text))
	return errors.WithStack(err)
}
