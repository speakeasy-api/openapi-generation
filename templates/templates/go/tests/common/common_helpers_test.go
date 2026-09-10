package tests

import (
	"context"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
)

// Test service URLs - read from environment with defaults for local development

//nolint:nolintlint,unused // usage is injected by internal testing.
func getHTTPBinPort() string {
	if port := os.Getenv("HTTPBIN_PORT"); port != "" {
		return port
	}
	return "35123"
}

//nolint:nolintlint,unused // usage is injected by internal testing.
func getAPITestServicePort() string {
	if port := os.Getenv("API_TEST_SERVICE_PORT"); port != "" {
		return port
	}
	return "35456"
}

//nolint:nolintlint,unused // usage is injected by internal testing.
func getHTTPBinURL() string {
	return "http://localhost:" + getHTTPBinPort()
}

//nolint:nolintlint,unused // usage is injected by internal testing.
func getAPITestServiceURL() string {
	return "http://localhost:" + getAPITestServicePort()
}

//nolint:nolintlint,unused // usage is injected by internal testing.
func recordTest(id string) {
	f, err := os.OpenFile("test-go-record.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	if _, err = f.WriteString(id + "\n"); err != nil {
		log.Println(err)
	}
}

//nolint:nolintlint,unused // usage is injected by internal testing.
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

//nolint:nolintlint,unused // usage is injected by internal testing.
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type RequestLogEntry struct {
	RequestURI  string
	RequestBody string
	StatusCode  int
}

type RequestRecorderClient struct {
	Log []RequestLogEntry
	mu  sync.Mutex
}

func NewRequestRecorderClient() *RequestRecorderClient {
	return &RequestRecorderClient{
		Log: make([]RequestLogEntry, 0),
	}
}

func (r *RequestRecorderClient) Do(req *http.Request) (*http.Response, error) {
	var requestBody string
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			requestBody = string(bodyBytes)
		}
		// Reset body for actual request
		req.Body = io.NopCloser(strings.NewReader(requestBody))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	}

	// Record the request
	r.mu.Lock()
	r.Log = append(r.Log, RequestLogEntry{
		RequestURI:  req.URL.String(),
		RequestBody: requestBody,
		StatusCode:  resp.StatusCode,
	})
	r.mu.Unlock()

	return resp, err
}

func (r *RequestRecorderClient) DoWithContext(ctx context.Context, req *http.Request) (*http.Response, error) {
	return r.Do(req.WithContext(ctx))
}
