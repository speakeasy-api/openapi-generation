//go:build !js || !wasm

package format

import (
	"context"
	"strings"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
)

func formatJavaDirect(t *testing.T, input string) string {
	t.Helper()
	ensureDPrint()
	result, err := javaFormatter.Format(context.Background(), "Test.java", []byte(input))
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}
	return string(result)
}

func TestFormatJavaFeatureGateDisabled(t *testing.T) {
	// When the java feature gate is not enabled, Format() should return data unchanged.
	SetFormattingEnabled("java", false)
	defer SetFormattingEnabled("java", false)

	input := []byte("public class Test{public void foo(){int x=1;}}")
	target := types.Target{Target: "java"}

	result, err := Format(context.Background(), target, "Test.java", input)
	if err != nil {
		t.Fatalf("Format() returned error: %v", err)
	}

	if string(result) != string(input) {
		t.Errorf("expected data to be returned unchanged when feature gate is disabled.\ngot:  %q\nwant: %q", string(result), string(input))
	}
}

func TestFormatJavaResilientToErrors(t *testing.T) {
	// When the formatter encounters invalid Java, Format() should return the
	// original data unchanged rather than an error, so that generation is not blocked.
	SetFormattingEnabled("java", true)
	defer SetFormattingEnabled("java", false)

	input := []byte("this is not valid java {{{")
	target := types.Target{Target: "java"}

	result, err := Format(context.Background(), target, "Test.java", input)
	if err != nil {
		t.Fatalf("Format() should not return an error on invalid input, got: %v", err)
	}

	if string(result) != string(input) {
		t.Errorf("expected original data to be returned on format failure.\ngot:  %q\nwant: %q", string(result), string(input))
	}
}

func TestFormatJavaFeatureGateEnabled(t *testing.T) {
	// When the java feature gate is enabled, Format() should return formatted data.
	SetFormattingEnabled("java", true)
	defer SetFormattingEnabled("java", false)

	input := []byte("public class Test{public void foo(){int x=1;}}")
	target := types.Target{Target: "java"}

	result, err := Format(context.Background(), target, "Test.java", input)
	if err != nil {
		t.Fatalf("Format() returned error: %v", err)
	}

	resultStr := string(result)

	// The formatted output should differ from the input (it should have proper spacing/indentation).
	if resultStr == string(input) {
		t.Error("expected formatted output to differ from input when feature gate is enabled")
	}

	// Basic sanity: the result should contain the class name and method.
	if !strings.Contains(resultStr, "class Test") {
		t.Errorf("expected formatted output to contain 'class Test', got:\n%s", resultStr)
	}
}

func TestFormatJavaNonJavaFilePassthrough(t *testing.T) {
	// Non-.java files should pass through unchanged even when the feature gate is on.
	SetFormattingEnabled("java", true)
	defer SetFormattingEnabled("java", false)

	target := types.Target{Target: "java"}

	xmlInput := []byte("<project><modelVersion>4.0.0</modelVersion></project>")
	result, err := Format(context.Background(), target, "pom.xml", xmlInput)
	if err != nil {
		t.Fatalf("Format() returned error for .xml file: %v", err)
	}
	if string(result) != string(xmlInput) {
		t.Errorf("expected .xml file to pass through unchanged.\ngot:  %q\nwant: %q", string(result), string(xmlInput))
	}

	gradleInput := []byte("apply plugin: 'java'")
	result, err = Format(context.Background(), target, "build.gradle", gradleInput)
	if err != nil {
		t.Fatalf("Format() returned error for .gradle file: %v", err)
	}
	if string(result) != string(gradleInput) {
		t.Errorf("expected .gradle file to pass through unchanged.\ngot:  %q\nwant: %q", string(result), string(gradleInput))
	}
}

