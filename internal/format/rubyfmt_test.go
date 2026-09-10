//go:build !js || !wasm

package format

import (
	"context"
	"strings"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
)

func TestFormatRuby_BasicFormatting(t *testing.T) {
	SetFormattingEnabled("ruby", true)
	defer SetFormattingEnabled("ruby", false)

	input := []byte(`def hello( name )
  puts "Hello, #{  name  }!"
  if true
    x = 1+2
  end
end
`)

	target := types.Target{Target: "ruby"}
	out, err := Format(context.Background(), target, "test.rb", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outStr := string(out)
	if !strings.Contains(outStr, "def hello(name)") {
		t.Errorf("expected parens normalized, got:\n%s", outStr)
	}
	if !strings.Contains(outStr, "x = 1 + 2") {
		t.Errorf("expected operator spacing, got:\n%s", outStr)
	}
}

func TestFormatRuby_NonRbFilePassthrough(t *testing.T) {
	SetFormattingEnabled("ruby", true)
	defer SetFormattingEnabled("ruby", false)

	input := []byte("not ruby content")
	target := types.Target{Target: "ruby"}

	out, err := Format(context.Background(), target, "Gemfile", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(out) != string(input) {
		t.Fatalf("expected non-.rb files to pass through unchanged")
	}
}

func TestFormatRuby_SyntaxError(t *testing.T) {
	SetFormattingEnabled("ruby", true)
	defer SetFormattingEnabled("ruby", false)

	input := []byte(`def broken(
  end end end
`)

	target := types.Target{Target: "ruby"}
	out, err := Format(context.Background(), target, "broken.rb", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(out) != string(input) {
		t.Fatalf("expected syntax error to return original data unchanged")
	}
}

func TestFormatRuby_ClassDefinition(t *testing.T) {
	SetFormattingEnabled("ruby", true)
	defer SetFormattingEnabled("ruby", false)

	input := []byte(`class MyClass
  attr_reader :name,  :age

  def initialize( name,age )
    @name = name
    @age=age
  end

  def to_s
    "#{@name} (#{@age})"
  end
end
`)

	target := types.Target{Target: "ruby"}
	out, err := Format(context.Background(), target, "my_class.rb", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outStr := string(out)
	if !strings.Contains(outStr, "def initialize(name, age)") {
		t.Errorf("expected param formatting, got:\n%s", outStr)
	}
	if !strings.Contains(outStr, "@age = age") {
		t.Errorf("expected assignment spacing, got:\n%s", outStr)
	}
}

func TestFormatRuby_IndentationFix(t *testing.T) {
	SetFormattingEnabled("ruby", true)
	defer SetFormattingEnabled("ruby", false)

	input := []byte(`class Foo
def bar
puts "hello"
end
end
`)

	target := types.Target{Target: "ruby"}
	out, err := Format(context.Background(), target, "indent.rb", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `class Foo
  def bar
    puts("hello")
  end
end
`
	if string(out) != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, string(out))
	}
}

func TestFormatRuby_TrailingWhitespace(t *testing.T) {
	SetFormattingEnabled("ruby", true)
	defer SetFormattingEnabled("ruby", false)

	input := []byte("def foo   \n  bar   \nend   \n")

	target := types.Target{Target: "ruby"}
	out, err := Format(context.Background(), target, "trailing.rb", input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(string(out), "   \n") {
		t.Errorf("expected trailing whitespace removed, got:\n%q", string(out))
	}
}
