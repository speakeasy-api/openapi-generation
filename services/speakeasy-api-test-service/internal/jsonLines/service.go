package jsonLines

import (
	"fmt"
	"net/http"
	"time"
)

func pushEvents(rw http.ResponseWriter, events [][]string) {
	for _, event := range events {
		for _, line := range event {
			fmt.Fprint(rw, line)
		}

		if f, ok := rw.(http.Flusher); ok {
			f.Flush()
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func pushChunks(rw http.ResponseWriter, chunks []string) {
	for _, chunk := range chunks {
		fmt.Fprint(rw, chunk)

		if f, ok := rw.(http.Flusher); ok {
			f.Flush()
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func HandleJSONLinesChunksRich(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "application/jsonl")

	pushChunks(rw, []string{
		"{\"name\": \"Peter\", \"skills\": [\"Go\"",
		", \"Python\"]}\n{\"name\": \"John\"",
		", \"skills\": [\"Go\", \"Rust\"]}\n",
	})
}

func HandleJSONLinesRich(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "application/jsonl")

	pushEvents(rw, [][]string{
		{
			"{\"name\": \"Peter\", \"skills\": [\"Go\", \"Python\"]}\n",
		},
		{
			"{\"name\": \"John\", \"skills\": [\"Go\", \"Rust\"]}\n",
		},
	})
}

func HandleJsonLinesDeserializationVerification(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "application/jsonl")

	pushEvents(rw, [][]string{
		{
			"{\"isFinished\": \"yes\"}\n",
		},
	})

}

// HandleJSONLinesHold writes two JSONL events immediately, then holds the
// connection open until the client disconnects (capped at 10 seconds). A
// client that buffers the response body to EOF before returning it cannot
// deliver the first event until the hold expires, so incremental delivery is
// distinguishable from buffered delivery without relying on tight timing.
func HandleJSONLinesHold(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "application/jsonl")

	fmt.Fprint(rw, "{\"name\": \"Peter\", \"skills\": [\"Go\", \"Python\"]}\n")
	fmt.Fprint(rw, "{\"name\": \"John\", \"skills\": [\"Go\", \"Rust\"]}\n")
	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	select {
	case <-r.Context().Done():
	case <-time.After(10 * time.Second):
	}
}
