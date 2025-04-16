package ingress

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
)

// buildRequestBody extracts the HttpRequest body from an HTTP request.
func buildRequestBody(route *ingressRoute, r *http.Request, sch *schema.Schema) ([]byte, error) {
	verb := &schema.Verb{}
	err := sch.ResolveToType(&schema.Ref{Name: route.verb, Module: route.module}, verb)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	request, ok := verb.Request.(*schema.Ref)
	if !ok {
		return nil, errors.Errorf("verb %s input must be a data structure", verb.Name)
	}

	var body []byte

	var requestMap map[string]any

	if metadata, ok := verb.GetMetadataIngress().Get(); ok && metadata.Type == "http" {
		pathParametersMap := map[string]string{}
		matchSegments(route.path, r.URL.Path, func(segment, value string) {
			pathParametersMap[segment] = value
		})
		pathParameters, err := manglePathParameters(pathParametersMap, request, sch)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		httpRequestBody, err := extractHTTPRequestBody(r, request, sch)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// Since the query and header parameters are a `map[string][]string`
		// we need to convert them before they go through the `transformFromAliasedFields` call
		// otherwise they will fail the type check.
		queryMap := make(map[string]any)
		for key, values := range r.URL.Query() {
			valuesAny := make([]any, len(values))
			for i, v := range values {
				valuesAny[i] = v
			}
			queryMap[key] = valuesAny
		}

		finalQueryParams, err := mangleQueryParameters(queryMap, r.URL.Query(), request, sch)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		headerMap := make(map[string]any)
		for key, values := range r.Header {
			valuesAny := make([]any, len(values))
			for i, v := range values {
				valuesAny[i] = v
			}
			headerMap[key] = valuesAny
		}

		requestMap = map[string]any{}
		requestMap["method"] = r.Method
		requestMap["path"] = r.URL.Path
		requestMap["pathParameters"] = pathParameters
		requestMap["query"] = finalQueryParams
		requestMap["headers"] = headerMap
		requestMap["body"] = httpRequestBody
	} else {
		return nil, errors.Errorf("no HTTP ingress metadata for verb %s", verb.Name)
	}

	requestMap, err = schema.TransformFromAliasedFields(sch, request, requestMap)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var opts []schema.EncodingOption
	if e, ok := slices.FindVariant[*schema.MetadataEncoding](verb.Metadata); ok && e.Lenient {
		opts = append(opts, schema.LenientMode())
	}
	err = schema.ValidateRequestMap(request, []string{request.String()}, requestMap, sch, opts...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	body, err = json.Marshal(requestMap)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return body, nil
}

func extractHTTPRequestBody(r *http.Request, ref *schema.Ref, sch *schema.Schema) (any, error) {
	bodyField, err := getField("body", ref, sch)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if ref, ok := bodyField.Type.(*schema.Ref); ok {
		if err := sch.ResolveToType(ref, &schema.Data{}); err == nil {
			return errors.WithStack2(buildRequestMap(r))
		}
	}

	bodyData, err := readRequestBody(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return errors.WithStack2(valueForData(bodyField.Type, bodyData))
}

// Takes the map of path parameters and transforms them into the appropriate type
func manglePathParameters(params map[string]string, ref *schema.Ref, sch *schema.Schema) (any, error) {

	paramsField, err := getField("pathParameters", ref, sch)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	switch m := paramsField.Type.(type) {
	case *schema.Map:
		ret := map[string]any{}
		for k, v := range params {
			ret[k] = v
		}
		return ret, nil
	case *schema.Ref:
		data := &schema.Data{}
		err := sch.ResolveToType(m, data)
		if err != nil {
			return nil, errors.Wrap(err, "failed to resolve path parameter type")
		}
		return errors.WithStack2(parsePathParams(params, data))
	default:
	}
	// This is a scalar, there should only be a single param
	// This is validated by the schema, we don't need extra validation here
	for _, val := range params {
		return errors.WithStack2(parseScalar(paramsField.Type, val))
	}
	// Empty map
	return map[string]any{}, nil
}

func parseScalar(paramsField schema.Type, val string) (any, error) {
	switch paramsField.(type) {
	case *schema.String:
		return val, nil
	case *schema.Int:
		parsed, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse int from path parameter")
		}
		return parsed, nil
	case *schema.Float:
		float, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse float from path parameter")
		}
		return float, nil
	case *schema.Bool:
		// TODO: is anything else considered truthy?
		return val == "true", nil
	default:
		return nil, errors.Errorf("unsupported path parameter type %T", paramsField)
	}
}

