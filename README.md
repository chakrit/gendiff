# GENDIFF

[![GoDoc](https://godoc.org/github.com/chakrit/gendiff?status.svg)](https://godoc.org/github.com/chakrit/gendiff)

Simple generic diff algorithm for Go.

### GET

```sh
$ go get -v -u github.com/chakrit/gendiff
```

### TERMS

* `L` or `Left` - Values on the left side of things. The "base" values.
* `R` or `Right` - Values on the right side of things. The "new" or "changed" values.
* `Match` - Item on the left matches the one on the right.
* `Insert` - Item on the right was not present on the left, it has been "inserted".
* `Delete` - Item on the left was not present on the right, it has been "deleted".

### USE

1. Implement [`gendiff.Interface`](https://godoc.org/github.com/chakrit/gendiff#Interface)
   on the values you wish to generate diffs from.
2. Call [`gendiff.Make()`](https://godoc.org/github.com/chakrit/gendiff#Make) to
   generate the diffs.
3. Loop on the resulting [`[]gendiff.Diff`](https://godoc.org/github.com/chakrit/gendiff#Diff)
   to inspect the diff. Switch on the `Op` field to determine what the diff entry
   means.

```go
type StringCompare struct {
	LeftLines []string
	RightLines []string
}


```
