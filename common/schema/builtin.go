package schema

import "github.com/block/ftl/common/reflect"

// BuiltinsSource is the schema source code for built-in types.
const BuiltinsSource = `
// Built-in types for FTL.
builtin module builtin {
  // A reference to a declaration in the schema.
  export data Ref {
    module String
    name String
  }

  // HTTP request structure used for HTTP ingress verbs.
  export data HttpRequest<Body, Path, Query> {
    method String
    path String
    pathParameters Path
    query Query
    headers {String: [String]}
    body Body
  }

  // HTTP response structure used for HTTP ingress verbs.
  export data HttpResponse<Body, Error> {
    status Int
    headers {String: [String]}
    // Either "body" or "error" must be present, not both.
    body Body?
    error Error?
  }

  export data Empty {}

  // CatchRequest is a request structure for catch verbs.
  export data CatchRequest<Req> {
    verb builtin.Ref
    request Req
    requestType String
    error String
  }

  // FailedEvent is used in dead letter topics.
  export data FailedEvent<Event> {
      event Event
      error String
  }
}
`

var builtinsModuleParsed = func() *Module {
	module, err := moduleParser.ParseString("", BuiltinsSource)
	if err != nil {
		panic(err)
	}
	return module
}()

// Builtins returns a [Module] containing built-in types.
func Builtins() *Module {
	return reflect.DeepCopy(builtinsModuleParsed)
}
