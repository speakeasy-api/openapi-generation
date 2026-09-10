package contenttypes

import (
	"regexp"
	"slices"
	"strings"
)

var (
	jsonEncodingRegex       = regexp.MustCompile(`^(application|text)\/([^+]+\+)*json.*`)
	xmlEncodingRegex        = regexp.MustCompile(`^(application|text)\/([^+]+\+)*xml.*`)
	yamlEncodingRegex       = regexp.MustCompile(`^(application|text)\/([^+]+\+)*ya?ml.*`)
	csvEncodingRegex        = regexp.MustCompile(`^(application|text)\/([^+]+\+)*csv.*`)
	multipartEncodingRegex  = regexp.MustCompile(`^multipart\/.*`)
	urlEncodedEncodingRegex = regexp.MustCompile(`^application\/x-www-form-urlencoded.*`)
	textPlainRegex          = regexp.MustCompile(`^text\/plain(\+.+)?`)
	eventStreamRegex        = regexp.MustCompile(`^text\/event-stream(\+.+)?`)
	// JSONL, X-NDJSON are both supported by the generator.
	// X-NDJSON is treated as JSONL for the purposes to sdk generation
	jsonLRegex       = regexp.MustCompile(`^(application|text)\/(([^+]+\+)*jsonl\b.*|([^+]+\+)*x-ndjson\b.*)`)
	jsonSeqRegex     = regexp.MustCompile(`^application\/([^+]+\+)*json-seq\b.*`)
	octetStreamRegex = regexp.MustCompile(`^application\/octet-stream.*`)
)

func IsJSON(contentType string) bool {
	return jsonEncodingRegex.MatchString(contentType) && !IsJsonL(contentType) && !IsJsonSeq(contentType)
}

func IsXML(contentType string) bool {
	return xmlEncodingRegex.MatchString(contentType)
}

func IsYAML(contentType string) bool {
	return yamlEncodingRegex.MatchString(contentType)
}

func IsCSV(contentType string) bool {
	return csvEncodingRegex.MatchString(contentType)
}

func IsMultipart(contentType string) bool {
	return multipartEncodingRegex.MatchString(contentType)
}

func IsURLEncoded(contentType string) bool {
	return urlEncodedEncodingRegex.MatchString(contentType)
}

func IsTextPlain(contentType string) bool {
	return textPlainRegex.MatchString(contentType)
}

func IsEventStream(contentType string) bool {
	return eventStreamRegex.MatchString(contentType)
}

func IsJsonL(contentType string) bool {
	return jsonLRegex.MatchString(contentType)
}

func IsJsonSeq(contentType string) bool {
	return jsonSeqRegex.MatchString(contentType)
}

func IsOctetStream(contentType string) bool {
	return octetStreamRegex.MatchString(contentType)
}

func DeDupeAcceptTypes(acceptTypes []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range acceptTypes {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func SortAcceptTypes(acceptTypes []string) []string {
	// sorts the accept types so the most specific types are first (e.g. application/json before application/*) and then the order is json, yaml, csv, text and then everything else
	slices.SortStableFunc(acceptTypes, func(a string, b string) int {
		switch {
		case a == "*/*":
			return 1
		case b == "*/*":
			return -1
		case IsJSON(a) && !IsJSON(b):
			return -1
		case !IsJSON(a) && IsJSON(b):
			return 1
		case IsJSON(a) && IsJSON(b):
			return compareContentTypeSpecifity(a, b)
		case IsYAML(a) && !IsYAML(b):
			return -1
		case !IsYAML(a) && IsYAML(b):
			return 1
		case IsYAML(a) && IsYAML(b):
			return compareContentTypeSpecifity(a, b)
		case IsCSV(a) && !IsCSV(b):
			return -1
		case !IsCSV(a) && IsCSV(b):
			return 1
		case IsCSV(a) && IsCSV(b):
			return compareContentTypeSpecifity(a, b)
		case IsTextPlain(a) && !IsTextPlain(b):
			return -1
		case !IsTextPlain(a) && IsTextPlain(b):
			return 1
		default:
			return compareContentTypeSpecifity(a, b)
		}
	})

	return acceptTypes
}

func compareContentTypeSpecifity(a string, b string) int {
	if strings.HasSuffix(a, "/*") && !strings.HasSuffix(b, "/*") {
		return 1
	} else if !strings.HasSuffix(a, "/*") && strings.HasSuffix(b, "/*") || strings.HasSuffix(a, "/*") && strings.HasSuffix(b, "/*") {
		return -1
	}

	if len(strings.Split(a, ";")) > 1 {
		return -1
	} else if len(strings.Split(b, ";")) > 1 {
		return 1
	}

	switch {
	case a == b:
		return 0
	case a < b:
		return -1
	default:
		return 1
	}
}
