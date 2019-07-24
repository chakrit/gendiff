// Package gendiff provides a generic diff algorithm that can be applied to any
// two list of data. For example, two slices of strings or two lists of objects.
//
// To use gendiff, first create a comparison type to store both list of data:
//
//     type Comparison struct {
//         LeftLines []string
//         RightLines []string
//     }
//
// Then, implement the `gendiff.Interface` interface and call `Make` on it.
// Optionally followed by a call to `Compact` if you wish to shorten the length
// of the diffs.
//
// The resulting list of `Diff` should then be enumerated to find out what
// operations to perform on the left list to arrive at the right list.
package gendiff

import (
	"fmt"
)

// Op marks the operation being done on a diff.
type Op int

const (
	// noOp means nothing has been done, useful as sentinel values.
	noOp = Op(iota)

	// Match means all items in the range on the left and right matched.
	Match = Op(iota)

	// Delete means items on the left was missing from the right.
	// This means the items on the left have been deleted.
	Delete

	// Insert means items on the right was missing from the left.
	// This means the items on the right have been newly inserted.
	Insert
)

// String returns the name of the operation.
func (o Op) String() string {
	switch o {
	case Match:
		return "match"
	case Delete:
		return "delete"
	case Insert:
		return "insert"
	default: // noOp and others
		return ""
	}
}

// Interface allows you to use this package with any arbitrary data type.
// The `Left` items are considered as the base values while the `Right` items
// will be used as the compare values.
type Interface interface {
	LeftLen() int
	RightLen() int
	Equal(l, r int) bool
}

// Diff represents a single edit operation being done on the given list.
// The `Op` field specifies the operation and the (Lstart, Lend) and
// (Rstart, Rend) specify the start and end indexes on the left and right list
// of items, respectively.
//
// For example, `Diff{Insert, 0, 0, 0, 3}` says that something has been inserted
// into the left list at index 0 and the items that were inserted can be found
// on the right list from index 0 to 3.
type Diff struct {
	Op     Op
	Lstart int
	Lend   int
	Rstart int
	Rend   int
}

// Len returns the length of the changes. If the operation is a `Match`, this is
// the number of items in the current series of matches. If the operation is a
// `Delete`, it is the number of items being deleted from the left list. If the
// operation is `Insert`, it is the number of items being inserted from the
// right list.
//
// Len returns -1 otherwise if `Op` is not a valid value.
func (d Diff) Len() int {
	switch d.Op {
	case Match:
		return d.Lend - d.Lstart
	case Delete:
		return d.Lend - d.Lstart
	case Insert:
		return d.Rend - d.Rstart
	default:
		return -1
	}
}

type cell struct {
	op     Op  // detected operation
	length int // cumulative length of longest common substring
}

func (c cell) String() string {
	return fmt.Sprintf("%s(%3d)", c.op.String()[:1], c.length)
}

