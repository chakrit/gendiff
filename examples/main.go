package main

import (
	"fmt"
	"github.com/chakrit/gendiff"
)

type StringCompare struct {
	LeftLines  []string
	RightLines []string
}

func (c *StringCompare) LeftLen() int        { return len(c.LeftLines) }
func (c *StringCompare) RightLen() int       { return len(c.RightLines) }
func (c *StringCompare) Equal(l, r int) bool { return c.LeftLines[l] == c.RightLines[r] }

func main() {
	compare := &StringCompare{
		LeftLines: []string{
			"the",
			"quick",
			"brown",
			"fox",
			"jumps",
			"over",
			"the",
			"lazy",
			"dog",
		},
		RightLines: []string{
			"the",
			"quick",
			"brown",
			"dog",
			"jumps",
			"over",
			"the",
			"lazy",
			"fox",
		},
	}

	diffs := gendiff.Make(compare)
	for _, d := range diffs {
		switch d.Op {
		case gendiff.Match:
			for i := d.Lstart; i < d.Lend; i++ {
				fmt.Println("    " + compare.LeftLines[i])
			}

		case gendiff.Delete:
			for i := d.Lstart; i < d.Lend; i++ {
				fmt.Println("--- " + compare.LeftLines[i])
			}

		case gendiff.Insert:
			for i := d.Rstart; i < d.Rend; i++ {
				fmt.Println("+++ " + compare.RightLines[i])
			}
		}
	}
}
