// Package encoding defines the internal encoding that FTL uses to encode and
// decode messages. It is currently JSON.
package encoding

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/alecthomas/errors"

	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/common/strcase"
)

var (
	optionMarshaler   = reflect.TypeFor[OptionMarshaler]()
	optionUnmarshaler = reflect.TypeFor[OptionUnmarshaler]()
)

type OptionMarshaler interface {
	Marshal(w *bytes.Buffer, encode func(v reflect.Value, w *bytes.Buffer) error) error
}
type OptionUnmarshaler interface {
	Unmarshal(d *json.Decoder, isNull bool, decode func(d *json.Decoder, v reflect.Value) error) error
}

func Marshal(v any) ([]byte, error) {
	w := &bytes.Buffer{}
	err := encodeValue(reflect.ValueOf(v), w)
	return w.Bytes(), errors.WithStack(err)
}

func encodeValue(v reflect.Value, w *bytes.Buffer) error {
	if !v.IsValid() {
		w.WriteString("null")
		return nil
	}

	if v.Kind() == reflect.Ptr {
		return errors.Errorf("pointer types are not supported: %s", v.Type())
	}

	t := v.Type()
	// Special-cased types
	switch {
	case reflection.IsKnownExternalType(t):
		// external types use the stdlib JSON encoding
		fallthrough

	case t == reflect.TypeFor[time.Time]():
		data, err := json.Marshal(v.Interface())
		if err != nil {
			return errors.WithStack(err)
		}
		w.Write(data)
		return nil

	case t.Implements(optionMarshaler):
		enc := v.Interface().(OptionMarshaler) //nolint:forcetypeassert
		return errors.WithStack(enc.Marshal(w, encodeValue))

	// TODO(Issue #1439): remove this special case by removing all usage of
	// json.RawMessage, which is not a type we support.
	case t == reflect.TypeFor[json.RawMessage]():
		data, err := json.Marshal(v.Interface())
		if err != nil {
			return errors.WithStack(err)
		}
		w.Write(data)
		return nil
	}

	switch v.Kind() {
	case reflect.Struct:
		return errors.WithStack(encodeStruct(v, w))

	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			if v.Type() == reflect.TypeFor[json.RawMessage]() {
				// json.RawMessage should be written as is (it's already JSON text)
				_, err := w.Write(v.Bytes())
				return errors.WithStack(err)
			}
			// Other []byte types are base64 encoded into a JSON string
			return errors.WithStack(encodeBytes(v, w))
		}
		return errors.WithStack(encodeSlice(v, w))

	case reflect.Map:
		return errors.WithStack(encodeMap(v, w))

	case reflect.String:
		return errors.WithStack(encodeString(v, w))

	case reflect.Int, reflect.Int64:
		return errors.WithStack(encodeInt(v, w))

	case reflect.Float64:
		return errors.WithStack(encodeFloat(v, w))

	case reflect.Bool:
		return errors.WithStack(encodeBool(v, w))

	case reflect.Interface:
		if t == reflect.TypeFor[any]() {
			return errors.WithStack(encodeValue(v.Elem(), w))
		}

		if vName, ok := reflection.GetVariantByType(v.Type(), v.Elem().Type()).Get(); ok {
			sumType := struct {
				Name  string
				Value any
			}{Name: vName, Value: v.Elem().Interface()}
			return errors.WithStack(encodeValue(reflect.ValueOf(sumType), w))
		}

		return errors.Errorf("the only supported interface types are enums or any, not %s", t)

	default:
		panic(fmt.Sprintf("unsupported type: %s", v.Type()))
	}
}

func encodeStruct(v reflect.Value, w *bytes.Buffer) error {
	w.WriteRune('{')
	afterFirst := false
	for i := range v.NumField() {
		ft := v.Type().Field(i)
		fv := v.Field(i)
		// TODO: If these fields are skipped, the ingress encoder will not include
		// them in the output. There should ideally be no relationship between
		// the ingress encoder and the encoding package, but for now this is the
		// simplest solution.

		// t := ft.Type
		// // Types that can be skipped if they're zero.
		// if (t.Kind() == reflect.Slice && fv.Len() == 0) ||
		// 	(t.Kind() == reflect.Map && fv.Len() == 0) ||
		// 	(t.String() == "ftl.Unit" && fv.IsZero()) ||
		// 	(strings.HasPrefix(t.String(), "ftl.Option[") && fv.IsZero()) ||
		// 	(t == reflect.TypeOf((*any)(nil)).Elem() && fv.IsZero()) {
		// 	continue
		// }
		if isTaggedOmitempty(v, i) && fv.IsZero() {
			continue
		}
		if afterFirst {
			w.WriteRune(',')
		}
		afterFirst = true
		w.WriteString(`"` + strcase.ToLowerCamel(ft.Name) + `":`)
		if err := encodeValue(fv, w); err != nil {
			return errors.WithStack(err)
		}
	}
	w.WriteRune('}')
	return nil
}

