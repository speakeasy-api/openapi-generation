//go:build !js || !wasm

package format

import (
	"context"
	"strings"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
)

// formatCS is a test helper that runs the C# formatter.
func formatCS(t *testing.T, fileName string, input []byte) string {
	t.Helper()
	SetFormattingEnabled("csharp", true)
	result, err := Format(context.Background(), types.Target{Target: "csharp"}, fileName, input)
	if err != nil {
		t.Fatalf("Format() returned error: %v", err)
	}
	return string(result)
}

// ---------------------------------------------------------------------------
// Smoke tests
// ---------------------------------------------------------------------------

func TestFormatCSharp_FormatsCSFiles(t *testing.T) {
	input := []byte("public class Test{public void Foo(){int x=1;}}")
	result := formatCS(t, "Test.cs", input)

	if result == string(input) {
		t.Error("expected formatted output to differ from input")
	}
	if !strings.Contains(result, "class Test") {
		t.Errorf("expected 'class Test' in output, got:\n%s", result)
	}
}

func TestFormatCSharp_NonCsFilePassthrough(t *testing.T) {
	csproj := []byte(`<Project Sdk="Microsoft.NET.Sdk"><PropertyGroup></PropertyGroup></Project>`)
	result := formatCS(t, "Test.csproj", csproj)
	if result != string(csproj) {
		t.Errorf("expected .csproj file to pass through unchanged.\ngot:  %q\nwant: %q", result, string(csproj))
	}
}

func TestFormatCSharp_UnityTarget(t *testing.T) {
	SetFormattingEnabled("unity", true)
	input := []byte("public class Test{public void Foo(){int x=1;}}")
	result, err := Format(context.Background(), types.Target{Target: "unity"}, "Test.cs", input)
	if err != nil {
		t.Fatalf("Format() returned error: %v", err)
	}
	if string(result) == string(input) {
		t.Error("expected unity target to trigger C# formatting")
	}
}

func TestFormatCSharp_EmptyInput(t *testing.T) {
	result := formatCS(t, "Empty.cs", []byte(""))
	if len(result) != 0 {
		t.Errorf("expected empty output for empty input, got: %q", result)
	}
}

func TestFormatCSharp_Idempotent(t *testing.T) {
	input := []byte(`using System;

namespace TestApp
{
    public class Calculator
    {
        private int value;

        public Calculator(int initial)
        {
            this.value = initial;
        }

        public int Add(int n)
        {
            return value + n;
        }
    }
}
`)

	first := formatCS(t, "Calculator.cs", input)
	second := formatCS(t, "Calculator.cs", []byte(first))

	if first != second {
		t.Errorf("formatting is not idempotent.\nfirst pass:\n%s\nsecond pass:\n%s", first, second)
	}
}

// ---------------------------------------------------------------------------
// dotnet format conventions — these should all PASS
// ---------------------------------------------------------------------------

func TestFormatCSharp_AllmanBraces(t *testing.T) {
	// dotnet format: opening brace on its own line (Allman style).
	input := []byte(`public class Test
{
    public void Foo()
    {
        if (true)
        {
            var x = 1;
        }
    }
}
`)
	result := formatCS(t, "Test.cs", input)

	lines := strings.Split(result, "\n")
	foundBraceOnOwnLine := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "{" {
			foundBraceOnOwnLine = true
			break
		}
	}
	if !foundBraceOnOwnLine {
		t.Errorf("expected Allman-style braces ('{' on own line), got:\n%s", result)
	}
}

func TestFormatCSharp_KRBracesToAllman(t *testing.T) {
	// dotnet format: K&R braces should be converted to Allman.
	input := []byte(`namespace TestApp
{
    public class Converter
    {
        public void Process()
        {
            if (Nullable.GetUnderlyingType(objectType) != null) {
                objectType = Nullable.GetUnderlyingType(objectType);
            }

            try {
                DoWork();
            } catch(Exception e) {
                Log(e);
            }
        }
    }
}
`)
	result := formatCS(t, "Converter.cs", input)

	for _, line := range strings.Split(result, "\n") {
		trimmed := strings.TrimSpace(line)
		if (strings.HasPrefix(trimmed, "if ") || trimmed == "try" || strings.HasPrefix(trimmed, "catch")) &&
			strings.HasSuffix(trimmed, "{") {
			t.Errorf("expected Allman braces, found K&R on line: %q\nfull output:\n%s", trimmed, result)
		}
	}
	if strings.Contains(result, "} catch") {
		t.Errorf("expected catch on its own line, got K&R style:\n%s", result)
	}
}

