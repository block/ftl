package ingress

import (
	"encoding/json"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/block/ftl/common/schema"
)

func TestOpenAPI(t *testing.T) {
	s, err := schema.ParseString("ingress", `
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

			// Example usage of path and query params
			// curl http://localhost:8891/get/wicket?age=10
			export verb get(builtin.HttpRequest<Unit, http.GetPathParams, http.GetQueryParams>) builtin.HttpResponse<http.GetResponse, http.ApiError>
				+ingress http GET /get/{name}

			// Example POST request with a JSON body
			// curl -X POST http://localhost:8891/post -d '{"name": "wicket", "age": 10}'
			export verb post(builtin.HttpRequest<http.PostRequest, Unit, Unit>) builtin.HttpResponse<http.PostResponse, http.ApiError> 
				+ingress http POST /post

			// Posts a string and returns a string
			export verb postStrings(builtin.HttpRequest<String, Unit, Unit>) builtin.HttpResponse<String, http.ApiError>
				+ingress http POST /post/name
		}
	`)
	assert.NoError(t, err)

	// Convert the schema to OpenAPI
	swagger, err := SchemaToOpenAPI(s)
	assert.NoError(t, err)

	// Convert the OpenAPI spec to JSON
	actualJSON, err := json.MarshalIndent(swagger, "", "  ")
	assert.NoError(t, err)

	// Define the expected JSON
	expectedJSON := `{
  "swagger": "2.0",
  "info": {
    "description": "API generated from FTL schema",
    "title": "FTL API",
    "version": "1.0.0"
  },
  "paths": {
    "/get/{name}": {
      "get": {
        "description": "Example usage of path and query params\ncurl http://localhost:8891/get/wicket?age=10",
        "consumes": [
          "application/json"
        ],
        "produces": [
          "application/json"
        ],
        "operationId": "http.get",
        "parameters": [
          {
            "name": "name",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string"
            }
          },
          {
            "name": "age",
            "in": "query",
            "schema": {
              "type": "integer",
              "format": "int64"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Successful response",
            "schema": {
              "$ref": "#/definitions/http.GetResponse"
            }
          },
          "400": {
            "description": "Error response",
            "schema": {
              "$ref": "#/definitions/http.ApiError"
            }
          }
        }
      }
    },
    "/post": {
      "post": {
        "description": "Example POST request with a JSON body\ncurl -X POST http://localhost:8891/post -d '{\"name\": \"wicket\", \"age\": 10}'",
        "consumes": [
          "application/json"
        ],
        "produces": [
          "application/json"
        ],
        "operationId": "http.post",
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/http.PostRequest"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Successful response",
            "schema": {
              "$ref": "#/definitions/http.PostResponse"
            }
          },
          "400": {
            "description": "Error response",
            "schema": {
              "$ref": "#/definitions/http.ApiError"
            }
          }
        }
      }
    },
    "/post/name": {
      "post": {
        "description": "Posts a string and returns a string",
        "consumes": [
          "application/json"
        ],
        "produces": [
          "application/json"
        ],
        "operationId": "http.postStrings",
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "type": "string"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Successful response",
            "schema": {
              "type": "string"
            }
          },
          "400": {
            "description": "Error response",
            "schema": {
              "$ref": "#/definitions/http.ApiError"
            }
          }
        }
      }
    }
  },
  "definitions": {
    "http.ApiError": {
      "type": "object",
      "required": [
        "message"
      ],
      "properties": {
        "message": {
          "type": "string"
        }
      }
    },
    "http.GetResponse": {
      "type": "object",
      "required": [
        "name"
      ],
      "properties": {
        "age": {
          "type": "integer",
          "format": "int64"
        },
        "name": {
          "type": "string"
        }
      }
    },
    "http.PostRequest": {
      "type": "object",
      "required": [
        "name",
        "age"
      ],
      "properties": {
        "age": {
          "type": "integer",
          "format": "int64"
        },
        "name": {
          "type": "string"
        }
      }
    },
    "http.PostResponse": {
      "type": "object",
      "required": [
        "name",
        "age"
      ],
      "properties": {
        "age": {
          "type": "integer",
          "format": "int64"
        },
        "name": {
          "type": "string"
        }
      }
    }
  }
}`

	assert.Equal(t, expectedJSON, string(actualJSON))
}