func isTaggedOmitempty(v reflect.Value, i int) bool {
	tag := v.Type().Field(i).Tag
	tagVals := strings.Split(tag.Get("json"), ",")
	for _, tagVal := range tagVals {
		if strings.TrimSpace(tagVal) == "omitempty" {
			return true
		}
	}
	return false
}

func encodeBytes(v reflect.Value, w *bytes.Buffer) error {
	data := base64.StdEncoding.EncodeToString(v.Bytes())
	fmt.Fprintf(w, "%q", data)
	return nil
}

func encodeSlice(v reflect.Value, w *bytes.Buffer) error {
	w.WriteRune('[')
	for i := range v.Len() {
		if i > 0 {
			w.WriteRune(',')
		}
		if err := encodeValue(v.Index(i), w); err != nil {
			return errors.WithStack(err)
		}
	}
	w.WriteRune(']')
	return nil
}

func encodeMap(v reflect.Value, w *bytes.Buffer) error {
	w.WriteRune('{')
	for i, key := range v.MapKeys() {
		if i > 0 {
			w.WriteRune(',')
		}
		w.WriteRune('"')
		w.WriteString(key.String())
		w.WriteString(`":`)
		if err := encodeValue(v.MapIndex(key), w); err != nil {
			return errors.WithStack(err)
		}
	}
	w.WriteRune('}')
	return nil
}

func encodeBool(v reflect.Value, w *bytes.Buffer) error {
	if v.Bool() {
		w.WriteString("true")
	} else {
		w.WriteString("false")
	}
	return nil
}

func encodeInt(v reflect.Value, w *bytes.Buffer) error {
	fmt.Fprintf(w, "%d", v.Int())
	return nil
}

func encodeFloat(v reflect.Value, w *bytes.Buffer) error {
	fmt.Fprintf(w, "%g", v.Float())
	return nil
}

func encodeString(v reflect.Value, w *bytes.Buffer) error {
	fmt.Fprintf(w, "%q", v.String())
	return nil
}

func Unmarshal(data []byte, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.Errorf("unmarshal expects a non-nil pointer")
	}

	d := json.NewDecoder(bytes.NewReader(data))
	return errors.WithStack(decodeValue(d, rv.Elem()))
}

func decodeValue(d *json.Decoder, v reflect.Value) error {
	if !v.CanSet() {
		return errors.Errorf("cannot set value: %s", v.Type())
	}

	if v.Kind() == reflect.Ptr {
		return errors.Errorf("pointer types are not supported: %s", v.Type())
	}

	t := v.Type()
	// Special-case types
	switch {
	case reflection.IsKnownExternalType(t):
		// external types use the stdlib JSON decoding
		fallthrough

	case t == reflect.TypeFor[time.Time]():
		return errors.WithStack(d.Decode(v.Addr().Interface()))

	case v.CanAddr() && v.Addr().Type().Implements(optionUnmarshaler):
		v = v.Addr()
		fallthrough

	case t.Implements(optionUnmarshaler):
		if v.IsNil() {
			v.Set(reflect.New(t.Elem()))
		}
		dec := v.Interface().(OptionUnmarshaler) //nolint:forcetypeassert
		return errors.WithStack(handleIfNextTokenIsNull(d, func(d *json.Decoder) error {
			return errors.WithStack(dec.Unmarshal(d, true, decodeValue))
		}, func(d *json.Decoder) error {
			return errors.WithStack(dec.Unmarshal(d, false, decodeValue))
		}))
	}

	switch v.Kind() {
	case reflect.Bool:
		token, err := d.Token()
		if err != nil {
			return errors.Wrap(err, "failed to read bool token")
		}
		switch n := token.(type) {
		case float64:
			v.SetBool(n != 0)
		case int64:
			v.SetBool(n != 0)
		case bool:
			v.SetBool(n)
		case string:
			v.SetBool(n == "true")
		default:
			return errors.Errorf("cannot convert %T to bool", token)
		}
		return nil

	case reflect.Struct:
		return errors.WithStack(decodeStruct(d, v))

	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			if v.Type() == reflect.TypeFor[json.RawMessage]() {
				// Use standard library's decoding for json.RawMessage.
				// It expects a pointer, so we use v.Addr().Interface().
				return errors.WithStack(d.Decode(v.Addr().Interface()))
			}
			// For other []byte types (not json.RawMessage), use the existing decodeBytes
			// which expects the JSON string value to be base64 encoded.
			return errors.WithStack(decodeBytes(d, v))
		}
		return errors.WithStack(decodeSlice(d, v))

	case reflect.Map:
		return errors.WithStack(decodeMap(d, v))

	case reflect.Interface:
		if reflection.IsSumTypeDiscriminator(v.Type()) {
			return errors.WithStack(decodeSumType(d, v))
		}

		if v.Type().NumMethod() != 0 {
			return errors.Errorf("the only supported interface types are enums or any, not %s", v.Type())
		}
		fallthrough

	default:
		return errors.WithStack(d.Decode(v.Addr().Interface()))
	}
}

