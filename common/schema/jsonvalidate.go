package schema

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/errors"
	sets "github.com/deckarep/golang-set/v2"
)

type path []string

func (p path) String() string {
	return strings.TrimLeft(strings.Join(p, ""), ".")
}

type encodingOptions struct {
	lenient bool
}

type EncodingOption func(option *encodingOptions)

func LenientMode() EncodingOption {
	return func(eo *encodingOptions) {
		eo.lenient = true
	}
}

// ValidateJSONValue validates a given JSON value against the provided schema.
func ValidateJSONValue(fieldType Type, path path, value any, sch *Schema, opts ...EncodingOption) error {
	cfg := &encodingOptions{}
	for _, opt := range opts {
		opt(cfg)
	}
	return errors.WithStack(validateJSONValue(fieldType, path, value, sch, cfg))
}

func validateJSONValue(fieldType Type, path path, value any, sch *Schema, opts *encodingOptions) error { //nolint:maintidx
	var typeMatches bool
	switch fieldType := fieldType.(type) {
	case *Any:
		typeMatches = true

	case *Unit:
		if valueMap, ok := value.(map[string]any); !ok || len(valueMap) != 0 {
			return errors.Errorf("%s must be an empty map", path)
		}
		return nil

	case *Time:
		str, ok := value.(string)
		if !ok {
			return errors.Errorf("time %s must be an RFC3339 formatted string", path)
		}
		_, err := time.Parse(time.RFC3339Nano, str)
		if err != nil {
			return errors.Wrapf(err, "time %s must be an RFC3339 formatted string", path)
		}
		return nil

	case *Int:
		switch value := value.(type) {
		case int64:
			typeMatches = true
		case float64:
			// Check if the float64 is a whole number
			if value == float64(int64(value)) {
				typeMatches = true
			}
		case string:
			if _, err := strconv.ParseInt(value, 10, 64); err == nil {
				typeMatches = true
			}
		}

	case *Float:
		switch value := value.(type) {
		case float64:
			typeMatches = true
		case string:
			if _, err := strconv.ParseFloat(value, 64); err == nil {
				typeMatches = true
			}
		}

	case *String:
		_, typeMatches = value.(string)

	case *Bool:
		switch value := value.(type) {
		case bool:
			typeMatches = true
		case string:
			if _, err := strconv.ParseBool(value); err == nil {
				typeMatches = true
			}
		}

	case *Array:
		valueSlice, ok := value.([]any)
		if !ok {
			return errors.Errorf("%s is not a slice", path)
		}

		elementType := fieldType.Element
		for i, elem := range valueSlice {
			elemPath := append(path, fmt.Sprintf("[%d]", i)) //nolint:gocritic
			if err := validateJSONValue(elementType, elemPath, elem, sch, opts); err != nil {
				return errors.WithStack(err)
			}
		}
		typeMatches = true

	case *Map:
		valueMap, ok := value.(map[string]any)
		if !ok {
			return errors.Errorf("%s is not a map", path)
		}

		keyType := fieldType.Key
		valueType := fieldType.Value
		for key, elem := range valueMap {
			elemPath := append(path, fmt.Sprintf("[%q]", key)) //nolint:gocritic
			if err := validateJSONValue(keyType, elemPath, key, sch, opts); err != nil {
				return errors.WithStack(err)
			}
			if err := validateJSONValue(valueType, elemPath, elem, sch, opts); err != nil {
				return errors.WithStack(err)
			}
		}
		typeMatches = true

	case *Ref:
		typeDecl, err := sch.ResolveType(fieldType)
		if err != nil {
			return errors.Wrap(err, "failed to resolve reference")
		}
		return errors.WithStack(validateJSONValue(typeDecl, path, value, sch, opts))

	case *Data:
		if valueMap, ok := value.(map[string]any); ok {
			transformedMap, err := TransformFromAliasedFields(sch, fieldType, valueMap)
			if err != nil {
				return errors.Wrap(err, "failed to transform aliased fields")
			}

			if err := validateMap(fieldType, path, transformedMap, sch, opts); err != nil {
				return errors.WithStack(err)
			}
			typeMatches = true
		}

	case *TypeAlias:
		return errors.WithStack(validateJSONValue(fieldType.Type, path, value, sch, opts))

	case *Enum:
		var inputName any
		inputName = value
		for _, v := range fieldType.Variants {
			switch t := v.Value.(type) {
			case *StringValue:
				if valueStr, ok := value.(string); ok {
					if t.Value == valueStr {
						typeMatches = true
						break
					}
				}
			case *IntValue:
				switch value := value.(type) {
				case int, int64:
					if t.Value == value {
						typeMatches = true
						break
					}
				case float64:
					if float64(t.Value) == value {
						typeMatches = true
						break
					}
				}
			case *TypeValue:
				if reqVariant, ok := value.(map[string]any); ok {
					vName, ok := reqVariant["name"]
					if !ok {
						return errors.Errorf(`missing name field in enum type %q: expected structure is `+
							"{\"name\": \"<variant name>\", \"value\": <variant value>}", value)
					}
					vNameStr, ok := vName.(string)
					if !ok {
						return errors.Errorf(`invalid type for enum %q; name field must be a string, was %T`,
							fieldType, vName)
					}
					inputName = fmt.Sprintf("%q", vNameStr)

					vValue, ok := reqVariant["value"]
					if !ok {
						return errors.Errorf(`missing value field in enum type %q: expected structure is `+
							"{\"name\": \"<variant name>\", \"value\": <variant value>}", value)
					}

					if v.Name == vNameStr {
						return errors.WithStack(validateJSONValue(t.Value, path, vValue, sch, opts))
					}
				} else {
					return errors.Errorf(`malformed enum type %s: expected structure is `+
						"{\"name\": \"<variant name>\", \"value\": <variant value>}", path)
				}
			}
		}
		if !typeMatches {
			return errors.Errorf("%s is not a valid variant of enum %s", inputName, fieldType.Name)
		}

	case *Bytes:
		_, typeMatches = value.([]byte)
		if bodyStr, ok := value.(string); ok {
			_, err := base64.StdEncoding.DecodeString(bodyStr)
			if err != nil {
				return errors.Errorf("%s is not a valid base64 string", path)
			}
			typeMatches = true
		}

	case *Optional:
		if value == nil {
			typeMatches = true
		} else {
			return errors.WithStack(validateJSONValue(fieldType.Type, path, value, sch, opts))
		}
	}

	if !typeMatches {
		return errors.Errorf("%s has wrong type, expected %s found %T", path, fieldType, value)
	}
	return nil
}