func TestFormatJavaBasicFormatting(t *testing.T) {
	input := "public class Test{public void foo(){int x=1;}}"
	result := formatJavaDirect(t, input)

	// Should have proper indentation (4 spaces per the config).
	if !strings.Contains(result, "    ") {
		t.Errorf("expected 4-space indentation in formatted output, got:\n%s", result)
	}

	// Should have space before opening brace.
	if !strings.Contains(result, "class Test {") {
		t.Errorf("expected 'class Test {' with space before brace, got:\n%s", result)
	}

	// Should have the method on its own line with indentation.
	lines := strings.Split(result, "\n")
	foundMethod := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "public void foo()") {
			foundMethod = true
			// Verify it is indented.
			if !strings.HasPrefix(line, "    ") {
				t.Errorf("expected method to be indented with 4 spaces, got: %q", line)
			}
			break
		}
	}
	if !foundMethod {
		t.Errorf("expected to find 'public void foo()' in formatted output, got:\n%s", result)
	}
}

func TestFormatJavaNestedIndentation(t *testing.T) {
	input := `public class Outer{public class Inner{public void doStuff(){if(true){System.out.println("hello");}}}}` + "\n"
	result := formatJavaDirect(t, input)

	lines := strings.Split(result, "\n")

	// Verify multiple indentation levels exist (4, 8, 12 spaces).
	indentLevels := map[int]bool{}
	for _, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}
		spaces := len(line) - len(strings.TrimLeft(line, " "))
		if spaces > 0 {
			indentLevels[spaces] = true
		}
	}

	// We expect at least indentation at 4 and 8 spaces (class body and inner class body).
	if !indentLevels[4] {
		t.Errorf("expected 4-space indentation level, found levels: %v\nformatted:\n%s", indentLevels, result)
	}
	if !indentLevels[8] {
		t.Errorf("expected 8-space indentation level, found levels: %v\nformatted:\n%s", indentLevels, result)
	}
}

func TestFormatJavaTrailingWhitespaceRemoval(t *testing.T) {
	input := "public class Test {   \n    public void foo() {   \n        int x = 1;   \n    }   \n}   \n"
	result := formatJavaDirect(t, input)

	lines := strings.Split(result, "\n")
	for i, line := range lines {
		if line != strings.TrimRight(line, " \t") {
			t.Errorf("line %d has trailing whitespace: %q", i+1, line)
		}
	}
}

func TestFormatJavaImportFormatting(t *testing.T) {
	input := `import java.util.List;
import java.util.Map;

public class Test {
    private List<String> items;
    private Map<String, Integer> counts;
}
`
	result := formatJavaDirect(t, input)

	// Imports should still be present and properly formatted.
	if !strings.Contains(result, "import java.util.List;") {
		t.Errorf("expected import statement for List to be present, got:\n%s", result)
	}
	if !strings.Contains(result, "import java.util.Map;") {
		t.Errorf("expected import statement for Map to be present, got:\n%s", result)
	}

	// Imports should appear before the class definition.
	importIdx := strings.Index(result, "import java.util.List;")
	classIdx := strings.Index(result, "public class Test")
	if importIdx >= classIdx {
		t.Errorf("expected imports to appear before the class definition, got:\n%s", result)
	}
}

func TestFormatJavaMethodChainShortStaysOnOneLine(t *testing.T) {
	input := `public class Test {
    public void foo() {
        var x = builder.build();
    }
}
`
	result := formatJavaDirect(t, input)

	// Short chain call should remain on one line.
	if !strings.Contains(result, "builder.build()") {
		t.Errorf("expected short chain 'builder.build()' to stay on one line, got:\n%s", result)
	}
}

func TestFormatJavaMethodChainLongBreaks(t *testing.T) {
	input := `public class Test {
    public void foo() {
        var result = SomeVeryLongBuilderClassName.newBuilder().setFirstProperty("value1").setSecondProperty("value2").setThirdProperty("value3").setFourthProperty("value4").build();
    }
}
`
	result := formatJavaDirect(t, input)

	// A long chain should be broken across multiple lines. Count the lines
	// that contain a dot-prefixed method call.
	lines := strings.Split(result, "\n")
	dotCallLines := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, ".set") || strings.HasPrefix(trimmed, ".build") || strings.HasPrefix(trimmed, ".new") {
			dotCallLines++
		}
	}

	if dotCallLines < 2 {
		t.Errorf("expected long method chain to be broken across multiple lines (found %d dot-call lines), got:\n%s", dotCallLines, result)
	}
}

