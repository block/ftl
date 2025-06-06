package ingress

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/rpc/headers"
	"github.com/block/ftl/internal/timelineclient"
)

// handleHTTP HTTP ingress routes.
func (s *service) handleHTTP(startTime time.Time, sch *schema.Schema, requestKey key.Request, routesForMethod []ingressRoute, w http.ResponseWriter, r *http.Request, client routing.CallClient) {
	ctx := rpc.WithRequestKey(r.Context(), requestKey)
	logger := log.FromContext(ctx).Scope(fmt.Sprintf("ingress:%s:%s", r.Method, r.URL.Path))
	logger.Debugf("Start ingress request")

	if ServeOpenAPI(sch, w, r) {
		return
	}

	routeOpt := getIngressRoute(routesForMethod, r.URL.Path)
	var route *ingressRoute
	var ok bool
	if route, ok = routeOpt.Get(); !ok {
		http.NotFound(w, r)
		metrics.Request(ctx, r.Method, r.URL.Path, optional.None[*schemapb.Ref](), startTime, optional.Some("route not found"))
		return
	}
	logger = logger.Module(route.module)

	verbRef := &schemapb.Ref{Module: route.module, Name: route.verb}

	deploymentKey, ok := s.routeTable.Current().GetDeployment(route.module).Get()
	if !ok {
		http.Error(w, "deployment not found", http.StatusInternalServerError)
		metrics.Request(ctx, r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.Some("deployment not found"))
		return
	}

	ingressEvent := timelineclient.Ingress{
		RequestKey:      requestKey,
		StartTime:       startTime,
		DeploymentKey:   deploymentKey,
		Verb:            &schema.Ref{Name: route.verb, Module: route.module},
		RequestMethod:   r.Method,
		RequestPath:     r.URL.Path,
		RequestHeaders:  r.Header.Clone(),
		ResponseHeaders: make(http.Header),
	}

	body, err := buildRequestBody(route, r, sch)
	if err != nil {
		// Only log at debug, as this is a client side error
		logger.Debugf("bad request: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		metrics.Request(ctx, r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.Some("bad request"))
		s.recordIngressErrorEvent(ctx, ingressEvent, http.StatusBadRequest, err.Error())
		return
	}
	ingressEvent.RequestBody = body

	creq := connect.NewRequest(&ftlv1.CallRequest{
		Metadata: &ftlv1.Metadata{},
		Verb:     verbRef,
		Body:     body,
	})
	headers.SetRequestKey(creq.Header(), requestKey)

	resp, err := client.Call(ctx, creq)
	if err != nil {
		logger.Errorf(err, "failed to call verb")
		if connectErr := new(connect.Error); errors.As(err, &connectErr) {
			httpCode := connectCodeToHTTP(connectErr.Code())
			http.Error(w, http.StatusText(httpCode), httpCode)
			metrics.Request(ctx, r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.Some("failed to call verb: connect error"))
			s.recordIngressErrorEvent(ctx, ingressEvent, http.StatusInternalServerError, connectErr.Error())
		} else {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			metrics.Request(ctx, r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.Some("failed to call verb: internal server error"))
			s.recordIngressErrorEvent(ctx, ingressEvent, http.StatusInternalServerError, err.Error())
		}
		return
	}
	switch msg := resp.Msg.Response.(type) {
	case *ftlv1.CallResponse_Body:
		verb := &schema.Verb{}
		err = sch.ResolveToType(&schema.Ref{Name: route.verb, Module: route.module}, verb)
		if err != nil {
			logger.Errorf(err, "could not resolve schema type for verb %s", route.verb)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			metrics.Request(ctx, r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.Some("could not resolve schema type for verb"))
			s.recordIngressErrorEvent(ctx, ingressEvent, http.StatusInternalServerError, err.Error())
			return
		}
		var responseBody []byte
		var rawBody []byte
		if metadata, ok := verb.GetMetadataIngress().Get(); ok && metadata.Type == "http" {
			var response HTTPResponse
			if err := json.Unmarshal(msg.Body, &response); err != nil {
				logger.Errorf(err, "could not unmarhal response for verb %s", verb)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				metrics.Request(ctx, r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.Some("could not unmarhal response for verb"))
				s.recordIngressErrorEvent(ctx, ingressEvent, http.StatusInternalServerError, err.Error())
				return
			}
			rawBody = response.Body
			var responseHeaders http.Header
			responseBody, responseHeaders, err = ResponseForVerb(sch, verb, response)
			if err != nil {
				logger.Errorf(err, "could not create response for verb %s", verb.Name)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				metrics.Request(ctx, r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.Some("could not create response for verb"))
				s.recordIngressErrorEvent(ctx, ingressEvent, http.StatusInternalServerError, err.Error())
				return
			}

			for k, v := range responseHeaders {
				w.Header()[k] = v
				ingressEvent.ResponseHeaders.Set(k, v[0])
			}

			statusCode := http.StatusOK

			// Override with status from verb if provided
			if response.Status != 0 {
				statusCode = response.Status
				w.WriteHeader(statusCode)
			}

			ingressEvent.ResponseStatus = statusCode
		} else {
			w.WriteHeader(http.StatusOK)
			ingressEvent.ResponseStatus = http.StatusOK
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			ingressEvent.ResponseHeaders.Set("Content-Type", "application/json; charset=utf-8")
			responseBody = msg.Body
			rawBody = responseBody
		}
		ingressEvent.ResponseBody = rawBody
		_, err = w.Write(responseBody)
		if err == nil {
			metrics.Request(ctx, r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.None[string]())
			s.timelineClient.Publish(ctx, ingressEvent)
		} else {
			logger.Errorf(err, "could not write response body")
			metrics.Request(ctx, r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.Some("could not write response body"))
			s.recordIngressErrorEvent(ctx, ingressEvent, http.StatusInternalServerError, err.Error())
		}

	case *ftlv1.CallResponse_Error_:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		metrics.Request(ctx, r.Method, r.URL.Path, optional.Some(verbRef), startTime, optional.Some("call response: internal server error"))
		s.recordIngressErrorEvent(ctx, ingressEvent, http.StatusInternalServerError, msg.Error.Message)
	}
}

func (s *service) recordIngressErrorEvent(
	ctx context.Context,
	ingressEvent timelineclient.Ingress,
	statusCode int,
	errorMsg string,
) {
	ingressEvent.ResponseStatus = statusCode
	ingressEvent.Error = optional.Some(errorMsg)
	s.timelineClient.Publish(ctx, ingressEvent)
}

// Copied from the Apache-licensed connect-go source.
func connectCodeToHTTP(code connect.Code) int {
	switch code {
	case connect.CodeCanceled:
		return 408
	case connect.CodeUnknown:
		return 500
	case connect.CodeInvalidArgument:
		return 400
	case connect.CodeDeadlineExceeded:
		return 408
	case connect.CodeNotFound:
		return 404
	case connect.CodeAlreadyExists:
		return 409
	case connect.CodePermissionDenied:
		return 403
	case connect.CodeResourceExhausted:
		return 429
	case connect.CodeFailedPrecondition:
		return 412
	case connect.CodeAborted:
		return 409
	case connect.CodeOutOfRange:
		return 400
	case connect.CodeUnimplemented:
		return 404
	case connect.CodeInternal:
		return 500
	case connect.CodeUnavailable:
		return 503
	case connect.CodeDataLoss:
		return 500
	case connect.CodeUnauthenticated:
		return 401
	default:
		return 500 // same as CodeUnknown
	}
}