func TestFormatCSharp_TryCatchAllmanBraces(t *testing.T) {
	// dotnet format: catch/finally on their own line.
	input := []byte(`namespace TestApp
{
    public class Handler
    {
        public void Run()
        {
            try
            {
                DoWork();
            }
            catch (Exception e)
            {
                Log(e);
            }
            finally
            {
                Cleanup();
            }
        }
    }
}
`)
	result := formatCS(t, "Handler.cs", input)

	if strings.Contains(result, "} catch") || strings.Contains(result, "}catch") {
		t.Errorf("expected Allman-style catch (on own line), got K&R style:\n%s", result)
	}
	if strings.Contains(result, "} finally") || strings.Contains(result, "}finally") {
		t.Errorf("expected Allman-style finally (on own line), got K&R style:\n%s", result)
	}
}

func TestFormatCSharp_NamespaceIndentation(t *testing.T) {
	// dotnet format: contents inside namespace {} are indented.
	input := []byte(`namespace Speakeasy.OpenAPI
{
    using Newtonsoft.Json;
    using Speakeasy.OpenAPI.Utils;

    public class SimpleObject
    {
        [JsonProperty("str")]
        public string Str { get; set; } = default!;
    }
}
`)
	result := formatCS(t, "SimpleObject.cs", input)

	if !strings.Contains(result, "    public class SimpleObject") {
		t.Errorf("expected class inside namespace to be indented with 4 spaces, got:\n%s", result)
	}
	if !strings.Contains(result, "    using Newtonsoft.Json;") {
		t.Errorf("expected using inside namespace to be indented with 4 spaces, got:\n%s", result)
	}
	if !strings.Contains(result, "        public string Str") {
		t.Errorf("expected property to be at 8-space indent (namespace + class), got:\n%s", result)
	}
}

func TestFormatCSharp_MultiLineParamsPreserved(t *testing.T) {
	// dotnet format: multi-line parameter lists stay on separate lines.
	input := []byte(`namespace Speakeasy.OpenAPI
{
    public class SDKBaseException : Exception
    {
        public SDKBaseException(
            string message,
            HttpRequestMessage request,
            HttpResponseMessage response,
            string body
        ) : base(message)
        {
            Message = message;
        }
    }
}
`)
	result := formatCS(t, "SDKBaseException.cs", input)

	if !strings.Contains(result, "string message,") {
		t.Errorf("expected 'string message,' on its own line, got:\n%s", result)
	}
	if !strings.Contains(result, "HttpRequestMessage request,") {
		t.Errorf("expected 'HttpRequestMessage request,' on its own line, got:\n%s", result)
	}
}

func TestFormatCSharp_PreservesUsingStatements(t *testing.T) {
	input := []byte(`using System;
using System.Collections.Generic;
using System.Linq;

namespace TestApp
{
    public class Service
    {
        private List<string> items;
    }
}
`)
	result := formatCS(t, "Service.cs", input)

	if !strings.Contains(result, "using System;") {
		t.Errorf("expected 'using System;' to be preserved, got:\n%s", result)
	}
	if !strings.Contains(result, "using System.Collections.Generic;") {
		t.Errorf("expected 'using System.Collections.Generic;' to be preserved, got:\n%s", result)
	}
}

func TestFormatCSharp_PropertyFormatting(t *testing.T) {
	input := []byte(`public class Model
{
    public string Name { get; set; }
    public int Value { get; set; }
}
`)
	result := formatCS(t, "Model.cs", input)

	if !strings.Contains(result, "Name") {
		t.Errorf("expected property 'Name' to be preserved, got:\n%s", result)
	}
	if !strings.Contains(result, "get") || !strings.Contains(result, "set") {
		t.Errorf("expected get/set accessors to be preserved, got:\n%s", result)
	}
}