// ValidateRequestMap validates a given JSON map against the provided schema.
func ValidateRequestMap(ref *Ref, path path, request map[string]any, sch *Schema, opts ...EncodingOption) error {
	cfg := &encodingOptions{}
	for _, opt := range opts {
		opt(cfg)
	}
	symbol, err := sch.ResolveRequestResponseType(ref)
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(validateRequestMap(symbol, path, request, sch, cfg))
}

func validateRequestMap(symbol Symbol, path path, request map[string]any, sch *Schema, opts *encodingOptions) error {
	var errs []error
	switch symbol := symbol.(type) {
	case *Data:
		errs = append(errs, validateMap(symbol, path, request, sch, opts))
	case *Map:
		// TODO: Validate key/value types-
		//
	case *Any:
	default:
		errs = append(errs, errors.Errorf("unsupported symbol type %T", symbol))
	}
	return errors.WithStack(errors.Join(errs...))
}

// validateMap validates a given JSON map against the provided Data structure.
func validateMap(data *Data, path path, request map[string]any, sch *Schema, opts *encodingOptions) error {
	var errs []error
	validFields := sets.NewSet[string]()
	for _, field := range data.Fields {
		validFields.Add(field.Name)
		fieldPath := append(path, "."+field.Name) //nolint:gocritic

		value, haveValue := request[field.Name]
		if !haveValue && !allowMissingField(field) {
			errs = append(errs, errors.Errorf("%s is required", fieldPath))
			continue
		}

		if haveValue {
			err := validateJSONValue(field.Type, fieldPath, value, sch, opts)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	if !opts.lenient {
		for key := range request {
			if !validFields.Contains(key) {
				errs = append(errs, errors.Errorf("%s is not a valid field", append(path, "."+key)))
			}
		}
	}
	return errors.WithStack(errors.Join(errs...))
}

// Fields of these types can be omitted from the JSON representation.
func allowMissingField(field *Field) bool {
	switch field.Type.(type) {
	case *Optional, *Any, *Array, *Map, *Bytes, *Unit:
		return true

	case *Bool, *Ref, *Float, *Int, *String, *Time, *Data, *Enum, *TypeAlias:
	}
	return false
}

func TransformAliasedFields(sch *Schema, t Type, obj any, aliaser func(obj map[string]any, field *Field) string) error {
	if obj == nil {
		return nil
	}
	switch t := t.(type) {
	case *Ref:
		dt, err := sch.ResolveType(t)
		if err != nil {
			return errors.Wrapf(err, "failed to resolve ref %s", t)
		}
		return errors.WithStack(TransformAliasedFields(sch, dt, obj, aliaser))

	case *Data: // NOTE: assumed to have been monomorphised already.
		m, ok := obj.(map[string]any)
		if !ok {
			return errors.Errorf("%s: expected map, got %T", t.Pos, obj)
		}
		for _, field := range t.Fields {
			name := aliaser(m, field)
			if err := TransformAliasedFields(sch, field.Type, m[name], aliaser); err != nil {
				return errors.WithStack(err)
			}
		}

	case *Enum:
		if t.IsValueEnum() {
			return nil
		}

		// type enum
		m, ok := obj.(map[string]any)
		if !ok {
			return errors.Errorf("%s: expected map, got %T", t.Pos, obj)
		}
		name, ok := m["name"]
		if !ok {
			return errors.Errorf("%s: expected type enum request to have 'name' field", t.Pos)
		}
		nameStr, ok := name.(string)
		if !ok {
			return errors.Errorf("%s: expected 'name' field to be a string, got %T", t.Pos, name)
		}

		value, ok := m["value"]
		if !ok {
			return errors.Errorf("%s: expected type enum request to have 'value' field", t.Pos)
		}

		for _, v := range t.Variants {
			if v.Name == nameStr {
				if err := TransformAliasedFields(sch, v.Value.(*TypeValue).Value, value, aliaser); err != nil { //nolint:forcetypeassert
					return errors.Wrapf(err, "%s", v.Name)
				}
			}
		}
	case *TypeAlias:
		return errors.WithStack(TransformAliasedFields(sch, t.Type, obj, aliaser))

	case *Array:
		a, ok := obj.([]any)
		if !ok {
			return errors.Errorf("%s: expected array, got %T", t.Pos, obj)
		}
		for i, elem := range a {
			if err := TransformAliasedFields(sch, t.Element, elem, aliaser); err != nil {
				return errors.Wrapf(err, "[%d]", i)
			}
		}

	case *Map:
		m, ok := obj.(map[string]any)
		if !ok {
			return errors.Errorf("%s: expected map, got %T", t.Pos, obj)
		}
		for key, value := range m {
			if err := TransformAliasedFields(sch, t.Key, key, aliaser); err != nil {
				return errors.Wrapf(err, "[%q] (key)", key)
			}
			if err := TransformAliasedFields(sch, t.Value, value, aliaser); err != nil {
				return errors.Wrapf(err, "[%q] (value)", key)
			}
		}

	case *Optional:
		return errors.WithStack(TransformAliasedFields(sch, t.Type, obj, aliaser))

	case *Any, *Bool, *Bytes, *Float, *Int,
		*String, *Time, *Unit:
	}
	return nil
}

func TransformFromAliasedFields(sch *Schema, t Type, request map[string]any) (map[string]any, error) {
	return request, errors.WithStack(TransformAliasedFields(sch, t, request, func(obj map[string]any, field *Field) string {
		if jsonAlias, ok := field.Alias(AliasKindJSON).Get(); ok {
			if _, ok := obj[field.Name]; !ok && obj[jsonAlias] != nil {
				obj[field.Name] = obj[jsonAlias]
				delete(obj, jsonAlias)
			}
		}
		return field.Name
	}))
}

func TransformToAliasedFields(sch *Schema, t Type, request map[string]any) (map[string]any, error) {
	return request, errors.WithStack(TransformAliasedFields(sch, t, request, func(obj map[string]any, field *Field) string {
		if jsonAlias, ok := field.Alias(AliasKindJSON).Get(); ok && field.Name != jsonAlias {
			obj[jsonAlias] = obj[field.Name]
			delete(obj, field.Name)
			return jsonAlias
		}
		return field.Name
	}))
}

// ValidateJSONCall validates a given JSON request against the provided schema when calling a verb.
func ValidateJSONCall(jsonBytes []byte, verbRef *Ref, sch *Schema) error {
	var request any
	err := json.Unmarshal(jsonBytes, &request)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal request body")
	}

	decl, ok := sch.Resolve(verbRef).Get()
	if !ok {
		return errors.Errorf("unknown verb: %s", verbRef.String())
	}
	verb, ok := decl.(*Verb)
	if !ok {
		return errors.Errorf("%s is not a verb", verbRef.String())
	}

	if err := ValidateJSONValue(verb.Request, nil, request, sch); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