// Takes the map of query parameters and transforms them into the appropriate type
func mangleQueryParameters(params map[string]any, underlying map[string][]string, ref *schema.Ref, sch *schema.Schema) (any, error) {

	paramsField, err := getField("query", ref, sch)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	switch m := paramsField.Type.(type) {
	case *schema.Map:
		if _, ok := m.Value.(*schema.Array); ok {
			return params, nil
		} else if _, ok := m.Value.(*schema.String); ok {
			// We need to turn them into straight strings
			newParams := map[string]any{}
			for k, v := range underlying {
				if len(v) > 0 {
					newParams[k] = v[0]
				}
			}
			return newParams, nil
		}
	case *schema.Ref:
		data := &schema.Data{}
		err := sch.ResolveToType(m, data)
		if err != nil {
			return nil, errors.Wrap(err, "failed to resolve query parameter type")
		}
		return errors.WithStack2(parseQueryParams(underlying, data))
	case *schema.Unit:
		return params, nil
	default:
	}

	return nil, errors.Errorf("unsupported query parameter type %v", paramsField.Type)
}

func valueForData(typ schema.Type, data []byte) (any, error) {
	switch typ.(type) {
	case *schema.Ref:
		var bodyMap map[string]any
		err := json.Unmarshal(data, &bodyMap)
		if err != nil {
			return nil, errors.Wrap(err, "HTTP request body is not valid JSON")
		}
		return bodyMap, nil

	case *schema.Array:
		var rawData []json.RawMessage
		err := json.Unmarshal(data, &rawData)
		if err != nil {
			return nil, errors.Wrap(err, "HTTP request body is not a valid JSON array")
		}

		arrayData := make([]any, len(rawData))
		for i, rawElement := range rawData {
			var parsedElement any
			err := json.Unmarshal(rawElement, &parsedElement)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse array element")
			}
			arrayData[i] = parsedElement
		}

		return arrayData, nil

	case *schema.Map:
		var bodyMap map[string]any
		err := json.Unmarshal(data, &bodyMap)
		if err != nil {
			return nil, errors.Wrap(err, "HTTP request body is not valid JSON")
		}
		return bodyMap, nil

	case *schema.Bytes:
		return data, nil

	case *schema.String:
		return string(data), nil

	case *schema.Int:
		intVal, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse integer from request body")
		}
		return intVal, nil

	case *schema.Float:
		floatVal, err := strconv.ParseFloat(string(data), 64)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse float from request body")
		}
		return floatVal, nil

	case *schema.Bool:
		boolVal, err := strconv.ParseBool(string(data))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse boolean from request body")
		}
		return boolVal, nil

	case *schema.Unit:
		return map[string]any{}, nil

	default:
		return nil, errors.Errorf("unsupported data type %T", typ)
	}
}

func readRequestBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()
	bodyData, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading request body")
	}
	return bodyData, nil
}

func buildRequestMap(r *http.Request) (map[string]any, error) {
	switch r.Method {
	case http.MethodPost, http.MethodPut:
		var bodyMap map[string]any
		err := json.NewDecoder(r.Body).Decode(&bodyMap)
		if err != nil {
			return nil, errors.Wrap(err, "HTTP request body is not valid JSON")
		}

		return bodyMap, nil
	default:
		return nil, nil
	}
}

