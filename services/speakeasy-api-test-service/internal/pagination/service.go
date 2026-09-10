package pagination

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
)

type LimitOffsetRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Page   int `json:"page"`
}

type CursorRequest struct {
	Cursor int `json:"cursor"`
}

type NonNumericCursorRequest struct {
	Cursor string `json:"cursor"`
}

type PaginationResponse struct {
	NumPages    int           `json:"numPages"`
	ResultArray []interface{} `json:"resultArray"`
	Next        *string       `json:"next,omitempty"`
	Cursor      *string       `json:"cursor"`
}

type PageInfo struct {
	NumPages int     `json:"numPages"`
	Next     *string `json:"next,omitempty"`
}
type PaginationResponseDeep struct {
	ResultArray []interface{} `json:"resultArray"`
	PageInfo    PageInfo      `json:"pageInfo"`
}

// Insecure reversable hashing for string cursors
func hash(s string) (int, error) {
	return strconv.Atoi(s)
}
func unhash(h int) string {
	return strconv.Itoa(h)
}

const total = 20

func HandleLimitOffsetPage(w http.ResponseWriter, r *http.Request) {
	queryLimit := r.FormValue("limit")
	queryPage := r.FormValue("page")

	var pagination LimitOffsetRequest
	hasBody := true
	if err := json.NewDecoder(r.Body).Decode(&pagination); err != nil {
		hasBody = false
	}
	limit := getValue(queryLimit, hasBody, pagination.Limit)
	if limit == 0 {
		limit = 20
	}
	page := getValue(queryPage, hasBody, pagination.Page)

	start := (page - 1) * limit

	res := PaginationResponse{
		NumPages:    int(math.Ceil(float64(total) / float64(limit))),
		ResultArray: make([]interface{}, 0),
	}

	for i := start; i < total && len(res.ResultArray) < limit; i++ {
		res.ResultArray = append(res.ResultArray, i)
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		w.WriteHeader(500)
	}
}

func HandleLimitOffsetOffset(w http.ResponseWriter, r *http.Request) {
	queryLimit := r.FormValue("limit")
	queryOffset := r.FormValue("offset")

	var pagination LimitOffsetRequest
	json.NewDecoder(r.Body).Decode(&pagination)
	hasBody := pagination.Limit != 0 || pagination.Offset != 0

	limit := getValue(queryLimit, hasBody, pagination.Limit)
	if limit == 0 {
		limit = 20
	}
	offset := getValue(queryOffset, hasBody, pagination.Offset)

	res := PaginationResponse{
		NumPages:    int(math.Ceil(float64(total) / float64(limit))),
		ResultArray: make([]interface{}, 0),
	}

	for i := offset; i < total && len(res.ResultArray) < limit; i++ {
		res.ResultArray = append(res.ResultArray, i)
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		w.WriteHeader(500)
	}
}

func HandleCursor(w http.ResponseWriter, r *http.Request) {
	queryCursor := r.FormValue("cursor")

	var pagination CursorRequest
	hasBody := true
	if err := json.NewDecoder(r.Body).Decode(&pagination); err != nil {
		hasBody = false
	}

	cursor := getValue(queryCursor, hasBody, pagination.Cursor)
	resultArray := make([]interface{}, 0)

	for i := cursor + 1; i < total && len(resultArray) < 15; i++ {
		resultArray = append(resultArray, i)
	}

	w.Header().Set("Content-Type", "application/json")
	res := PaginationResponse{
		NumPages:    0,
		ResultArray: resultArray,
	}

	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		w.WriteHeader(500)
	}

}

func HandleCursorResponseEnvelope(w http.ResponseWriter, r *http.Request) {
	queryCursor := r.FormValue("cursor")

	var pagination CursorRequest
	hasBody := true
	if err := json.NewDecoder(r.Body).Decode(&pagination); err != nil {
		hasBody = false
	}

	cursor := getValue(queryCursor, hasBody, pagination.Cursor)
	resultArray := make([]interface{}, 0)

	for i := cursor + 1; i < total && len(resultArray) < 15; i++ {
		resultArray = append(resultArray, i)
	}

	w.Header().Set("Content-Type", "application/json")
	var lastItem *string
	// conditionally set lastItem
	if len(resultArray) > 0 {
		idx := strconv.Itoa(resultArray[len(resultArray)-1].(int))
		lastItem = &idx
	} else {
		lastItem = nil
	}
	res := PaginationResponseDeep{
		PageInfo: PageInfo{
			Next: lastItem,
		},
		ResultArray: resultArray,
	}

	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		w.WriteHeader(500)
	}

}

