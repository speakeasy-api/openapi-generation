package namer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bits-and-blooms/bitset"
)

// stringSet is a custom type for representing a stringSet of strings.
type stringSet struct {
	bitSet        *bitset.BitSet
	stringToIndex *map[string]int
	indexToString *[]string
}

// Set label in the set
func (s *stringSet) Set(label string) {
	index := (*s.stringToIndex)[label]
	s.bitSet.Set(uint(index))
}

// Clear label from the set
func (s *stringSet) Clear(label string) {
	index := (*s.stringToIndex)[label]
	s.bitSet.Clear(uint(index))
}

// Check if label is in the set
func (s *stringSet) Test(label string) bool {
	index := (*s.stringToIndex)[label]
	return s.bitSet.Test(uint(index))
}

func (s *stringSet) IsSuperSet(other *stringSet) bool {
	return s.bitSet.IsSuperSet(other.bitSet)
}

func (s *stringSet) IsSubSet(other *stringSet) bool {
	return other.IsSuperSet(s)
}

func (s *stringSet) Count() int {
	return int(s.bitSet.Count())
}

func (s *stringSet) Clone() stringSet {
	return stringSet{
		bitSet:        s.bitSet.Clone(),
		stringToIndex: s.stringToIndex,
		indexToString: s.indexToString,
	}
}

// Equal returns true if s and other contain exactly the same elements.
func (s *stringSet) Equal(other *stringSet) bool {
	return s.bitSet.Equal(other.bitSet)
}

func (s *stringSet) String() string {
	slice := s.AsSlice()
	return strings.Join(slice, ",")
}

func (s *stringSet) AsSlice() []string {
	slice := make([]string, s.Count())
	sliceUint := make([]uint, s.Count())
	for i, index := range s.bitSet.AsSlice(sliceUint) {
		slice[i] = (*s.indexToString)[index]
	}
	sort.Strings(slice)
	return slice
}

func StringSetFrom(elements []string, stringToIndex *map[string]int, indexToString *[]string) stringSet {
	s := stringSet{
		bitSet:        bitset.New(uint(len(elements))),
		stringToIndex: stringToIndex,
		indexToString: indexToString,
	}
	for _, element := range elements {
		index, ok := (*s.stringToIndex)[element]
		if !ok {
			panic(fmt.Sprintf("element %s not found in stringToIndex", element))
		}
		s.bitSet.Set(uint(index))
	}
	return s
}
