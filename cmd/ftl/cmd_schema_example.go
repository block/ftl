package main

import (
	"fmt"

	"github.com/alecthomas/types/must"
	"github.com/block/ftl/common/schema"
)

var exampleSchema = must.Get(schema.ParseString("schema", `
// Built-in types for FTL.
builtin module builtin {
  // CatchRequest is a request structure for catch verbs.
  export data CatchRequest<Req> {
    verb builtin.Ref
    request Req
    requestType String
    error String
  }

  export data Empty {
  }

  // FailedEvent is used in dead letter topics.
  export data FailedEvent<Event> {
    event Event
    error String
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

  // A reference to a declaration in the schema.
  export data Ref {
    module String
    name String
  }
}

module http {
  export data ApiError {
    message String +alias json "message"
  }

  export data GetPathParams {
    name String +alias json "name"
  }

  export data GetQueryParams {
    age Int? +alias json "age"
  }

  export data GetResponse {
    name String +alias json "name"
    age Int? +alias json "age"
  }

  export data PostRequest {
    name String +alias json "name"
    age Int +alias json "age"
  }

  export data PostResponse {
    name String +alias json "name"
    age Int +alias json "age"
  }

  export verb get(builtin.HttpRequest<Unit, http.GetPathParams, http.GetQueryParams>) builtin.HttpResponse<http.GetResponse, http.ApiError>  
    +ingress http GET /get/{name}

  export verb post(builtin.HttpRequest<http.PostRequest, Unit, Unit>) builtin.HttpResponse<http.PostResponse, http.ApiError>  
    +ingress http POST /post
}

module time {
  export data TimeRequest {
  }

  export data TimeResponse {
    time Time
  }

  verb internal(time.TimeRequest) time.TimeResponse

  // Time returns the current time.
  export verb time(time.TimeRequest) time.TimeResponse  
    +calls time.internal
}

module echo {
  config default String

  // An echo request.
  export data EchoRequest {
    name String? +alias json "name"
  }

  export data EchoResponse {
    message String +alias json "message"
  }

  // Echo returns a greeting with the current time.
  export verb echo(echo.EchoRequest) echo.EchoResponse  
    +calls time.time
    +config echo.default
}
`))

type schemaExampleCmd struct {
}

func (c *schemaExampleCmd) Run() error {
	fmt.Print(exampleSchema)
	return nil
}