func HandleURL(w http.ResponseWriter, r *http.Request) {
	attemptsString := r.FormValue("attempts")
	isReferencePath := r.FormValue("is-reference-path")
	var attempts int
	if attemptsString != "" {
		var err error
		attempts, err = strconv.Atoi(attemptsString)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("attempts must be an integer"))
			return
		}
	}

	res := PaginationResponse{
		NumPages:    0,
		ResultArray: make([]interface{}, 0),
	}

	// Return 9, 6, then 3 results for 18 total results.
	for i := 0; i < total && len(res.ResultArray) < (attempts*3); i++ {
		res.ResultArray = append(res.ResultArray, i)
	}

	if attempts > 1 {
		baseURL := fmt.Sprintf("%s://%s", r.URL.Scheme, r.Host)
		if r.URL.Scheme == "" { // Fallback if Scheme is not available
			baseURL = fmt.Sprintf("http://%s", r.Host)
		}

		if isReferencePath == "true" {
			baseURL = r.URL.Path
		} else {
			baseURL = fmt.Sprintf("%s%s", baseURL, r.URL.Path)
		}

		nextUrl := fmt.Sprintf("%s?attempts=%d", baseURL, attempts-1)
		res.Next = &nextUrl
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		w.WriteHeader(500)
	}
}

func HandleNonNumericCursor(w http.ResponseWriter, r *http.Request) {
	vals, ok := r.URL.Query()["endCursor"]
	var endCursor *string
	if ok && len(vals) > 0 {
		endCursor = &vals[0]
	}

	limit := 15

	queryCursor := r.FormValue("cursor")
	var pagination NonNumericCursorRequest
	hasBody := true
	if err := json.NewDecoder(r.Body).Decode(&pagination); err != nil {
		hasBody = false
	}
	cursor := getNonNumericValue(queryCursor, hasBody, pagination.Cursor)

	res := PaginationResponse{
		NumPages:    0,
		ResultArray: make([]interface{}, 0),
	}

	var cursorI, _ = hash(cursor)
	for i := cursorI + 1; i < total && len(res.ResultArray) < limit; i++ {
		res.ResultArray = append(res.ResultArray, unhash(i))
	}

	if len(res.ResultArray) == limit {
		cursor, _ := res.ResultArray[len(res.ResultArray)-1].(string)
		res.Cursor = &cursor
	} else if endCursor != nil {
		res.Cursor = endCursor
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		w.WriteHeader(500)
	}
}

func HandleLimitOffsetDeepOutputsPage(w http.ResponseWriter, r *http.Request) {
	queryLimit := r.FormValue("limit")
	queryPage := r.FormValue("page")

	var pagination LimitOffsetRequest
	hasBody := true
	if err := json.NewDecoder(r.Body).Decode(&pagination); err != nil {
		hasBody = false
	}
	limit := getValue(queryLimit, hasBody, pagination.Limit)
	if limit == 0 {
		limit = 20
	}
	page := getValue(queryPage, hasBody, pagination.Page)

	start := (page - 1) * limit

	res := PaginationResponseDeep{
		PageInfo: PageInfo{
			NumPages: int(math.Ceil(float64(total) / float64(limit))),
		},
		ResultArray: make([]interface{}, 0),
	}

	for i := start; i < total && len(res.ResultArray) < limit; i++ {
		res.ResultArray = append(res.ResultArray, i)
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		w.WriteHeader(500)
	}
}

type NullableLimitResponse struct {
	Results    []string `json:"results"`
	NextCursor *string  `json:"next_cursor"`
}

// HandleCursorNullable serves cursor pagination, returning a page of items and
// the next cursor. By default the cursor stops on the last non-empty page. With
// `?null_terminal=true` the cursor advances whenever a page yields items, so
// once the items are exhausted a final page is still returned whose `results`
// is an explicit null (a nil slice marshals to JSON null).
func HandleCursorNullable(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.FormValue("cursor")
	limitStr := r.FormValue("limit")
	nullTerminal := r.FormValue("null_terminal") == "true"

	cursor := 0
	if cursorStr != "" {
		c, err := strconv.Atoi(cursorStr)
		if err == nil {
			cursor = c
		}
	}

	limit := 10
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 {
			limit = l
		}
	}

	var results []string
	for i := cursor; i < total && len(results) < limit; i++ {
		results = append(results, fmt.Sprintf("item_%d", i))
	}

	var nextCursor *string
	if nullTerminal {
		// Advance while a page still yields items so the null terminating page
		// (empty -> null results) is always fetched and traversed.
		if len(results) > 0 {
			nc := strconv.Itoa(cursor + len(results))
			nextCursor = &nc
		}
	} else if cursor+len(results) < total {
		nc := strconv.Itoa(cursor + len(results))
		nextCursor = &nc
	}

	res := NullableLimitResponse{
		Results:    results,
		NextCursor: nextCursor,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(500)
	}
}