func TestFormatJavaIdempotent(t *testing.T) {
	input := `public class Test {

    private int value;

    public Test(int value) {
        this.value = value;
    }

    public int getValue() {
        return value;
    }
}
`
	first := formatJavaDirect(t, input)
	second := formatJavaDirect(t, first)

	if first != second {
		t.Errorf("formatting is not idempotent.\nfirst pass:\n%s\nsecond pass:\n%s", first, second)
	}
}

func TestFormatJavaNoLeadingCommas(t *testing.T) {
	// Verify the full Format pipeline (with pre/post-processing) does not
	// produce lines starting with a comma -- a known dprint bug with
	// trailing "//" comments and inline comments in argument lists.
	SetFormattingEnabled("java", true)
	defer SetFormattingEnabled("java", false)
	ensureDPrint()
	target := types.Target{Target: "java"}

	input := `public class Test {
    public void foo() {
        globals.pathParamsAsStream()
                .filter(entry -> !pathParams.containsKey(entry.getKey()))
                .forEach(entry -> pathParams.put(entry.getKey(), //
                        pathEncode(entry.getValue(), false)));
    }

    private static final Map<Class<?>, Object> CONVERSIONS = Map.of(//
            BigInteger.class, o -> new BigIntegerString((BigInteger) o), //
            BigDecimal.class, o -> new BigDecimalString((BigDecimal) o));

    public void bar() {
        request.bodyPublisher().ifPresentOrElse(
                // if body is present, set it
                bodyPublisher -> builder.method(method, bodyPublisher),
                // otherwise, use empty body
                () -> builder.GET());
    }
}
`
	result, err := Format(context.Background(), target, "Test.java", []byte(input))
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	resultStr := string(result)
	for i, line := range strings.Split(resultStr, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, ",") {
			t.Errorf("line %d starts with comma (formatter workaround failed): %q\nfull output:\n%s", i+1, line, resultStr)
		}
	}
}

func TestStripTrailingEmptyComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "bare trailing //",
			input: "foo.bar() //\n    .baz();",
			want:  "foo.bar()\n    .baz();",
		},
		{
			name:  "trailing // with spaces",
			input: "foo.bar() //  \n    .baz();",
			want:  "foo.bar()\n    .baz();",
		},
		{
			name:  "real comment preserved",
			input: "foo.bar() // real comment\n    .baz();",
			want:  "foo.bar() // real comment\n    .baz();",
		},
		{
			name:  "no comment unchanged",
			input: "foo.bar()\n    .baz();",
			want:  "foo.bar()\n    .baz();",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := stripTrailingEmptyComments(tc.input)
			if got != tc.want {
				t.Errorf("stripTrailingEmptyComments(%q):\n  got:  %q\n  want: %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestFixCommaAfterComment(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "spurious comma after comment removed",
			input: "        // some comment\n        ,\n        arg",
			want:  "        // some comment\n        arg",
		},
		{
			name:  "no comma after comment unchanged",
			input: "        // some comment\n        arg",
			want:  "        // some comment\n        arg",
		},
		{
			name:  "normal comma not affected",
			input: "        arg1,\n        arg2",
			want:  "        arg1,\n        arg2",
		},
		{
			name:  "comma on line with other content not affected",
			input: "        // some comment\n        , arg2\n        arg3",
			want:  "        // some comment\n        , arg2\n        arg3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := fixCommaAfterComment(tc.input)
			if got != tc.want {
				t.Errorf("fixCommaAfterComment():\n  got:  %q\n  want: %q", got, tc.want)
			}
		})
	}
}
