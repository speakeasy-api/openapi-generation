package format

import (
	"bytes"
)

// NoFormatMarker is the sentinel emitted by the `skipFormat` template
// function to signal that the rendered file must skip the formatting pipeline.
// It must appear as the very first bytes of the rendered file. The marker
// (and its trailing newline, if present) is stripped before the file is
// written to disk.
const NoFormatMarker = "__SPEAKEASY_NO_FORMAT__"

func ShouldSkipFormat(data []byte) (bool, []byte) {
	if bytes.HasPrefix(data, []byte(NoFormatMarker)) {
		return true, bytes.TrimPrefix(data[len(NoFormatMarker):], []byte("\n"))
	}

	return false, data
}