func TestFormatCSharp_TrailingWhitespace(t *testing.T) {
	// dotnet format: no trailing whitespace.
	input := []byte("public class Test   \n{   \n    public void Foo()   \n    {   \n    }   \n}   \n")
	result := formatCS(t, "Test.cs", input)

	for i, line := range strings.Split(result, "\n") {
		if line != strings.TrimRight(line, " \t") {
			t.Errorf("line %d has trailing whitespace: %q", i+1, line)
		}
	}
}

func TestFormatCSharp_ExtraSpaceRemoval(t *testing.T) {
	// dotnet format: extra whitespace between tokens is normalized.
	input := []byte(`namespace TestApp
{
    public class Converter
    {
        public bool CanConvert(System.Type objectType)
        {
            var  nullableType = Nullable.GetUnderlyingType(objectType);
            if (nullableType  !=  null)
            {
                return  nullableType.IsEnum;
            }
            return objectType.IsEnum;
        }
    }
}
`)
	result := formatCS(t, "Converter.cs", input)

	if strings.Contains(result, "var  ") {
		t.Errorf("expected extra space in 'var  nullableType' to be removed, got:\n%s", result)
	}
	if strings.Contains(result, "  !=  ") {
		t.Errorf("expected extra spaces around '!=' to be normalized, got:\n%s", result)
	}
}

func TestFormatCSharp_KeywordSpacing(t *testing.T) {
	// dotnet format: space after control flow keywords.
	input := []byte(`namespace TestApp
{
    public class Processor
    {
        public void Run()
        {
            foreach(var item in items)
            {
                if(item != null)
                {
                    Process(item);
                }
            }

            while(condition)
            {
                Step();
            }
        }
    }
}
`)
	result := formatCS(t, "Processor.cs", input)

	if strings.Contains(result, "foreach(") {
		t.Errorf("expected space after 'foreach', got:\n%s", result)
	}
	if strings.Contains(result, "if(") {
		t.Errorf("expected space after 'if', got:\n%s", result)
	}
	if strings.Contains(result, "while(") {
		t.Errorf("expected space after 'while', got:\n%s", result)
	}
}

func TestFormatCSharp_GenericTypeSpacing(t *testing.T) {
	// dotnet format: no space before generic angle brackets.
	input := []byte(`namespace TestApp
{
    public class Registry
    {
        private static readonly Dictionary<string, Color> _knownValues =
            new Dictionary<string, Color>()
            {
                ["red"] = Red,
                ["green"] = Green,
            };
    }
}
`)
	result := formatCS(t, "Registry.cs", input)

	if strings.Contains(result, "Dictionary <") {
		t.Errorf("expected no space before generic <, got:\n%s", result)
	}
}

func TestFormatCSharp_SingleLineIfPreserved(t *testing.T) {
	// dotnet format: single-line if statements (without else) stay on one line.
	input := []byte(`namespace TestApp
{
    public class Guard
    {
        public void Check(object value)
        {
            if (value == null) throw new ArgumentNullException(nameof(value));
            if (ReferenceEquals(this, value)) return;
        }
    }
}
`)
	result := formatCS(t, "Guard.cs", input)

	if !strings.Contains(result, "if (value == null) throw") {
		t.Errorf("expected single-line if/throw preserved on one line, got:\n%s", result)
	}
	if !strings.Contains(result, "if (ReferenceEquals(this, value)) return;") {
		t.Errorf("expected single-line if/return preserved on one line, got:\n%s", result)
	}
}

func TestFormatCSharp_SwitchCaseIndentation(t *testing.T) {
	// dotnet format: case labels are indented inside switch blocks.
	input := []byte(`namespace TestApp
{
    public class Handler
    {
        public void Run(string type)
        {
            switch (type)
            {
                case "json":
                    DoJson();
                    break;
                default:
                    throw new Exception("unknown");
            }
        }
    }
}
`)
	result := formatCS(t, "Handler.cs", input)

	// Case labels should be indented one level inside the switch braces.
	if !strings.Contains(result, "                case \"json\":") {
		t.Errorf("expected case labels indented inside switch, got:\n%s", result)
	}
	if !strings.Contains(result, "                default:") {
		t.Errorf("expected default label indented inside switch, got:\n%s", result)
	}
}

