package eventstreams

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type sseEvent struct {
	ID    *string `json:"id"`
	Data  string  `json:"data"`
	Event *string `json:"event,omitempty"`
	Retry *int64  `json:"retry,omitempty"`
}

func pushChunks(rw http.ResponseWriter, chunks []string) {
	for _, chunk := range chunks {
		fmt.Fprintln(rw, chunk)

		if f, ok := rw.(http.Flusher); ok {
			f.Flush()
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func pushEvents(rw http.ResponseWriter, events [][]string) {
	for _, event := range events {
		for _, line := range event {
			fmt.Fprintln(rw, line)
		}
		fmt.Fprintln(rw, "")

		if f, ok := rw.(http.Flusher); ok {
			f.Flush()
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func HandleEventStreamJSON(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/event-stream")

	pushEvents(rw, [][]string{
		{
			`data: {"content": "Hello"}`,
		},

		{
			`data: {"content": " "}`,
		},

		{
			`data: {"content": "world"}`,
		},

		{
			`data: {"content": "!"}`,
		},
	})
}

// HandleEventStreamMalformed emits a well-formed frame followed by a frame
// whose data violates the jsonEvent schema (missing required `content`). The
// generated stream decoder must surface this as a ResponseValidationError
// instead of leaking a raw pydantic ValidationError mid-iteration.
func HandleEventStreamMalformed(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/event-stream")

	pushEvents(rw, [][]string{
		{
			`data: {"content": "Hello"}`,
		},
		{
			`data: {"unexpected": true}`,
		},
	})
}

func HandleEventStreamText(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/event-stream")

	pushEvents(rw, [][]string{
		{
			`data: Hello`,
		},

		{
			`data:  `,
		},

		{
			`data: world`,
		},

		{
			`data: !`,
		},
	})
}

func HandleEventStreamMultiLine(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/event-stream")

	pushEvents(rw, [][]string{
		{
			`data: YHOO`,
			`data: +2`,
			`data: 10`,
		},
	})
}

func HandleEventStreamRich(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/event-stream")

	pushEvents(rw, [][]string{
		{
			`id: job-1`,
			`event: completion`,
			`data: {"completion": "Hello", "stop_reason": null, "model": "jeeves-1"}`,
		},

		{
			`event: heartbeat`,
			`data: ping`,
			`retry: 3000`,
		},

		{
			`id: job-1`,
			`event: completion`,
			`data: {"completion": "world!", "stop_reason": "stop_sequence", "model": "jeeves-1"}`,
		},
	})
}

func HandleEventStreamChat(rw http.ResponseWriter, r *http.Request) {
	var req struct {
		Stream *bool `json:"stream"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	if req.Stream != nil && !*req.Stream {
		rw.Header().Add("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rw).Encode([]map[string]any{
			{"data": map[string]any{"content": "Hello"}},
			{"data": map[string]any{"content": " "}},
			{"data": map[string]any{"content": "world"}},
			{"data": map[string]any{"content": "!"}},
		})
		return
	}

	rw.Header().Add("Content-Type", "text/event-stream")

	pushEvents(rw, [][]string{
		{
			`data: {"content": "Hello"}`,
		},

		{
			`data: {"content": " "}`,
		},

		{
			`data: {"content": "world"}`,
		},

		{
			`data: {"content": "!"}`,
		},

		{
			`data: [DONE]`,
		},
	})
}

// HandleEventStreamChatOverload returns SSE when the client signals streaming
// (Accept: text/event-stream or body.stream=true), otherwise a JSON array
// matching `chatCompletionResult`.
func HandleEventStreamChatOverload(rw http.ResponseWriter, r *http.Request) {
	var req struct {
		Stream bool   `json:"stream"`
		Prompt string `json:"prompt"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.Stream || strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		HandleEventStreamChat(rw, r)
		return
	}

	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(rw).Encode([]map[string]any{
		{"data": map[string]any{"content": "Hello"}},
		{"data": map[string]any{"content": " "}},
		{"data": map[string]any{"content": "world"}},
		{"data": map[string]any{"content": "!"}},
	})
}

func HandleEventStreamChatFlatten(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/event-stream")

	pushEvents(rw, [][]string{
		{
			`data: {"content": "Hello"}`,
		},

		{
			`data: {"content": " "}`,
		},

		{
			`data: {"content": "world"}`,
		},

		{
			`data: {"content": "!"}`,
		},

		{
			`data: [DONE]`,
		},
	})
}

func HandleEventStreamChatChunked(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/event-stream")

	pushChunks(rw, []string{
		"data: {\"content\": ",
		"\"Hello\"}\n\ndata: {\"content\": \" \"}",
		"data: {\"content\": \"world\"}",
		"data: {\"content\": \"!\"}\n\ndata: [DONE]\n",
		"\ndata: {\"content\": \"Post sentinel data\"}\n\n",
	})
}

func HandleEventStreamDifferentDataSchemas(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/event-stream")

	pushEvents(rw, [][]string{
		{
			`id: event-1`,
			`event: message`,
			`data: {"id": 123, "content": "Here is your url"}`,
		},
		{
			`id: event-2`,
			`event: url`,
			`data: {"url": "https://example.com"}`,
		},
		{
			`id: event-3`,
			`event: message`,
			`data: {"content": "Have a great day!"}`,
		},
		{
			`id: event-4`,
			`event: array`,
			`data: [1, 2, 3, 4]`,
		},
		{
			`id: event-5`,
			`event: primitive`,
			`data: true`,
		},
		{
			`id: event-6`,
			`event: primitive`,
			`data: 3.14159`,
		},
	})
}

func HandleEventStreamStayOpen(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "text/event-stream")
	rw.Header().Add("Cache-Control", "no-cache")
	rw.Header().Add("Connection", "keep-alive")

	// Send events 1, 2, 3 immediately
	fmt.Fprintln(rw, "data: event 1")
	fmt.Fprintln(rw, "")
	fmt.Fprintln(rw, "data: event 2")
	fmt.Fprintln(rw, "")
	fmt.Fprintln(rw, "data: event 3")
	fmt.Fprintln(rw, "")

	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	// Wait 100ms then send event 4
	time.Sleep(100 * time.Millisecond)
	fmt.Fprintln(rw, "data: event 4")
	fmt.Fprintln(rw, "")

	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	// Wait another 100ms then send sentinel event
	time.Sleep(100 * time.Millisecond)
	fmt.Fprintln(rw, "data: [SENTINEL]")
	fmt.Fprintln(rw, "")

	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	// Keep the connection open until client closes
	// Monitor the request context to detect when client disconnects
	<-r.Context().Done()
}

func HandleEventStreamUnionWithComments(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/event-stream")

	// First status event
	fmt.Fprint(rw, "event: status\n")
	fmt.Fprint(rw, `data: {"status": "started", "message": "Initializing"}`)
	fmt.Fprint(rw, "\n\n")

	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	time.Sleep(100 * time.Millisecond)

	// Standalone comment as its own double-newline-delimited block between events.
	// This is the key scenario: a comment that is NOT part of an event block but
	// appears as its own block between two events.
	fmt.Fprint(rw, ": heartbeat\n\n")

	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	time.Sleep(100 * time.Millisecond)

	// Progress event after the standalone comment
	fmt.Fprint(rw, "event: progress\n")
	fmt.Fprint(rw, `data: {"percent": 50, "detail": "Half done"}`)
	fmt.Fprint(rw, "\n\n")

	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	time.Sleep(100 * time.Millisecond)

	// Second status event
	fmt.Fprint(rw, "event: status\n")
	fmt.Fprint(rw, "data: {\"status\": \"running\",\n")
	fmt.Fprint(rw, "data:  \"message\": \"Processing items\"}\n\n")

	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	time.Sleep(100 * time.Millisecond)

	// Final progress event
	fmt.Fprint(rw, "event: progress\n")
	fmt.Fprint(rw, `data: {"percent": 100, "detail": "Complete"}`)
	fmt.Fprint(rw, "\n\n")

	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}
}

func HandleEventStreamPartialWithComments(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/event-stream")

	// Send the first packet with a partial message and a comment
	fmt.Fprint(rw, ": This is a comment\n")
	fmt.Fprint(rw, "data: {\"message\": \"Hello ")

	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	time.Sleep(100 * time.Millisecond)

	// Complete the first message with LF,LF boundary and add another comment
	fmt.Fprint(rw, "from SSE\"}\n\n")
	fmt.Fprint(rw, ": Another comment line\n")

	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	time.Sleep(100 * time.Millisecond)

	// Send a complete event with CR,CR boundary
	fmt.Fprint(rw, "id: msg-2\n")
	fmt.Fprint(rw, "event: update\n")
	fmt.Fprint(rw, ": Comment before data\n")
	fmt.Fprint(rw, "data: {\"status\": \"processing\", \"progress\": 50}\r\r")

	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	time.Sleep(100 * time.Millisecond)

	// Send with CR,LF,CR,LF boundary
	fmt.Fprint(rw, ": This is a multiline\r\n")
	fmt.Fprint(rw, ": comment that spans\r\n")
	fmt.Fprint(rw, ": multiple lines\r\n")
	fmt.Fprint(rw, "id: msg-3\r\n")
	fmt.Fprint(rw, "data: {\"status\": \"complete\",\r\n")
	fmt.Fprint(rw, "data:  \"progress\": 100,\r\n")
	fmt.Fprint(rw, "data:  \"result\": \"Success\"}\r\n\r\n")

	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	time.Sleep(100 * time.Millisecond)

	// Mix boundaries within same message group - CR for lines, LF,LF for message end
	fmt.Fprint(rw, ": Mixed line endings\r")
	fmt.Fprint(rw, "event: mixed\n")
	fmt.Fprint(rw, "id: msg-4\r")
	fmt.Fprint(rw, "data: {\"test\": \"mixed boundaries\"}\n\n")

	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	time.Sleep(100 * time.Millisecond)

	// Another variant with CR,CR ending
	fmt.Fprint(rw, "data: {\"another\": \"test\"}\r")
	fmt.Fprint(rw, ": Comment with CR\r")
	fmt.Fprint(rw, "id: msg-5\r\r")

	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}

	time.Sleep(100 * time.Millisecond)

	// Send a final comment and done signal with standard LF,LF
	fmt.Fprint(rw, ": Stream ending\n")
	fmt.Fprint(rw, "data: [DONE]\n\n")

	if f, ok := rw.(http.Flusher); ok {
		f.Flush()
	}
}

func HandleEventStreamOptionalData(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/event-stream")

	pushEvents(rw, [][]string{
		{
			// Event with data field present
			`event: message`,
			`data: {"content": "Hello, this event has data"}`,
			`id: event-1`,
		},

		{
			// Event without data field (data is optional)
			`event: heartbeat`,
			`id: event-2`,
		},

		{
			// Event with data field present
			`event: message`,
			`data: {"content": "Another message with data"}`,
			`id: event-3`,
		},

		{
			// Event without data field (data is optional)
			`event: ping`,
			`id: event-4`,
		},

		{
			// Final event with data
			`event: complete`,
			`data: {"content": "Stream finished"}`,
			`id: event-5`,
		},
	})
}

func HandleEventStreamMixedData(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/event-stream")

	pushEvents(rw, [][]string{
		{
			`event: completion`,
			`data: {"id": 1, "content": "Hello world"}`,
		},
		{
			`event: text`,
			`data: Processing your request...`,
		},
		{
			`event: loading`,
			`data: Almost done`,
		},
		{
			`event: completion`,
			`data: {"id": 2, "content": "Done!"}`,
		},
	})
}

// HandleEventStreamWPTCompliance covers SSE parsing edge cases adapted from WPT [1]
// to comply with section 9.2.6 of the HTML Living Standard [2].
// [1] https://github.com/web-platform-tests/wpt/tree/master/eventsource
// [2] https://html.spec.whatwg.org/multipage/server-sent-events.html#parsing-an-event-stream
//
// Implemented WPT files (letter = section):
//   - format-bom.any.js, format-bom-2.any.js (A: BOM handling)
//   - format-null-character.any.js (B: null in data)
//   - format-field-parsing.any.js (C: field parsing)
//   - format-leading-space.any.js (D: tab/space after colon)
//   - format-comments.any.js (E: comment lines, buffer stress)
//   - format-field-unknown.any.js (F: unknown fields)
//   - format-field-data.any.js (G: data field variations)
//   - format-newlines.any.js (H: CR/LF/CRLF)
//   - format-field-retry*.any.js (I: retry field)
//   - format-field-event*.any.js (J: event type)
//   - format-field-id*.js (K: ID field)
//   - event-data.any.js (L: multi-message parsing)
//   - format-utf-8.any.js (M: UTF-8 multibyte)
//   - SSE spec (N: colon in value)
//
// Skipped WPT files:
//   - eventsource-*.js: Browser API (constructor, close, reconnect, CORS, etc.)
//   - format-mime-*.any.js: HTTP content-type validation
//   - request-*.any.js: HTTP headers (Accept, Cache-Control)
//   - request-*.window.js: HTTP credentials, redirects, status codes
//   - dedicated-worker/*: Web Workers
//   - shared-worker/*: Web Workers
//   - format-data-before-final-empty-line.any.js: Reconnection behavior
func HandleEventStreamWPTCompliance(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/event-stream")

	ptr := func(s string) *string { return &s }
	intPtr := func(i int64) *int64 { return &i }

	sendEvent := func(msg string) {
		fmt.Fprint(rw, msg)

		if f, ok := rw.(http.Flusher); ok {
			f.Flush()
		}
		time.Sleep(20 * time.Millisecond)
	}

	var expected []sseEvent
	expectData := func(id, data string) {
		expected = append(expected, sseEvent{ID: ptr(id), Data: data})
	}

	// ---------------------------------------------------------------------------
	// A: BOM handling (format-bom.any.js, format-bom-2.any.js)
	// ---------------------------------------------------------------------------
	// BOM at start gets stripped
	sendEvent("\uFEFFid: A1\ndata: bom-start\n\n")
	expectData("A1", "bom-start")
	// BOM mid-stream corrupts field name
	sendEvent("id: A2\n\uFEFFdata: ignored\ndata: bom-midstream\n\n")
	expectData("A2", "bom-midstream")
	// Only first BOM is stripped; second corrupts field
	sendEvent("\uFEFF\uFEFFdata: ignored\nid: A3\ndata: double-bom\n\n")
	expectData("A3", "double-bom")

	// ---------------------------------------------------------------------------
	// B: Null character (format-null-character.any.js)
	// ---------------------------------------------------------------------------
	// Null char preserved in data value
	sendEvent("id: B1\ndata: null\x00char\n\n")
	expectData("B1", "null\x00char")

	// ---------------------------------------------------------------------------
	// C: Field parsing (format-field-parsing.any.js)
	// ---------------------------------------------------------------------------
	// Only one leading space after colon is stripped
	sendEvent("id: C1\ndata:  double-space\n\n")
	expectData("C1", " double-space")
	// Field names are case-sensitive
	sendEvent("id: C2\nData: ignored\nDATA: also-ignored\ndata: case-sensitive\n\n")
	expectData("C2", "case-sensitive")
	// Null char in field name invalidates the line
	sendEvent("id: C3\ndata\x00: ignored\ndata: field-null\n\n")
	expectData("C3", "field-null")
	// CR terminators, null at line start, unknown fields (da-ta, data_5), space-prefixed lines
	sendEvent("id: C4\ndata:\x00\ndata:  2\rData:1\ndata\x00:2\ndata:1\r\x00data:4\nda-ta:3\rdata_5\ndata:3\rdata:\r\n data:32\ndata:4\n\n")
	expectData("C4", "\x00\n 2\n1\n3\n\n4")

	// ---------------------------------------------------------------------------
	// D: Leading space/tab (format-leading-space.any.js)
	// ---------------------------------------------------------------------------
	// Tab after colon is preserved (only space is stripped)
	sendEvent("id: D1\ndata:\ttab-preserved\n\n")
	expectData("D1", "\ttab-preserved")
	// CR terminator, empty data value (space stripped), multi-line joining
	sendEvent("id: D2\ndata:\ttest\rdata: \ndata:test\n\n")
	expectData("D2", "\ttest\n\ntest")

	// ---------------------------------------------------------------------------
	// E: Comments (format-comments.any.js)
	// ---------------------------------------------------------------------------
	// Comment lines (starting with :) are ignored; null in comment also ignored
	sendEvent("id: E1\ndata:1\r:\x00\n:\rdata:2\n:comment\rdata:3\n:data:fail\r:another\ndata:4\n\n")
	expectData("E1", "1\n2\n3\n4")
	// Buffer stress test: 2049 char comment (colon included) to cross common buffer boundaries
	longComment := ":" + strings.Repeat("x", 2048)
	sendEvent(fmt.Sprintf("id: E2\ndata:1\n%s\ndata:2\n%s\ndata:3\n\n", longComment, longComment))
	expectData("E2", "1\n2\n3")

	// ---------------------------------------------------------------------------
	// F: Unknown fields (format-field-unknown.any.js)
	// ---------------------------------------------------------------------------
	// Unknown fields ignored
	sendEvent("id: F1\ndata:test\n data\ndata\nfoobar:xxx\njustsometext\n:thisisacomment\ndata:test\n\n")
	expectData("F1", "test\n\ntest")
	// " data" (with leading space) is unknown field name
	sendEvent("id: F2\n data: ignored\ndata: space-prefix\n\n")
	expectData("F2", "space-prefix")

	// ---------------------------------------------------------------------------
	// G: Data field (format-field-data.any.js)
	// ---------------------------------------------------------------------------
	// Empty data value
	sendEvent("id: G1\ndata:\n\n")
	expectData("G1", "")
	// "data" without colon treated as empty value
	sendEvent("id: G2\ndata\ndata\n\n")
	expectData("G2", "\n")
	// Multiple data lines joined with newlines
	sendEvent("id: G3\ndata: line1\ndata: line2\ndata: line3\n\n")
	expectData("G3", "line1\nline2\nline3")
	// Space after colon is optional
	sendEvent("id: G4\ndata:test\n\n")
	expectData("G4", "test")

	// ---------------------------------------------------------------------------
	// H: Newlines (format-newlines.any.js)
	// ---------------------------------------------------------------------------
	// CR as line terminator
	sendEvent("id: H1\rdata: cr-ending\r\r")
	expectData("H1", "cr-ending")
	// Mixed CRLF and LF line endings
	sendEvent("id: H2\ndata:test\r\ndata\ndata:test\r\n\r")
	expectData("H2", "test\n\ntest")

	// ---------------------------------------------------------------------------
	// I: Retry field (format-field-retry*.any.js)
	// ---------------------------------------------------------------------------
	// Leading zeros parsed as decimal, not octal
	sendEvent("id: I1\nretry: 03000\ndata: retry-valid\n\n")
	expected = append(expected, sseEvent{ID: ptr("I1"), Data: "retry-valid", Retry: intPtr(3000)})
	// Bogus retry (1000x) ignored; previous valid (3000) retained
	sendEvent("id: I2\nretry: 3000\nretry: 1000x\ndata: retry-bogus\n\n")
	expected = append(expected, sseEvent{ID: ptr("I2"), Data: "retry-bogus", Retry: intPtr(3000)})
	// "retry" without colon has no effect
	sendEvent("id: I3\nretry\ndata: retry-empty\n\n")
	expectData("I3", "retry-empty")

	// ---------------------------------------------------------------------------
	// J: Event type (format-field-event*.any.js)
	// ---------------------------------------------------------------------------
	// Custom event type
	sendEvent("id: J1\nevent: custom\ndata: custom-event\n\n")
	expected = append(expected, sseEvent{ID: ptr("J1"), Data: "custom-event", Event: ptr("custom")})
	// Empty event type
	sendEvent("id: J2\nevent: \ndata: empty-event\n\n")
	expected = append(expected, sseEvent{ID: ptr("J2"), Data: "empty-event", Event: ptr("")})

	// ---------------------------------------------------------------------------
	// K: ID field (format-field-id*.js)
	// ---------------------------------------------------------------------------
	// Non-numeric string ID
	sendEvent("id: K1\ndata: has-id\n\n")
	expectData("K1", "has-id")
	// Sets ID for persistence test below
	sendEvent("id: K2\ndata: id-set\n\n")
	expectData("K2", "id-set")
	// No ID field sent; last event ID "K2" persists from previous event
	sendEvent("data: id-persists\n\n")
	expected = append(expected, sseEvent{ID: ptr("K2"), Data: "id-persists"})
	// Empty ID resets last event ID
	sendEvent("id:\ndata: id-reset\n\n")
	expectData("", "id-reset")
	// ID without colon resets last event ID
	sendEvent("id\ndata: id-no-colon\n\n")
	expectData("", "id-no-colon")
	// Sets ID for null-ID test below
	sendEvent("id: K6\ndata: before-null\n\n")
	expectData("K6", "before-null")
	// ID containing null char is ignored; last event ID "K6" persists
	sendEvent("id: bad\x00id\ndata: null-id-ignored\n\n")
	expected = append(expected, sseEvent{ID: ptr("K6"), Data: "null-id-ignored"})

	// ---------------------------------------------------------------------------
	// L: Multi-message parsing (event-data.any.js)
	// ---------------------------------------------------------------------------
	sendEvent("id: L1\ndata:msg\ndata:msg\n\n")
	expectData("L1", "msg\nmsg")
	sendEvent("id: L2\ndata\n\n")
	expectData("L2", "")
	sendEvent("id: L3\ndata:end\n\n")
	expectData("L3", "end")

	// ---------------------------------------------------------------------------
	// M: UTF-8 (format-utf-8.any.js)
	// ---------------------------------------------------------------------------
	// UTF-8 multibyte characters preserved
	sendEvent("id: M1\ndata: ok\u2026\n\n")
	expectData("M1", "ok\u2026")

	// ---------------------------------------------------------------------------
	// N: Colon in value (SSE spec)
	// ---------------------------------------------------------------------------
	// Only first colon is field separator; subsequent colons are part of value
	sendEvent("id: N1\ndata: has:colon:in:value\n\n")
	expectData("N1", "has:colon:in:value")

	// Send all expectations in a final message (escaped array)
	expectedData, _ := json.Marshal(expected)
	escapedData, _ := json.Marshal(string(expectedData))
	sendEvent(fmt.Sprintf("id: Z\nevent: expected\ndata: %s\n\n", escapedData))
}

func HandleEventStreamChatHeartbeat(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/event-stream")

	pushEvents(rw, [][]string{
		{
			`data: {"content": "Hello1"}`,
		},

		{
			// Heartbeat with no data field — must be skipped by parsers
			// that expect required data (e.g. chatCompletionEvent)
			`event: heartbeat`,
		},

		{
			`data: {"content": "Hello 2"}`,
		},

		{
			`data: {"content": "!"}`,
		},

		{
			`data: [DONE]`,
		},
	})
}

// HandleEventStreamLargeEvent streams a small lead-in event, one very large
// event, and a small trailing event. The large event's data is a JSON object
// whose content field is `size` bytes ("S" + "a"... + "E"), well above the
// 64KB cap of Go's default bufio.Scanner token size. The whole response is
// written in fixed 1536-byte chunks with a flush after each one and no
// delays, so client parsers must reassemble a single event across many small
// reads without rescanning the accumulated buffer from the start each time.
func HandleEventStreamLargeEvent(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Add("Content-Type", "text/event-stream")

	size := 2 * 1024 * 1024
	if v := r.FormValue("size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 2 && n <= 32*1024*1024 {
			size = n
		}
	}

	var sb strings.Builder
	sb.WriteString(`data: {"content": "start"}` + "\n\n")
	sb.WriteString(`data: {"content": "S`)
	sb.WriteString(strings.Repeat("a", size-2))
	sb.WriteString(`E"}` + "\n\n")
	sb.WriteString(`data: {"content": "end"}` + "\n\n")
	payload := sb.String()

	flusher, _ := rw.(http.Flusher)
	const chunkSize = 1536
	for off := 0; off < len(payload); off += chunkSize {
		end := off + chunkSize
		if end > len(payload) {
			end = len(payload)
		}
		if _, err := io.WriteString(rw, payload[off:end]); err != nil {
			return
		}
		if flusher != nil {
			flusher.Flush()
		}
	}
}

type splitBoundaryPayload struct {
	Kind string   `json:"kind"`
	Tags []string `json:"tags"`
}

// boundary terminators recognized by SSE parsers, in the same order as the
// MESSAGE_BOUNDARIES tables in the generated parsers.
var splitBoundaryShapes = []struct {
	tag      string
	boundary string
	split    int // bytes flushed in the first chunk; remainder in the second
}{
	{"lf-lf-1-1", "\n\n", 1},
	{"cr-cr-1-1", "\r\r", 1},
	{"lf-cr-1-1", "\n\r", 1},
	{"crlf-cr-2-1", "\r\n\r", 2},
	{"cr-lfcr-1-2", "\r\n\r", 1},
	{"crlf-crlf-2-2", "\r\n\r\n", 2},
	{"cr-lfcrlf-1-3", "\r\n\r\n", 1},
	{"crlfcr-lf-3-1", "\r\n\r\n", 3},
	{"crlf-lf-2-1", "\r\n\n", 2},
	{"cr-lflf-1-2", "\r\n\n", 1},
	{"cr-crlf-1-2", "\r\r\n", 1},
	{"crcr-lf-2-1", "\r\r\n", 2},
	{"lf-crlf-1-2", "\n\r\n", 1},
	{"lfcr-lf-2-1", "\n\r\n", 2},
}

// HandleEventStreamSplitBoundaries emits one SSE event per entry in
// splitBoundaryShapes, deliberately flushing the boundary terminator across
// a network seam in the shape described by the tag. After all sample events
// a final "expected" event carries a JSON array of the sample tags in
// emission order. A parser that only inspects the newest chunk in isolation
// will drop or misalign events; a correct parser reassembles each
// terminator and yields the events in order.
func HandleEventStreamSplitBoundaries(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Add("Content-Type", "text/event-stream")

	flusher, _ := rw.(http.Flusher)
	flush := func(s string) {
		_, _ = io.WriteString(rw, s)
		if flusher != nil {
			flusher.Flush()
		}
	}

	var tags []string
	for _, shape := range splitBoundaryShapes {
		tags = append(tags, shape.tag)
		payload, _ := json.Marshal(splitBoundaryPayload{Kind: "sample", Tags: []string{shape.tag}})
		body := "data: " + string(payload)
		flush(body + shape.boundary[:shape.split])
		flush(shape.boundary[shape.split:])
	}

	manifest, _ := json.Marshal(splitBoundaryPayload{Kind: "expected", Tags: tags})
	flush("data: " + string(manifest) + "\n\n")
}
