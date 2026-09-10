package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/acceptHeaders"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/clientcredentials"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/delay"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/ecommerce"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/errors"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/eventstreams"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/jsonLines"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/method"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/middleware"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/optionalNullableFields"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/pagination"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/polling"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/proxy"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/readonlywriteonly"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/redirects"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/reflect"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/responseBodies"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/responseHeaders"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/retries"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/smartunion"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/xNdJson"

	"github.com/gorilla/mux"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/auth"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/requestbody"
)

var bindArg = flag.String("b", ":8080", "Bind address")

func main() {
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/oauth2/token", auth.HandleOAuth2InspectToken).Methods(http.MethodGet)
	r.HandleFunc("/oauth2/token", auth.HandleOAuth2).Methods(http.MethodPost)
	r.HandleFunc("/auth/token", auth.HandleAuthInspectToken).Methods(http.MethodGet)
	r.HandleFunc("/auth", auth.HandleAuth).Methods(http.MethodPost)
	r.HandleFunc("/auth/customsecurity/{customSchemeType}", auth.HandleCustomAuth).Methods(http.MethodGet)
	r.HandleFunc("/delay/{seconds}", delay.HandleDelay).Methods(http.MethodGet)
	r.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	}).Methods(http.MethodGet)
	r.HandleFunc("/requestbody", requestbody.HandleRequestBody).Methods(http.MethodPost)
	r.HandleFunc("/requestbody/multipart-form/files", requestbody.HandleMultipartFormFiles).Methods(http.MethodPost)
	r.HandleFunc("/response-headers/{statusCode}", responseHeaders.HandleResponseHeaders).Methods(http.MethodGet)
	r.HandleFunc("/error-response-headers/{statusCode}", responseHeaders.HandleErrorResponseHeaders).Methods(http.MethodGet)
	r.HandleFunc("/vendorjson", responseHeaders.HandleVendorJsonResponseHeaders).Methods(http.MethodGet)
	r.HandleFunc("/pagination/limitoffset/page", pagination.HandleLimitOffsetPage).Methods(http.MethodGet, http.MethodPut)
	r.HandleFunc("/pagination/limitoffset/deep_outputs/page", pagination.HandleLimitOffsetDeepOutputsPage).Methods(http.MethodGet, http.MethodPut)
	r.HandleFunc("/pagination/limitoffset/offset", pagination.HandleLimitOffsetOffset).Methods(http.MethodGet, http.MethodPut)
	r.HandleFunc("/pagination/cursor", pagination.HandleCursor).Methods(http.MethodGet, http.MethodPut)
	r.HandleFunc("/pagination/cursor/response_envelope", pagination.HandleCursorResponseEnvelope).Methods(http.MethodGet, http.MethodPut)
	r.HandleFunc("/pagination/url", pagination.HandleURL).Methods(http.MethodGet)
	r.HandleFunc("/pagination/cursor_non_numeric", pagination.HandleNonNumericCursor).Methods(http.MethodGet)
	r.HandleFunc("/pagination/cursor/nullable", pagination.HandleCursorNullable).Methods(http.MethodGet)
	r.HandleFunc("/pagination/cursor/deep_nested_outputs", pagination.HandleCursorDeepNestedOutputs).Methods(http.MethodGet)
	r.HandleFunc("/pagination/limitoffset/nullable_offset", pagination.HandleLimitOffsetNullable).Methods(http.MethodGet)
	r.HandleFunc("/pagination/limitoffset/nullable_body", pagination.HandleLimitOffsetPageNullableBody).Methods(http.MethodPut)
	r.HandleFunc("/retries", retries.HandleRetries).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/errors/errorUnion/malformed", errors.HandleMalformedErrorUnionResponse).Methods(http.MethodGet)
	r.HandleFunc("/errors/{status_code}", errors.HandleErrors).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/errors/{status_code}/malformed/{shape}", errors.HandleMalformedErrors).Methods(http.MethodGet)
	r.HandleFunc("/optional", acceptHeaders.HandleAcceptHeaderMultiplexing).Methods(http.MethodGet)
	r.HandleFunc("/readonlyorwriteonly", readonlywriteonly.HandleReadOrWrite).Methods(http.MethodPost)
	r.HandleFunc("/readonlyandwriteonly", readonlywriteonly.HandleReadAndWrite).Methods(http.MethodPost)
	r.HandleFunc("/writeonlyoutput", readonlywriteonly.HandleWriteOnlyOutput).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/json", eventstreams.HandleEventStreamJSON).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/text", eventstreams.HandleEventStreamText).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/malformed", eventstreams.HandleEventStreamMalformed).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/multiline", eventstreams.HandleEventStreamMultiLine).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/rich", eventstreams.HandleEventStreamRich).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/chat", eventstreams.HandleEventStreamChat).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/sse-overload", eventstreams.HandleEventStreamChatOverload).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/chat-flat", eventstreams.HandleEventStreamChatFlatten).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/chat-chunked", eventstreams.HandleEventStreamChat).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/chat-heartbeat", eventstreams.HandleEventStreamChatHeartbeat).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/large-event-small-chunks", eventstreams.HandleEventStreamLargeEvent).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/split-boundaries", eventstreams.HandleEventStreamSplitBoundaries).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/differentdataschemas", eventstreams.HandleEventStreamDifferentDataSchemas).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/stayopen", eventstreams.HandleEventStreamStayOpen).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/union-with-comments", eventstreams.HandleEventStreamUnionWithComments).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/partial-with-comments", eventstreams.HandleEventStreamPartialWithComments).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/optionaldata", eventstreams.HandleEventStreamOptionalData).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/mixeddata", eventstreams.HandleEventStreamMixedData).Methods(http.MethodPost)
	r.HandleFunc("/eventstreams/wpt-compliance", eventstreams.HandleEventStreamWPTCompliance).Methods(http.MethodGet)
	r.HandleFunc("/jsonl", jsonLines.HandleJSONLinesRich).Methods(http.MethodGet)
	r.HandleFunc("/jsonl/deserialization_verification", jsonLines.HandleJsonLinesDeserializationVerification).Methods(http.MethodGet)
	r.HandleFunc("/jsonl/chunks", jsonLines.HandleJSONLinesChunksRich).Methods(http.MethodGet)
	r.HandleFunc("/jsonl/hold", jsonLines.HandleJSONLinesHold).Methods(http.MethodGet)
	r.HandleFunc("/x-ndjson", xNdJson.HandleXNdJsonLinesRich).Methods(http.MethodGet)
	r.HandleFunc("/x-ndjson/chunks", xNdJson.HandleXNdJsonLinesChunksRich).Methods(http.MethodGet)
	r.HandleFunc("/clientcredentials/token", clientcredentials.HandleTokenRequest).Methods(http.MethodPost)
	r.HandleFunc("/clientcredentials/authenticatedrequest", clientcredentials.HandleAuthenticatedRequest).Methods(http.MethodPost)
	r.HandleFunc("/clientcredentials/alt/token", clientcredentials.HandleAltTokenRequest).Methods(http.MethodPost)
	r.HandleFunc("/clientcredentials/alt/authenticatedrequest", clientcredentials.HandleAuthenticatedRequest).Methods(http.MethodPost)
	r.HandleFunc("/reflect", reflect.HandleReflect).Methods(http.MethodPost)
	r.HandleFunc("/html", responseBodies.HandleHtmlGet).Methods(http.MethodGet)
	r.HandleFunc("/xml", responseBodies.HandleXmlGet).Methods(http.MethodGet)
	r.HandleFunc("/smartunion/nestedunion", smartunion.HandleNestedUnion).Methods(http.MethodPost)
	r.HandleFunc("/method/delete", method.HandleDelete).Methods(http.MethodDelete)
	r.HandleFunc("/method/get", method.HandleGet).Methods(http.MethodGet)
	r.HandleFunc("/method/head", method.HandleHead).Methods(http.MethodHead)
	r.HandleFunc("/method/options", method.HandleOptions).Methods(http.MethodOptions)
	r.HandleFunc("/method/patch", method.HandlePatch).Methods(http.MethodPatch)
	r.HandleFunc("/method/post", method.HandlePost).Methods(http.MethodPost)
	r.HandleFunc("/method/put", method.HandlePut).Methods(http.MethodPut)
	r.HandleFunc("/method/trace", method.HandleTrace).Methods(http.MethodTrace)
	r.HandleFunc("/followRedirect/oldPage", redirects.HandleRedirectOldPage).Methods(http.MethodGet)
	r.HandleFunc("/followRedirect/newPage", redirects.HandleRedirectNewPage).Methods(http.MethodGet)
	r.HandleFunc("/responseObjectWithOptionalTrueNullableTrueFieldDeserializes", optionalNullableFields.HandleObjectWithOptionalTrueNullableTrue).Methods(http.MethodGet)
	r.HandleFunc("/responseObjectWithOptionalFalseNullableTrueFieldDeserializes", optionalNullableFields.HandleObjectWithOptionalFalseNullableTrue).Methods(http.MethodGet)
	r.HandleFunc("/responseObjectWithOptionalTrueNullableFalseFieldDeserializes", optionalNullableFields.HandleObjectWithOptionalTrueNullableFalse).Methods(http.MethodGet)
	r.HandleFunc("/responseObjectWithOptionalFalseNullableFalseFieldDeserializes", optionalNullableFields.HandleObjectWithOptionalFalseNullableFalse).Methods(http.MethodGet)

	// Polling endpoints
	r.HandleFunc("/polling/delaySeconds", polling.DelaySecondsGetHandler).Methods(http.MethodGet)
	r.HandleFunc("/polling/failureCriteria/responseBody", polling.FailureCriteriaResponseBodyGetHandler).Methods(http.MethodGet)
	r.HandleFunc("/polling/failureCriteria/responseBody/nameOverride", polling.FailureCriteriaResponseBodyNameOverrideGetHandler).Methods(http.MethodGet)
	r.HandleFunc("/polling/failureCriteria/responseBody/nested", polling.FailureCriteriaResponseBodyNestedGetHandler).Methods(http.MethodGet)
	r.HandleFunc("/polling/failureCriteria/statusCode", polling.FailureCriteriaStatusCodeGetHandler).Methods(http.MethodGet)
	r.HandleFunc("/polling/intervalSeconds", polling.IntervalSecondsGetHandler).Methods(http.MethodGet)
	r.HandleFunc("/polling/limitCount", polling.LimitCountGetHandler).Methods(http.MethodGet)
	r.HandleFunc("/polling/successCriteria/responseBody", polling.SuccessCriteriaResponseBodyGetHandler).Methods(http.MethodGet)
	r.HandleFunc("/polling/successCriteria/responseBody/nested", polling.SuccessCriteriaResponseBodyNestedGetHandler).Methods(http.MethodGet)
	r.HandleFunc("/polling/successCriteria/statusCode", polling.SuccessCriteriaStatusCodeGetHandler).Methods(http.MethodGet)

	// Ecommerce endpoints with OAuth2 protection
	oauth2router := r.NewRoute().Subrouter()
	oauth2router.Use(middleware.OAuth2)
	oauth2router.HandleFunc("/ecommerce/products", ecommerce.HandleListProducts).Methods(http.MethodGet)
	oauth2router.HandleFunc("/ecommerce/products", ecommerce.HandleCreateProduct).Methods(http.MethodPost)
	oauth2router.HandleFunc("/ecommerce/products/{id}", ecommerce.HandleFetchProduct).Methods(http.MethodGet)
	oauth2router.HandleFunc("/ecommerce/products/{id}", ecommerce.HandleUpdateProduct).Methods(http.MethodPut)
	oauth2router.HandleFunc("/ecommerce/products/{id}", ecommerce.HandleDeleteProduct).Methods(http.MethodDelete)
	oauth2router.HandleFunc("/ecommerce/products/{id}/inventory", ecommerce.HandleUpdateProductStock).Methods(http.MethodPut)

	handler := middleware.Fault(r)
	handler = middleware.Teapot(handler)
	handler = proxy.ProxyHandler(handler)

	bind := ":8080"
	if bindArg != nil {
		bind = *bindArg
		if !strings.HasPrefix(bind, ":") {
			bind = ":" + bind
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go auth.StartTokenDBCompaction(ctx)

	ln, err := net.Listen("tcp", bind)
	if err != nil {
		log.Fatal(err)
	}

	// Emit the actual port for boot-services.sh to parse.
	// Critical when bind=":0" (ephemeral port mode).
	_, port, _ := net.SplitHostPort(ln.Addr().String())
	fmt.Printf("LISTEN_PORT=%s\n", port)

	log.Printf("Listening on :%s\n", port)
	if err := http.Serve(ln, handler); err != nil {
		log.Fatal(err)
	}
}