func decodeStruct(d *json.Decoder, v reflect.Value) error {
	if err := expectDelim(d, '{'); err != nil {
		return errors.WithStack(err)
	}

	for d.More() {
		keyToken, err := d.Token()
		if err != nil {
			return errors.WithStack(err)
		}
		key, ok := keyToken.(string)
		if !ok {
			return errors.Errorf("expected string key, got %T (value: %v)", keyToken, keyToken)
		}

		field := v.FieldByNameFunc(func(s string) bool {
			return strcase.ToLowerCamel(s) == key
		})
		if !field.IsValid() {
			// Skip the value token for unknown fields
			if _, err := d.Token(); err != nil {
				return errors.Wrapf(err, "failed to skip unknown field %s", key)
			}
			continue
		}

		fieldTypeStr := field.Type().String()
		switch {
		case fieldTypeStr == "*Unit" || fieldTypeStr == "Unit":
			if fieldTypeStr == "*Unit" && field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			// Skip the value token for Unit types
			if _, err := d.Token(); err != nil {
				return errors.Wrapf(err, "failed to skip Unit field %s", key)
			}
		default:
			if err := decodeValue(d, field); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	// consume the closing delimiter of the object
	_, err := d.Token()
	return errors.WithStack(err)
}

func decodeBytes(d *json.Decoder, v reflect.Value) error {
	var b []byte
	if err := d.Decode(&b); err != nil {
		return errors.WithStack(err)
	}
	v.SetBytes(b)
	return nil
}

func decodeSlice(d *json.Decoder, v reflect.Value) error {
	if err := expectDelim(d, '['); err != nil {
		return errors.WithStack(err)
	}

	for d.More() {
		newElem := reflect.New(v.Type().Elem()).Elem()
		if err := decodeValue(d, newElem); err != nil {
			return errors.WithStack(err)
		}
		v.Set(reflect.Append(v, newElem))
	}
	// consume the closing delimiter of the slice
	_, err := d.Token()
	return errors.WithStack(err)
}

func decodeMap(d *json.Decoder, v reflect.Value) error {
	if err := expectDelim(d, '{'); err != nil {
		return errors.WithStack(err)
	}

	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	valType := v.Type().Elem()
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return errors.WithStack(err)
		}

		newElem := reflect.New(valType).Elem()
		if err := decodeValue(d, newElem); err != nil {
			return errors.WithStack(err)
		}

		v.SetMapIndex(reflect.ValueOf(key), newElem)
	}
	// consume the closing delimiter of the map
	_, err := d.Token()
	return errors.WithStack(err)
}

func decodeSumType(d *json.Decoder, v reflect.Value) error {
	var sumType struct {
		Name  string
		Value json.RawMessage
	}
	err := d.Decode(&sumType)
	if err != nil {
		return errors.WithStack(err)
	}
	if sumType.Name == "" {
		return errors.Errorf("no name found for type enum variant")
	}
	if sumType.Value == nil {
		return errors.Errorf("no value found for type enum variant")
	}

	variantType, ok := reflection.GetVariantByName(v.Type(), sumType.Name).Get()
	if !ok {
		return errors.Errorf("no enum variant found by name %s", sumType.Name)
	}

	out := reflect.New(variantType)
	if err := decodeValue(json.NewDecoder(bytes.NewReader(sumType.Value)), out.Elem()); err != nil {
		return errors.WithStack(err)
	}
	if !out.Type().AssignableTo(v.Type()) {
		return errors.Errorf("cannot assign %s to %s", out.Type(), v.Type())
	}
	v.Set(out.Elem())

	return nil
}

func expectDelim(d *json.Decoder, expected json.Delim) error {
	token, err := d.Token()
	if err != nil {
		return errors.WithStack(err)
	}
	delim, ok := token.(json.Delim)
	if !ok || delim != expected {
		return errors.Errorf("expected delimiter %q, got %q", expected, token)
	}
	return nil
}

func handleIfNextTokenIsNull(d *json.Decoder, ifNullFn func(*json.Decoder) error, elseFn func(*json.Decoder) error) error {
	isNull, err := isNextTokenNull(d)
	if err != nil {
		return errors.WithStack(err)
	}
	if isNull {
		err = ifNullFn(d)
		if err != nil {
			return errors.WithStack(err)
		}
		// Consume the null token
		_, err := d.Token()
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	}
	return errors.WithStack(elseFn(d))
}

// isNextTokenNull implements a cheap/dirty version of `Peek()`, which json.Decoder does
// not support.
//
// It reads the buffered data and checks if the next token is "null" without actually consuming the token.
func isNextTokenNull(d *json.Decoder) (bool, error) {
	s, err := io.ReadAll(d.Buffered())
	if err != nil {
		return false, errors.WithStack(err)
	}

	remaining := s[bytes.IndexFunc(s, isDelim)+1:]
	secondDelim := bytes.IndexFunc(remaining, isDelim)
	if secondDelim == -1 {
		secondDelim = len(remaining) // No delimiters found, read until the end
	}

	return strings.TrimSpace(string(remaining[:secondDelim])) == "null", nil
}

func isDelim(r rune) bool {
	switch r {
	case ',', ':', '{', '}', '[', ']':
		return true
	}
	return false
}