func TestFormatCSharp_BasicFormatting(t *testing.T) {
	input := []byte(`namespace TestNamespace{public class MyClass{public void DoSomething(string name,int value){if(value>0){Console.WriteLine(name);}else{Console.WriteLine("none");}}}}
`)
	result := formatCS(t, "MyClass.cs", input)

	if !strings.Contains(result, "    ") {
		t.Errorf("expected 4-space indentation in formatted output, got:\n%s", result)
	}
	if !strings.Contains(result, "class MyClass") {
		t.Errorf("expected 'class MyClass' in output, got:\n%s", result)
	}
	if !strings.Contains(result, "DoSomething") {
		t.Errorf("expected 'DoSomething' in output, got:\n%s", result)
	}
}

func TestFormatCSharp_RealisticSDKFile(t *testing.T) {
	input := []byte(`#nullable enable
namespace Speakeasy.OpenAPI
{
    using Newtonsoft.Json;
    using Speakeasy.OpenAPI.Utils;
    using System;
    using System.Net.Http;

    /// <summary>
    /// Base Exception for API Errors.
    /// </summary>
    public class SDKBaseException : Exception
    {
        /// <summary>
        /// Error Message
        /// </summary>
        public override string Message { get; }

        /// <summary>
        /// HTTP Request
        /// </summary>
        public HttpRequestMessage Request { get; }

        /// <summary>
        /// HTTP Response
        /// </summary>
        public HttpResponseMessage Response { get; }

        /// <summary>
        /// HTTP response body
        /// </summary>
        public string Body { get; }

        public SDKBaseException(
            string message,
            HttpRequestMessage request,
            HttpResponseMessage response,
            string body
        ) : this(message, request, response, body, null) {}

        public SDKBaseException(
            string message,
            HttpRequestMessage request,
            HttpResponseMessage response,
            string body,
            Exception? innerException
        ) : base(message, innerException)
        {
            Message = $"{message.TrimEnd('.')}. Body: {body}.";
            Request = request;
            Response = response;
            Body = body;
        }

        /// <summary>
        /// Detailed Error Message
        /// </summary>
        public override string ToString()
        {
            var innerMessage = string.IsNullOrEmpty(InnerException?.Message) ? "" : $"\n{InnerException.Message}";
            return $"Status: {Response.StatusCode}. {Message}{innerMessage}";
        }

    }
}
`)
	result := formatCS(t, "SDKBaseException.cs", input)

	if !strings.Contains(result, "    public class SDKBaseException") {
		t.Errorf("expected class indented inside namespace, got:\n%s", result)
	}
	if !strings.Contains(result, "    using Newtonsoft.Json;") {
		t.Errorf("expected using statements indented inside namespace, got:\n%s", result)
	}
	if !strings.Contains(result, "string message,") {
		t.Errorf("expected multi-line constructor params preserved, got:\n%s", result)
	}
	if !strings.Contains(result, "/// <summary>") {
		t.Errorf("expected XML doc comments preserved, got:\n%s", result)
	}
	if !strings.Contains(result, "        public override string Message") {
		t.Errorf("expected properties at 8-space indent, got:\n%s", result)
	}
}

// ---------------------------------------------------------------------------
// Known divergences from dotnet format
//
// These tests document the expected dotnet format behavior. They are skipped
// because clang-format cannot currently match these conventions. If a future
// clang-format version or config change resolves a divergence, the skip
// condition will no longer trigger and the test will start passing.
// ---------------------------------------------------------------------------

func TestFormatCSharp_ClosingParenOnOwnLine(t *testing.T) {
	// dotnet format: when parameters span multiple lines, the closing )
	// stays on its own line, aligned with the opening statement.
	//
	// DIVERGENCE: clang-format moves ) to the end of the last parameter line.
	input := []byte(`namespace TestApp
{
    public class Service
    {
        public Task<Response> DoWorkAsync(
            string param1,
            int param2,
            CancellationToken? cancellationToken = null
        );
    }
}
`)
	result := formatCS(t, "Service.cs", input)

	// Each param should still be on its own line.
	if !strings.Contains(result, "string param1,") {
		t.Errorf("expected params on separate lines, got:\n%s", result)
	}

	// dotnet format expects ) on its own line.
	hasClosingParenOnOwnLine := false
	for _, line := range strings.Split(result, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == ");" {
			hasClosingParenOnOwnLine = true
			break
		}
	}
	if !hasClosingParenOnOwnLine {
		t.Skip("known clang-format divergence: closing ) moves to last parameter line instead of its own line")
	}
}

