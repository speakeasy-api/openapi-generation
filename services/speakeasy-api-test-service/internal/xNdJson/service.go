package xNdJson

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

func HandleXNdJsonLinesChunksRich(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "application/x-ndjson")

	pushChunks(rw, []string{
		"{\"name\": \"Peter\", \"skills\": [\"Go\"",
		", \"Python\"]}\n{\"name\": \"John\"",
		", \"skills\": [\"Go\", \"Rust\"]}\n",
	})
}

func HandleXNdJsonLinesRich(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "application/x-ndjson")

	pushEvents(rw, [][]string{
		{
			"{\"name\": \"Peter\", \"skills\": [\"Go\", \"Python\"]}\n",
		},
		{
			"{\"name\": \"John\", \"skills\": [\"Go\", \"Rust\"]}\n",
		},
	})
}
