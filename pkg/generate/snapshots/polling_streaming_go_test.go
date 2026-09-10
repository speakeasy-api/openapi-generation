package snapshots

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Verifies the cancel lifecycle of polling operations with streaming
// responses: a stream parsed by a non-final poll attempt must be disposed of
// without cancelling the shared timeout context, and the final returned
// stream must own a live cancel that is released on Close.
func TestGoPollingStreamingCancelLifecycle(t *testing.T) {
	t.Parallel()

	spec := `openapi: 3.1.0
info:
  title: Polling Streaming API
  version: 1.0.0
servers:
  - url: https://api.example.com
security:
  - apiKey: []
paths:
  /jobs/results:
    get:
      operationId: pollJobResults
      description: |-
        Polling operation whose partial-progress response (200) and completed
        response (201) are both JSONL streams. Polling continues while the
        server returns 200 partial streams and succeeds on the 201 stream.
      x-speakeasy-polling:
        - name: WaitForCompleted
          intervalSeconds: 1
          limitCount: 5
          successCriteria:
            - condition: $statusCode == 201
      responses:
        "200":
          description: Partial results
          content:
            application/jsonl:
              schema:
                $ref: "#/components/schemas/JobEvent"
        "201":
          description: Completed results
          content:
            application/jsonl:
              schema:
                $ref: "#/components/schemas/JobEvent"
  /jobs/sse:
    get:
      operationId: pollJobSse
      description: |-
        Polling operation whose success response is a server-sent event
        stream; a 202 JSON response means the job is still processing.
      x-speakeasy-polling:
        - name: WaitForReady
          intervalSeconds: 1
          limitCount: 5
          successCriteria:
            - condition: $statusCode == 200
      responses:
        "200":
          description: Event stream
          content:
            text/event-stream:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/SseJobEvent"
        "202":
          description: Still processing
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/JobStatus"
components:
  securitySchemes:
    apiKey:
      type: apiKey
      name: x-api-key
      in: header
  schemas:
    JobEvent:
      type: object
      properties:
        message:
          type: string
    SseJobEvent:
      type: object
      properties:
        data:
          $ref: "#/components/schemas/JobEvent"
    JobStatus:
      type: object
      properties:
        status:
          type: string
`

	genYaml := `go:
  packageName: github.com/example/pollingstreaming
`

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:    spec,
		GenYaml: genYaml,
		AfterGenerate: func(t *testing.T, tempDir string) {
			t.Helper()

			sdk := readGeneratedFile(t, tempDir, "sdk.go")

			// The polling wrapper hands each attempt a per-attempt cancel
			// chained to the shared timeout cancel, so an abandoned non-final
			// stream never takes the shared cancel with it.
			assert.Contains(t, sdk, "sharedStreamCancel := *streamCancel")
			assert.Contains(t, sdk, "attemptCtx, cancel := context.WithCancel(ctx)")
			assert.Contains(t, sdk, "req = req.WithContext(attemptCtx)")
			assert.Contains(t, sdk, "*streamCancel = attemptStreamCancel")

			runPollingStreamingLifecycleContract(t, tempDir)
		},
	})
}

