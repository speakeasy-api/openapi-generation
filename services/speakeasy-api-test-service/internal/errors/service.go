package errors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/internal/utils"

	"github.com/speakeasy-api/openapi-generation/services/speakeasy-api-test-service/pkg/models"
)

// HandleErrors serves the /errors/{status_code} endpoint, always responding with
// the requested HTTP status code.
//
// POST: the request body is echoed back verbatim as the response body.
// GET: canonical {message, code, type} error body is returned.
// When the request carries `?cause=true`, an additional error message
// is nested inside an optional+nullable `cause` object.
func HandleErrors(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	statusCode, ok := vars["status_code"]
	if !ok {
		utils.HandleError(w, fmt.Errorf("status_code is required"))
		return
	}

	statusCodeInt, err := strconv.Atoi(statusCode)
	if err != nil {
		utils.HandleError(w, fmt.Errorf("status_code must be an integer"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCodeInt)

	var res interface{}
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			utils.HandleError(w, err)
			return
		}

		if err := json.Unmarshal(body, &res); err != nil {
			utils.HandleError(w, err)
			return
		}
	} else {
		errBody := models.Error{
			Code:    statusCode,
			Message: "an error occurred",
			Type:    "internal",
		}
		if r.URL.Query().Get("cause") == "true" {
			errBody.Cause = &models.ErrorCause{Message: "underlying cause"}
		}
		res = errBody
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		utils.HandleError(w, err)
		return
	}
}

type MalformedRequest struct {
	StatusCode int    `path:"status_code"`
	Shape      string `path:"shape"`
}

var malformedBodies = map[string][]byte{
	"sse":            []byte("event: error\ndata: {\"error\":{\"message\":\"forced sse\",\"code\":\"too_many\"},\"event_type\":\"error\"}\n\n"),
	"plain":          []byte("internal server timeout"),
	"html":           []byte("<html>502 Bad Gateway</html>"),
	"malformed-json": []byte(`{"error":`),
	"empty":          nil,
	// Valid JSON that violates the declared 200 response schema (required
	// integer field `count`). Used with statusCode=200 to exercise the
	// responseSchemaValidation generator flag on the success path.
	"wrong-type":       []byte(`{"count":"not-an-int"}`),
	"missing-required": []byte(`{}`),
	// Valid JSON error body that violates the declared error schema: the typed
	// string field `code` is an object. Used with 4XX/5XX to exercise lenient
	// error handling that best-effort constructs the typed error.
	"wrong-type-error": []byte(`{"code":{"invalid":true},"message":"degraded error"}`),
}

func HandleMalformedErrors(w http.ResponseWriter, r *http.Request) {
	var req MalformedRequest
	if err := decodePathVars(mux.Vars(r), &req); err != nil {
		utils.HandleError(w, err)
		return
	}

	body, ok := malformedBodies[req.Shape]
	if !ok {
		utils.HandleError(w, fmt.Errorf("unknown shape: %s", req.Shape))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(req.StatusCode)
	if body != nil {
		_, _ = w.Write(body)
	}
}

// HandleMalformedErrorUnionResponse serves an error status whose body is a
// discriminated error union. The discriminator `tag` selects the known, mapped
// variant tag1 (taggedError1, whose `error` is declared a string), but the body
// carries `error` as an object, so the mapped variant fails schema validation.
func HandleMalformedErrorUnionResponse(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(`{"tag":"tag1","error":{"unexpected":"object"}}`))
}

func decodePathVars(vars map[string]string, dst any) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("decodePathVars: dst must be a pointer to struct")
	}
	s := v.Elem()
	t := s.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("path")
		if tag == "" {
			continue
		}
		raw, ok := vars[tag]
		if !ok {
			return fmt.Errorf("path parameter %q is required", tag)
		}
		switch field.Type.Kind() {
		case reflect.String:
			s.Field(i).SetString(raw)
		case reflect.Int:
			n, err := strconv.Atoi(raw)
			if err != nil {
				return fmt.Errorf("path parameter %q must be an integer", tag)
			}
			s.Field(i).SetInt(int64(n))
		default:
			return fmt.Errorf("decodePathVars: unsupported field type %s", field.Type.Kind())
		}
	}
	return nil
}
