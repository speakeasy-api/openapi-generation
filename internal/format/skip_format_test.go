//go:build !js || !wasm

package format

import (
	"context"
	"strings"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
)

func TestFormat_NoFormatMarker_StripsPrefixAndNewline(t *testing.T) {
	target := types.Target{Target: "typescript"}
	input := []byte(NoFormatMarker + "\nexport   const   x=1;\n")

	got, err := Format(context.Background(), target, "x.ts", input)
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	want := "export   const   x=1;\n"
	if string(got) != want {
		t.Fatalf("Format mismatch\nwant: %q\ngot:  %q", want, string(got))
	}
	if strings.Contains(string(got), NoFormatMarker) {
		t.Fatalf("marker leaked into output: %q", string(got))
	}
}

func TestFormat_NoFormatMarker_SkipsFormatterEvenForFormattableFile(t *testing.T) {
	target := types.Target{Target: "typescript"}
	ugly := "export   const   x=1;\n"
	input := []byte(NoFormatMarker + "\n" + ugly)

	got, err := Format(context.Background(), target, "x.ts", input)
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	if string(got) != ugly {
		t.Fatalf("expected formatter to be skipped; got reformatted output: %q", string(got))
	}
}

func TestFormat_NoFormatMarker_NoTrailingNewline(t *testing.T) {
	target := types.Target{Target: "typescript"}
	input := []byte(NoFormatMarker + "rest")

	got, err := Format(context.Background(), target, "x.ts", input)
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}

	if string(got) != "rest" {
		t.Fatalf("want %q got %q", "rest", string(got))
	}
}

func TestFormat_NoFormatMarker_OnlyHonouredAtPrefix(t *testing.T) {
	target := types.Target{Target: "typescript"}
	// marker appears mid-file, not at byte 0 — must not be stripped, and the
	// formatter must still run.
	input := []byte("export const x = 1;\n" + NoFormatMarker + "\n")

	got, err := Format(context.Background(), target, "x.ts", input)
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}
	if !strings.Contains(string(got), NoFormatMarker) {
		t.Fatalf("mid-file marker should not have been stripped: %q", string(got))
	}
}

func TestFormat_NoMarker_FormatterRuns(t *testing.T) {
	target := types.Target{Target: "typescript"}
	ugly := []byte("export   const   x=1;\n")

	got, err := Format(context.Background(), target, "x.ts", ugly)
	if err != nil {
		t.Fatalf("Format returned error: %v", err)
	}
	if string(got) == string(ugly) {
		t.Fatalf("expected dprint to reformat ugly input; got identical bytes: %q", string(got))
	}
}