// Make creates a list of `Diff`, that is a list of operations that can be
// performed on the left list to arrive at the right list state.
//
// Internally, it uses a variant of the longest-common-substring algorithm with
// dynamic programming to obtain a set of `Match` operations. And then all the
// operations in-between the matches are considered as `Insert` if it happens on
// the right values and `Delete` if it happens on the left values.
//
// Performance should be something close to `O(NL * NR)` where NL and NR is the
// number of items on the left and right list, respectively.
func Make(iface Interface) []Diff {
	var (
		llen, rlen = iface.LeftLen(), iface.RightLen()
		lidx, ridx = 0, 0
	)

	// table for dynamic programming an LCS solution
	// "left" is Y, "right" is X,
	// so index with table[y][x] or table[left str index][right str index]
	table := make([][]cell, llen+1, llen+1)
	for lidx = range table {
		table[lidx] = make([]cell, rlen+1, rlen+1)
	}

	// zeroes the empty string solution
	// (all inserts or all deletes)
	for lidx = range table { // empty "right" string, all were deleted
		table[lidx][0] = cell{Delete, 0}
	}
	for ridx = range table[0] { // empty "left" string, all were inserted
		table[0][ridx] = cell{Insert, 0}
	}

	// start point, both string empty, they match
	table[0][0] = cell{Match, 0}

	// compute lcs solution and label the table with the right operations
	for lidx = 1; lidx <= llen; lidx++ {
		for ridx = 1; ridx <= rlen; ridx++ {
			var (
				lcell  = table[lidx][ridx-1]   // neighbor towards the "left" string (x-1)
				rcell  = table[lidx-1][ridx]   // neighbor towards the "right" string (y-1)
				lrcell = table[lidx-1][ridx-1] // diagonal neighbor (x-1, y-1)
			)

			switch {
			case iface.Equal(lidx-1, ridx-1):
				// character match, extends the lcs counter
				table[lidx][ridx] = cell{op: Match, length: lrcell.length + 1}
			case lcell.length < rcell.length:
				// the "right" string has longer lcs which means we are sitting
				// on characters being deleted from the "left" string
				table[lidx][ridx] = cell{op: Delete, length: rcell.length}
			case lcell.length >= rcell.length:
				// the "left" string has longer lcs which means we are sitting
				// on an extra characters from the "right" string
				table[lidx][ridx] = cell{op: Insert, length: lcell.length}
			}
		}
	}

	// uncomment this block to dump the table to STDOUT, for debugging
	//fmt.Println("-------------------")
	//for lidx = range table {
	//	for ridx = range table[lidx] {
	//		fmt.Printf("%s ", table[lidx][ridx])
	//	}
	//	fmt.Println()
	//}
	//fmt.Println("-------------------")

	// reconstruct solution backwards
	var (
		diffs    []Diff
		lastcell = table[llen][rlen]
		lastdiff = Diff{lastcell.op, llen, llen, rlen, rlen}
	)

	record := func(op Op, lidx, ridx int) {
		lastdiff.Lstart = lidx
		lastdiff.Rstart = ridx
		if op != lastdiff.Op {
			diffs = append(diffs, lastdiff)
			lastdiff.Op = op
			lastdiff.Lend = lastdiff.Lstart
			lastdiff.Rend = lastdiff.Rstart
		}
	}

	lidx, ridx = llen, rlen
	for lidx > 0 || ridx > 0 {
		cell := table[lidx][ridx]
		record(cell.op, lidx, ridx)

		switch cell.op {
		case Match:
			lidx, ridx = lidx-1, ridx-1
		case Delete:
			lidx, ridx = lidx-1, ridx
		case Insert:
			lidx, ridx = lidx, ridx-1
		default:
			panic("DP table construction error, please file a bug report.")
		}
	}

	record(noOp, 0, 0) // eof signal to emit the last diff

	// since we construct solution backwards, we need to reverse it
	revdiffs := make([]Diff, len(diffs), len(diffs))
	for idx := range diffs {
		revdiffs[len(diffs)-idx-1] = diffs[idx]
	}
	return revdiffs
}

// Compact compacts the given list of Diffs such that a `Match`	longer than the
// specified context length are trimmed to the context length. This is useful
// if your application consider a long series of `Match` operations to be noise
// and want to visually focus on the actual changes rather than the matches.
//
// Compact will return `nil` if no modifications exist inside the given list of
// diffs. For example, running `Compact(diffs, 2)` on a series of
//
//     MMMMIIIMMMMDDDMMMM
//
// Where M=Match, I=Insert and D=Delete will result in the following:
//
//     MMIIIMM MMDDDMM
//
// The space in-between represents gaps in the index fields. You should handle
// them accordingly (i.e. print a "skip" line, or advance related indexes.)
func Compact(diffs []Diff, contextLen int) []Diff {
	prefix := func(match Diff) Diff {
		return Diff{
			Op:     Match,
			Lstart: match.Lstart,
			Lend:   match.Lstart + contextLen,
			Rstart: match.Rstart,
			Rend:   match.Rstart + contextLen,
		}
	}
	suffix := func(match Diff) Diff {
		return Diff{
			Op:     Match,
			Lstart: match.Lend - contextLen,
			Lend:   match.Lend,
			Rstart: match.Rend - contextLen,
			Rend:   match.Rend,
		}
	}

	// short special cases
	switch len(diffs) {
	case 0:
		return diffs

	case 1:
		if diffs[0].Op == Match {
			return nil
		} else {
			return diffs
		}

	case 2:
		if diffs[0].Op == Match && diffs[0].Len() > contextLen {
			return []Diff{suffix(diffs[0]), diffs[1]}
		} else if diffs[1].Op == Match && diffs[1].Len() > contextLen {
			return []Diff{diffs[0], prefix(diffs[1])}
		} else {
			return diffs
		}

	default:
		// invariant: len(diffs) >= 3
	}

	var (
		out []Diff

		first  = diffs[0]
		middle = diffs[1 : len(diffs)-1]
		last   = diffs[len(diffs)-1]
	)

	// first elem
	if first.Op == Match && first.Len() > contextLen {
		out = append(out, suffix(first))
	} else {
		out = append(out, first)
	}

	// in-between elems
	for _, d := range middle {
		if d.Op == Match && d.Len() > (contextLen*2) {
			out = append(out, prefix(d), suffix(d))
		} else {
			out = append(out, d)
		}
	}

	// last elem
	if last.Op == Match && last.Len() > contextLen {
		out = append(out, prefix(last))
	} else {
		out = append(out, last)
	}

	return out
}
