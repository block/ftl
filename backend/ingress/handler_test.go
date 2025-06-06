package ingress

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/encoding"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/timelineclient"
)

func TestIngress(t *testing.T) {
	sch, err := schema.ParseString("", `
	realm test {
		module test {
			export data AliasRequest {
				aliased String +alias json "alias"
			}

			export data PathParameterRequest {
				username String
			}

			export data MissingTypes {
				optional String?
				array [String]
				map {String: String}
				any Any
				unit Unit
			}

			export data JsonPayload {
				foo String
			}

			export verb getAlias(HttpRequest<Unit, Unit, test.AliasRequest>) HttpResponse<Empty, Empty>
				+ingress http GET /getAlias

			export verb getPath(HttpRequest<Unit, test.PathParameterRequest, Unit>) HttpResponse<Empty, Empty>
				+ingress http GET /getPath/{username}

			export verb postMissingTypes(HttpRequest<test.MissingTypes, Unit, Unit>) HttpResponse<Empty, Empty>
				+ingress http POST /postMissingTypes

			export verb postJsonPayload(HttpRequest<test.JsonPayload, Unit, Unit>) HttpResponse<Empty, Empty>
				+ingress http POST /postJsonPayload
		}
	}
	`)
	assert.NoError(t, err)

	routes := []ingressRoute{
		{path: "/getAlias", module: "test", verb: "getAlias"},
		{path: "/getPath/{username}", module: "test", verb: "getPath"},
		{path: "/postMissingTypes", module: "test", verb: "postMissingTypes"},
		{path: "/postJsonPayload", module: "test", verb: "postJsonPayload"},
	}

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	for _, test := range []struct {
		name       string
		method     string
		path       string
		query      url.Values
		payload    []byte
		response   optional.Option[HTTPResponse]
		statusCode int
	}{
		{name: "InvalidRoute",
			method:     "GET",
			path:       "/invalid",
			statusCode: http.StatusNotFound},
		{name: "GetAlias",
			method:     "GET",
			path:       "/getAlias",
			query:      url.Values{"alias": {"value"}},
			response:   optional.Some(HTTPResponse{Body: []byte(`{}`)}),
			statusCode: http.StatusOK},
	} {
		t.Run(test.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			rec.Body = &bytes.Buffer{}
			var response HTTPResponse
			var ok bool
			if response, ok = test.response.Get(); ok {
				response = HTTPResponse{Body: []byte(`{}`)}
			}
			req := httptest.NewRequest(test.method, test.path, bytes.NewBuffer(test.payload)).WithContext(ctx)
			req.URL.RawQuery = test.query.Encode()
			reqKey := key.NewRequestKey(key.OriginIngress, "test")
			assert.NoError(t, err)
			fv := &fakeVerbClient{response: response, t: t}

			eventSource := schemaeventsource.NewUnattached()
			testModule := &schema.Module{
				Name: "test",
				Runtime: &schema.ModuleRuntime{
					Deployment: &schema.ModuleRuntimeDeployment{
						DeploymentKey: key.NewDeploymentKey("test", "test"),
					},
					Runner: &schema.ModuleRuntimeRunner{
						Endpoint: "http://localhost:8080",
					},
				},
			}
			// Publish the test module to the event source
			assert.NoError(t, eventSource.PublishModuleForTest(testModule))

			svc := &service{
				view:           syncView(ctx, eventSource),
				client:         fv,
				timelineClient: timelineclient.NewClient(ctx, timelineclient.NullConfig),
				routeTable:     routing.New(ctx, eventSource),
			}
			svc.handleHTTP(time.Now(), sch, reqKey, routes, rec, req, fv)
			result := rec.Result()
			defer result.Body.Close()
			assert.Equal(t, test.statusCode, rec.Code, "%s: %s", result.Status, rec.Body.Bytes())
			if rec.Code >= 300 {
				return
			}
			assert.Equal(t, response.Body, rec.Body.Bytes())
		})
	}
}

type fakeVerbClient struct {
	response HTTPResponse
	t        *testing.T
}

func (r *fakeVerbClient) Call(ctx context.Context, req *connect.Request[ftlv1.CallRequest]) (*connect.Response[ftlv1.CallResponse], error) {
	body, err := encoding.Marshal(r.response)
	assert.NoError(r.t, err)
	return connect.NewResponse(&ftlv1.CallResponse{Response: &ftlv1.CallResponse_Body{Body: body}}), nil
}