func parseQueryParams(values map[string][]string, data *schema.Data) (map[string]any, error) {
	if jsonStr, ok := values["@json"]; ok {
		if len(values) > 1 {
			return nil, errors.Errorf("only '@json' parameter is allowed, but other parameters were found")
		}
		if len(jsonStr) > 1 {
			return nil, errors.Errorf("'@json' parameter must be provided exactly once")
		}

		return errors.WithStack2(decodeQueryJSON(jsonStr[0]))
	}

	queryMap := make(map[string]any)
	for key, value := range values {
		if hasInvalidQueryChars(key) {
			return nil, errors.Errorf("complex key %q is not supported, use '@json=' instead", key)
		}
		var field *schema.Field
		for _, f := range data.Fields {
			if jsonAlias, ok := f.Alias(schema.AliasKindJSON).Get(); (ok && jsonAlias == key) || f.Name == key {
				field = f
			}
			for _, typeParam := range data.TypeParameters {
				if typeParam.Name == key {
					field = &schema.Field{
						Name: key,
						Type: &schema.Ref{Pos: typeParam.Pos, Name: typeParam.Name},
					}
				}
			}
		}

		if field == nil {
			queryMap[key] = value
			continue
		}

		val, err := valueForField(field.Type, value)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse query parameter %q", key)
		}
		if v, ok := val.Get(); ok {
			queryMap[key] = v
		}
	}
	return queryMap, nil
}

func parsePathParams(values map[string]string, data *schema.Data) (map[string]any, error) {
	// This is annoyingly close to the query params logic, but just disimilar enough to not be easily refactored
	pathMap := make(map[string]any)
	for key, value := range values {
		var field *schema.Field
		for _, f := range data.Fields {
			if jsonAlias, ok := f.Alias(schema.AliasKindJSON).Get(); (ok && jsonAlias == key) || f.Name == key {
				field = f
			}
		}

		if field == nil {
			pathMap[key] = value
			continue
		}

		val, err := valueForField(field.Type, []string{value})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse path parameter %q", key)
		}
		if v, ok := val.Get(); ok {
			pathMap[key] = v
		}
	}
	return pathMap, nil
}

func valueForField(typ schema.Type, value []string) (optional.Option[any], error) {
	switch t := typ.(type) {
	case *schema.Bytes, *schema.Map, *schema.Time,
		*schema.Unit, *schema.Ref, *schema.Any:

	case *schema.Int, *schema.Float, *schema.String, *schema.Bool:
		if len(value) > 1 {
			return optional.None[any](), errors.Errorf("multiple values are not supported")
		}
		if hasInvalidQueryChars(value[0]) {
			return optional.None[any](), errors.Errorf("complex value %q is not supported, use '@json=' instead", value[0])
		}

		parsed, err := parseScalar(typ, value[0])
		if err != nil {
			return optional.None[any](), errors.Wrap(err, "failed to parse int from path parameter")
		}
		return optional.Some[any](parsed), nil

	case *schema.Array:
		for _, v := range value {
			if hasInvalidQueryChars(v) {
				return optional.None[any](), errors.Errorf("complex value %q is not supported, use '@json=' instead", v)
			}
		}
		ret := []any{}
		for _, v := range value {
			field, err := valueForField(t.Element, []string{v})
			if err != nil {
				return optional.None[any](), errors.Wrap(err, "failed to parse array element")
			}
			if unwrapped, ok := field.Get(); ok {
				ret = append(ret, unwrapped)
			}
		}
		return optional.Some[any](ret), nil

	case *schema.Optional:
		if len(value) > 0 {
			return errors.WithStack2(valueForField(t.Type, value))
		}

	default:
		return optional.None[any](), errors.Errorf("unsupported type %T", typ)
	}

	return optional.Some[any](value), nil
}

func decodeQueryJSON(query string) (map[string]any, error) {
	decodedJSONStr, err := url.QueryUnescape(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode '@json' query parameter")
	}

	// Unmarshal the JSON string into a map
	var resultMap map[string]any
	err = json.Unmarshal([]byte(decodedJSONStr), &resultMap)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse '@json' query parameter")
	}

	return resultMap, nil
}

func hasInvalidQueryChars(s string) bool {
	return strings.ContainsAny(s, "{}[]|\\^`")
}
