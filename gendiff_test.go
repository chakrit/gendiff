package gendiff

import (
	"fmt"
	"testing"

	r "github.com/stretchr/testify/require"
)

type testcase struct {
	left  string
	right string
	diff  []Diff
}

var _ Interface = testcase{}

func (c testcase) LeftLen() int        { return len(c.left) }
func (c testcase) RightLen() int       { return len(c.right) }
func (c testcase) Equal(l, r int) bool { return c.left[l] == c.right[r] }

func (c testcase) Name() string {
	return fmt.Sprintf("Diff %#v against %#v result in %d edits",
		c.left, c.right, len(c.diff))
}

var cases = []testcase{
	{"a", "", []Diff{
		{Delete, 0, 1, 0, 0},
	}},
	{"", "a", []Diff{
		{Insert, 0, 0, 0, 1},
	}},
	{"aaa", "bbb", []Diff{
		{Delete, 0, 3, 0, 0},
		{Insert, 3, 3, 0, 3},
	}},
	{"aBce", "acDe", []Diff{
		{Match, 0, 1, 0, 1},
		{Delete, 1, 2, 1, 1},
		{Match, 2, 3, 1, 2},
		{Insert, 3, 3, 2, 3},
		{Match, 3, 4, 3, 4},
	}},
	{"aBaCa", "aCaBa", []Diff{
		// the algorithm do a greedy match so the initial `aB` will match first
		//
		//      aB aC a
		//   aC aB    a
		//
		{Delete, 0, 2, 0, 0},
		{Match, 2, 4, 0, 2},
		{Insert, 4, 4, 2, 4},
		{Match, 4, 5, 4, 5},
	}},
	{"aaabbbccceee", "aaacccDDDeee", []Diff{
		{Match, 0, 3, 0, 3},
		{Delete, 3, 6, 3, 3},
		{Match, 6, 9, 3, 6},
		{Insert, 9, 9, 6, 9},
		{Match, 9, 12, 9, 12},
	}},
	{"bbbCCCddd", "CCCeee", []Diff{
		{Delete, 0, 3, 0, 0},
		{Match, 3, 6, 0, 3},
		{Delete, 6, 9, 3, 3},
		{Insert, 9, 9, 3, 6},
	}},
	{"AbbbbbbbCdddddddE", "bbbbbbbddddddd", []Diff{
		{Delete, 0, 1, 0, 0},
		{Match, 1, 8, 0, 7},
		{Delete, 8, 9, 7, 7},
		{Match, 9, 16, 7, 14},
		{Delete, 16, 17, 14, 14},
	}},
}

func TestMake(t *testing.T) {
	for _, test := range cases {
		t.Run(test.Name(), func(tt *testing.T) {
			r.Equal(tt, test.diff, Make(test))
		})
	}
}

const contextLen = 2

var compactCases = []testcase{
	// complete match should result in no diffs
	// since compacting would remove all meaningless matches and
	// the whole string matches
	{"abc", "abc", nil},
	{"abcdef", "abcdef", nil},

	// prefixes suffixes
	{"dddmmmm", "mmmm", []Diff{
		{Delete, 0, 3, 0, 0},
		{Match, 3, 5, 0, 2},
	}},
	{"mmmm", "mmmmddd", []Diff{
		{Match, 2, 4, 2, 4},
		{Insert, 4, 4, 4, 7},
	}},
	{"aBaCa", "aCaBa", []Diff{
		// NOTE: See same case above for explanation
		{Delete, 0, 2, 0, 0},
		{Match, 2, 4, 0, 2},
		{Insert, 4, 4, 2, 4},
		{Match, 4, 5, 4, 5},
	}},

	// longer strings
	{"axxxxxxx", "xxxxxxx", []Diff{
		{Delete, 0, 1, 0, 0},
		{Match, 1, 3, 0, 2},
	}},
	{"xxxxxxx", "xxxxxxxb", []Diff{
		{Match, 5, 7, 5, 7},
		{Insert, 7, 7, 7, 8},
	}},
	{"lllllllaaarrrrrrr", "lllllllrrrrrrr", []Diff{
		{Match, 5, 7, 5, 7},
		{Delete, 7, 10, 7, 7},
		{Match, 10, 12, 7, 9},
	}},
	{"abbbbbbbcddddddde", "bbbbbbbddddddd", []Diff{
		{Delete, 0, 1, 0, 0},
		{Match, 1, 3, 0, 2},
		{Match, 6, 8, 5, 7},
		{Delete, 8, 9, 7, 7},
		{Match, 9, 11, 7, 9},
		{Match, 14, 16, 12, 14},
		{Delete, 16, 17, 14, 14},
	}},
}

func TestCompact(t *testing.T) {
	for _, test := range compactCases {
		t.Run("Compact"+test.Name(), func(tt *testing.T) {
			r.Equal(tt, test.diff, Compact(Make(test), contextLen))
		})
	}
}