type DeepNestedData struct {
	Items []string `json:"items"`
}

type DeepNestedPaging struct {
	NextCursor *string `json:"next_cursor"`
}

type DeepNestedMeta struct {
	Paging *DeepNestedPaging `json:"paging"`
}

type DeepNestedOutputsResponse struct {
	Data DeepNestedData  `json:"data"`
	Meta *DeepNestedMeta `json:"meta,omitempty"`
}

// HandleCursorDeepNestedOutputs serves cursor pagination with outputs nested
// several levels deep ($.data.items and $.meta.paging.next_cursor) across a
// mixed required/optional/nullable hierarchy. The final page terminates the
// cursor chain according to final_page_style: "omit-meta" (default) leaves
// out the optional meta object, "null-paging" sets the nullable paging object
// to null, and "null-cursor" sets the nullable next_cursor leaf to null.
func HandleCursorDeepNestedOutputs(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.FormValue("cursor")
	limitStr := r.FormValue("limit")
	finalPageStyle := r.FormValue("final_page_style")

	cursor := 0
	if cursorStr != "" {
		if c, err := strconv.Atoi(cursorStr); err == nil {
			cursor = c
		}
	}

	limit := 5
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	items := make([]string, 0)
	for i := cursor; i < total && len(items) < limit; i++ {
		items = append(items, fmt.Sprintf("item_%d", i))
	}

	res := DeepNestedOutputsResponse{
		Data: DeepNestedData{Items: items},
	}
	if cursor+len(items) < total {
		next := strconv.Itoa(cursor + len(items))
		res.Meta = &DeepNestedMeta{
			Paging: &DeepNestedPaging{NextCursor: &next},
		}
	} else {
		switch finalPageStyle {
		case "null-paging":
			res.Meta = &DeepNestedMeta{Paging: nil}
		case "null-cursor":
			res.Meta = &DeepNestedMeta{Paging: &DeepNestedPaging{NextCursor: nil}}
		default: // "omit-meta"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(500)
	}
}

func HandleLimitOffsetNullable(w http.ResponseWriter, r *http.Request) {
	queryOffset := r.FormValue("offset")
	queryLimit := r.FormValue("limit")

	offset := 0
	if queryOffset != "" && queryOffset != "null" {
		if v, err := strconv.Atoi(queryOffset); err == nil {
			offset = v
		}
	}

	limit := 10
	if queryLimit != "" && queryLimit != "null" {
		if v, err := strconv.Atoi(queryLimit); err == nil && v > 0 {
			limit = v
		}
	}

	res := PaginationResponse{
		NumPages:    int(math.Ceil(float64(total) / float64(limit))),
		ResultArray: make([]interface{}, 0),
	}

	for i := offset; i < total && len(res.ResultArray) < limit; i++ {
		res.ResultArray = append(res.ResultArray, i)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(500)
	}
}

// HandleLimitOffsetPageNullableBody enforces the nullable-body wire contract:
// the first request carries the caller's null body verbatim, later pages a
// materialized body with page >= 2. Anything else is a 400.
func HandleLimitOffsetPageNullableBody(w http.ResponseWriter, r *http.Request) {
	const limit = 7

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	page := 1
	if trimmed := strings.TrimSpace(string(body)); trimmed != "" && trimmed != "null" {
		var pagination LimitOffsetRequest
		if err := json.Unmarshal(body, &pagination); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("a non-null body must be a JSON object"))
			return
		}
		if pagination.Page < 2 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("a materialized body must advance the page past the first page"))
			return
		}
		// a numeric limit below 1 was fabricated, not supplied by the caller
		var fields map[string]interface{}
		if err := json.Unmarshal(body, &fields); err == nil {
			if limitValue, ok := fields["limit"].(float64); ok && limitValue < 1 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("a materialized body must not carry a limit the caller never supplied"))
				return
			}
		}
		page = pagination.Page
	}

	start := (page - 1) * limit

	res := PaginationResponse{
		NumPages:    int(math.Ceil(float64(total) / float64(limit))),
		ResultArray: make([]interface{}, 0),
	}

	for i := start; i < total && len(res.ResultArray) < limit; i++ {
		res.ResultArray = append(res.ResultArray, i)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		w.WriteHeader(500)
	}
}

func getValue(queryValue string, hasBody bool, paginationValue int) int {
	if hasBody {
		return paginationValue
	} else {
		value, err := strconv.Atoi(queryValue)
		if err != nil {
			return 0
		}
		return value
	}
}

func getNonNumericValue(queryValue string, hasBody bool, paginationValue string) string {
	if hasBody {
		return paginationValue
	} else {
		if queryValue == "" {
			return "-1"
		} else {
			return queryValue
		}
	}
}