func runPollingStreamingLifecycleContract(t *testing.T, tempDir string) {
	t.Helper()

	path := filepath.Join(tempDir, "polling_streaming_lifecycle_contract_test.go")
	require.NoError(t, os.WriteFile(path, []byte(`package pollingstreaming

import (
    "context"
    "errors"
    "fmt"
    "net/http"
    "net/http/httptest"
    "sync"
    "testing"
    "time"

    "github.com/example/pollingstreaming/models/operations"
    "github.com/example/pollingstreaming/polling"
)

func TestPollingStreamCancelLifecycle(t *testing.T) {
    var mu sync.Mutex
    attempt := 0
    ctxDone := make(map[int]chan struct{})

    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        mu.Lock()
        attempt++
        n := attempt
        done := make(chan struct{})
        ctxDone[n] = done
        mu.Unlock()

        w.Header().Set("Content-Type", "application/jsonl")
        if n < 3 {
            w.WriteHeader(200)
        } else {
            w.WriteHeader(201)
        }
        fmt.Fprintln(w, `+"`"+`{"message":"hello"}`+"`"+`)
        w.(http.Flusher).Flush()

        // Hold the stream open until the client cancels/disconnects. The
        // bound keeps a disposal regression from hanging srv.Close() forever.
        select {
        case <-r.Context().Done():
        case <-time.After(30 * time.Second):
        }
        close(done)
    }))
    defer srv.Close()

    s := New(WithServerURL(srv.URL), WithSecurity("test"))

    res, err := s.PollJobResults(context.Background(),
        operations.WithOperationTimeout(60*time.Second),
        operations.WithPolling(s.PollJobResultsWaitForCompleted()),
    )
    if err != nil {
        t.Fatalf("PollJobResults: %v", err)
    }
    if res.HTTPMeta.Response.StatusCode != 201 {
        t.Fatalf("expected final 201, got %d", res.HTTPMeta.Response.StatusCode)
    }
    stream := res.JobEvent
    if stream == nil {
        t.Fatal("expected final attempt's JSONL stream")
    }

    // Streams parsed by the non-final attempts must be disposed of (their
    // request contexts cancelled, closing the connections) while the poll
    // went on to the final attempt.
    for _, n := range []int{1, 2} {
        mu.Lock()
        done := ctxDone[n]
        mu.Unlock()
        select {
        case <-done:
        case <-time.After(5 * time.Second):
            t.Fatalf("attempt %d stream was never disposed of (leaked body/connection)", n)
        }
    }

    // The final stream must still be alive and readable.
    if !stream.Next() {
        t.Fatalf("final stream read: %v", stream.Err())
    }
    event, err := stream.Value()
    if err != nil {
        t.Fatalf("final stream value: %v", err)
    }
    if event.Message == nil || *event.Message != "hello" {
        t.Fatalf("unexpected final stream payload: %+v", event)
    }
    mu.Lock()
    finalDone := ctxDone[3]
    mu.Unlock()
    // Give a prematurely fired cancel (the old defer-on-return behaviour)
    // time to reach the server before asserting the context is still live.
    select {
    case <-finalDone:
        t.Fatal("final attempt's context was cancelled before Close")
    case <-time.After(200 * time.Millisecond):
    }

    // Closing the final stream must fire its owned cancel, releasing the
    // request/timeout contexts well before the 60s operation timeout
    // (observable as the server seeing the request context done).
    if err := stream.Close(); err != nil {
        t.Fatalf("close: %v", err)
    }
    select {
    case <-finalDone:
    case <-time.After(5 * time.Second):
        t.Fatal("closing the final stream did not release its cancel/context")
    }
}

// When the poll limit is exhausted the wrapper returns the last attempt's
// response with polling.LimitCountError. That attempt's stream owns the
// cancel hand-off, so the wrapper must release it on the error path: the
// caller only holds an error and nothing else would close the stream before
// the operation timeout.
func TestPollingStreamLimitExhausted(t *testing.T) {
    var mu sync.Mutex
    attempt := 0
    ctxDone := make(map[int]chan struct{})

    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        mu.Lock()
        attempt++
        n := attempt
        done := make(chan struct{})
        ctxDone[n] = done
        mu.Unlock()

        w.Header().Set("Content-Type", "application/jsonl")
        w.WriteHeader(200)
        fmt.Fprintln(w, `+"`"+`{"message":"hello"}`+"`"+`)
        w.(http.Flusher).Flush()

        select {
        case <-r.Context().Done():
        case <-time.After(30 * time.Second):
        }
        close(done)
    }))
    defer srv.Close()

    s := New(WithServerURL(srv.URL), WithSecurity("test"))

    res, err := s.PollJobResults(context.Background(),
        operations.WithOperationTimeout(60*time.Second),
        operations.WithPolling(s.PollJobResultsWaitForCompleted(), polling.WithLimitCountOverride(2)),
    )
    var limitErr *polling.LimitCountError
    if !errors.As(err, &limitErr) {
        t.Fatalf("expected polling.LimitCountError, got %v", err)
    }
    if res == nil {
        t.Fatal("expected the last attempt's response alongside the error")
    }
    mu.Lock()
    total := attempt
    mu.Unlock()
    if total != 2 {
        t.Fatalf("expected 2 attempts, got %d", total)
    }

    // Every attempt's request context, the final one included, must be
    // released shortly after the wrapper returns.
    for n := 1; n <= 2; n++ {
        mu.Lock()
        done := ctxDone[n]
        mu.Unlock()
        select {
        case <-done:
        case <-time.After(5 * time.Second):
            t.Fatalf("attempt %d stream was not released after limit exhaustion", n)
        }
    }
}
`), 0o644))
	t.Cleanup(func() { _ = os.Remove(path) })

	cmd := exec.Command("go", "test", "-count=1", "-run", "TestPollingStream", ".")
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	require.NoError(t, os.Remove(path))
}
