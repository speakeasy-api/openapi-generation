package namer

import (
	"fmt"
	"reflect"
	"testing"
)

var stringToIndex = map[string]int{
	"test":     0,
	"another":  1,
	"":         2,
	"existing": 3,
	"missing":  4,
	"a":        5,
	"b":        6,
	"c":        7,
}

var indexToString = []string{
	"test",
	"another",
	"",
	"existing",
	"missing",
	"a",
	"b",
	"c",
}

func TestStringSetAdd(t *testing.T) {
	s := StringSetFrom([]string{"test"}, &stringToIndex, &indexToString)
	s.Set("test")
	if !s.Test("test") {
		t.Error("Set failed: expected 'test' to be in the set")
	}

	s.Set("another")
	if !s.Test("another") {
		t.Error("Set failed: expected 'another' to be in the set")
	}

	// Test adding empty string
	s.Set("")
	if !s.Test("") {
		t.Error("Set failed: empty string not added")
	}
}

func TestStringSetHas(t *testing.T) {
	fmt.Println("TestStringSetHas")
	s := StringSetFrom([]string{"existing", ""}, &stringToIndex, &indexToString)

	fmt.Println("s", s.String())

	if !s.Test("existing") {
		t.Error("Test failed: expected 'existing' to be present")
	}

	if s.Test("missing") {
		t.Error("Test failed: 'missing' should not be present")
	}

	if !s.Test("") {
		t.Error("Test failed: empty string should be present")
	}
}

func TestStringSetIsSubset(t *testing.T) {
	tests := []struct {
		name  string
		s     stringSet
		other stringSet
		want  bool
	}{
		{
			name:  "both empty",
			s:     StringSetFrom([]string{}, &stringToIndex, &indexToString),
			other: StringSetFrom([]string{}, &stringToIndex, &indexToString),
			want:  true,
		},
		{
			name:  "s empty, other not",
			s:     StringSetFrom([]string{}, &stringToIndex, &indexToString),
			other: StringSetFrom([]string{"test"}, &stringToIndex, &indexToString),
			want:  true,
		},
		{
			name:  "s is subset",
			s:     StringSetFrom([]string{"test"}, &stringToIndex, &indexToString),
			other: StringSetFrom([]string{"test", "another"}, &stringToIndex, &indexToString),
			want:  true,
		},
		{
			name:  "s not subset",
			s:     StringSetFrom([]string{"a"}, &stringToIndex, &indexToString),
			other: StringSetFrom([]string{"test"}, &stringToIndex, &indexToString),
			want:  false,
		},
		{
			name:  "subset with empty string",
			s:     StringSetFrom([]string{""}, &stringToIndex, &indexToString),
			other: StringSetFrom([]string{"", "a"}, &stringToIndex, &indexToString),
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.IsSubSet(&tt.other)
			if got != tt.want {
				t.Errorf("IsSubSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringSetEquals(t *testing.T) {
	tests := []struct {
		name  string
		s     stringSet
		other stringSet
		want  bool
	}{
		{
			name:  "equal sets",
			s:     StringSetFrom([]string{"a", "b"}, &stringToIndex, &indexToString),
			other: StringSetFrom([]string{"b", "a"}, &stringToIndex, &indexToString),
			want:  true,
		},
		{
			name:  "different lengths",
			s:     StringSetFrom([]string{"a"}, &stringToIndex, &indexToString),
			other: StringSetFrom([]string{"a", "b"}, &stringToIndex, &indexToString),
			want:  false,
		},
		{
			name:  "same length, different elements",
			s:     StringSetFrom([]string{"a", "c"}, &stringToIndex, &indexToString),
			other: StringSetFrom([]string{"a", "b"}, &stringToIndex, &indexToString),
			want:  false,
		},
		{
			name:  "empty sets",
			s:     StringSetFrom([]string{}, &stringToIndex, &indexToString),
			other: StringSetFrom([]string{}, &stringToIndex, &indexToString),
			want:  true,
		},
		{
			name:  "one empty, one not",
			s:     StringSetFrom([]string{}, &stringToIndex, &indexToString),
			other: StringSetFrom([]string{"a"}, &stringToIndex, &indexToString),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Equal(&tt.other)
			if got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringSetString(t *testing.T) {
	tests := []struct {
		name string
		s    stringSet
		want string
	}{
		{
			name: "empty set",
			s:    StringSetFrom([]string{}, &stringToIndex, &indexToString),
			want: "",
		},
		{
			name: "single element",
			s:    StringSetFrom([]string{"a"}, &stringToIndex, &indexToString),
			want: "a",
		},
		{
			name: "multiple elements",
			s:    StringSetFrom([]string{"b", "a"}, &stringToIndex, &indexToString),
			want: "a,b",
		},
		{
			name: "with empty string",
			s:    StringSetFrom([]string{"", "a"}, &stringToIndex, &indexToString),
			want: ",a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStringSetSetFromSlice(t *testing.T) {
	tests := []struct {
		name      string
		slice     []string
		wantCount int
	}{
		{
			name:  "empty slice",
			slice: []string{},
		},
		{
			name:  "nil slice",
			slice: nil,
		},
		{
			name:      "duplicates",
			slice:     []string{"test", "test", "another"},
			wantCount: 2,
		},
		{
			name:      "unsorted",
			slice:     []string{"another", "test"},
			wantCount: 2,
		},
		{
			name:      "count 3",
			slice:     []string{"", "a", "b"},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringSetFrom(tt.slice, &stringToIndex, &indexToString)
			if got.Count() != tt.wantCount {
				t.Errorf("Count() = %v, want %v", got.Count(), tt.wantCount)
			}
		})
	}
}

func TestStringSetAsSlice(t *testing.T) {
	s := StringSetFrom([]string{"test", "another"}, &stringToIndex, &indexToString)
	if s.Count() != 2 {
		t.Errorf("Count() = %v, want %v", s.Count(), 2)
	}

	got := s.AsSlice()
	if !reflect.DeepEqual(got, []string{"another", "test"}) {
		t.Errorf("AsSlice() = %v, want %v", got, []string{"another", "test"})
	}
}