func TestFormatCSharp_ConstructorInitializerOnClosingParenLine(t *testing.T) {
	// dotnet format: constructor initializer ": base(...)" stays on the
	// same line as the closing ), like:
	//     ) : base(message) {}
	//
	// DIVERGENCE: clang-format moves ) to the last param, then puts
	// ": base()" on a separate indented line.
	input := []byte(`namespace TestApp
{
    public class MyException : Exception
    {
        public MyException(
            string message,
            HttpRequestMessage request,
            string body
        ) : base(message) {}
    }
}
`)
	result := formatCS(t, "MyException.cs", input)

	if !strings.Contains(result, ": base(message)") {
		t.Errorf("expected ': base(message)' in output, got:\n%s", result)
	}

	// dotnet format expects ") : base(message)" on one line.
	if !strings.Contains(result, ") : base(message)") {
		t.Skip("known clang-format divergence: ') : base()' split across lines — ) moves to last param, : base() on new line")
	}
}

func TestFormatCSharp_EmptyBracesOnSameLine(t *testing.T) {
	// dotnet format: empty constructor/method bodies stay as {} on one line.
	//
	// DIVERGENCE: clang-format expands {} to { on one line, } on the next
	// (Allman style for all blocks, even empty ones).
	input := []byte(`namespace TestApp
{
    public class Derived : Base
    {
        public Derived(string msg) : base(msg) {}
    }
}
`)
	result := formatCS(t, "Derived.cs", input)

	if !strings.Contains(result, "{}") {
		t.Skip("known clang-format divergence: empty {} expanded to separate lines (Allman style applied to empty bodies)")
	}
}

func TestFormatCSharp_EnumAttributeOnSeparateLine(t *testing.T) {
	// dotnet format: [Attr] on its own line before enum value.
	//
	// DIVERGENCE: clang-format's C# parser joins short attributes onto the
	// same line as the enum member: [JsonProperty("80")] Eighty,
	input := []byte(`namespace TestApp
{
    public enum ServerPORT
    {
        [JsonProperty("80")]
        Eighty,
        [JsonProperty("8080")]
        EightThousandAndEighty,
    }
}
`)
	result := formatCS(t, "ServerPORT.cs", input)

	// dotnet format: attribute and value on separate lines.
	hasAttributeOnOwnLine := false
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[JsonProperty(") && strings.HasSuffix(trimmed, ")]") {
			// Next non-empty line should be the enum value, not on the same line.
			if i+1 < len(lines) {
				nextTrimmed := strings.TrimSpace(lines[i+1])
				if nextTrimmed != "" && !strings.HasPrefix(nextTrimmed, "[") {
					hasAttributeOnOwnLine = true
					break
				}
			}
		}
	}
	if !hasAttributeOnOwnLine {
		t.Skip("known clang-format divergence: enum [Attr] joined on same line as value instead of separate line")
	}
}

func TestFormatCSharp_NoSpaceBeforeIndexer(t *testing.T) {
	// dotnet format: no space between ) and [0] in indexer access.
	//
	// DIVERGENCE: clang-format's approximate C# parser confuses array
	// indexer access after method calls with attribute declarations,
	// inserting a spurious space: ") [0]" instead of ")[0]".
	// Upstream: https://github.com/llvm/llvm-project/issues/101487
	// (marked as fixed, but the fix did not land in LLVM 19.1.7)
	input := []byte(`namespace TestApp
{
    public static class EnumExtension
    {
        public static string Value(this MyEnum value)
        {
            return ((JsonPropertyAttribute)value.GetType().GetMember(value.ToString())[0].GetCustomAttributes(typeof(JsonPropertyAttribute), false)[0]).PropertyName ?? value.ToString();
        }
    }
}
`)
	result := formatCS(t, "EnumExtension.cs", input)

	if strings.Contains(result, ") [0]") {
		t.Skip("known clang-format divergence: spurious space before indexer ') [0]' instead of ')[0]' (upstream llvm/llvm-project#101487)")
	}
	// If we get here, the bug is fixed — verify correct form.
	if !strings.Contains(result, ")[0]") {
		t.Errorf("expected ')[0]' in output, got:\n%s", result)
	}
}
