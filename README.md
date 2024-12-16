An Adaptive Radix Tree Implementation in Go
====

[![Coverage Status](https://coveralls.io/repos/github/plar/go-adaptive-radix-tree/badge.svg?branch=master&v=1)](https://coveralls.io/github/plar/go-adaptive-radix-tree?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/plar/go-adaptive-radix-tree)](https://goreportcard.com/report/github.com/plar/go-adaptive-radix-tree) [![GoDoc](https://godoc.org/github.com/plar/go-adaptive-radix-tree?status.svg)](http://godoc.org/github.com/plar/go-adaptive-radix-tree)

> v1.x.x is a deprecated version. All development will be done in [v2](https://github.com/plar/go-adaptive-radix-tree).

This library provides a Go implementation of the Adaptive Radix Tree (ART).

Features:
* Lookup performance surpasses highly tuned alternatives
* Support for highly efficient insertions and deletions
* Space efficient
* Performance is comparable to hash tables
* Maintains the data in sorted order, which enables additional operations like range scan and prefix lookup
* `O(k)` search/insert/delete operations, where `k` is the length of the key
* Minimum / Maximum value lookups
* Ordered iteration
* Prefix-based iteration
* Support for keys with null bytes, any byte array could be a key

# Usage

```go
package main

import (
    "fmt"
    "github.com/plar/go-adaptive-radix-tree"
)

func main() {
	// Initialize a new Adaptive Radix Tree
	tree := art.New()

	// Insert key-value pairs into the tree
	tree.Insert(art.Key("apple"), "A sweet red fruit")
	tree.Insert(art.Key("banana"), "A long yellow fruit")
	tree.Insert(art.Key("cherry"), "A small red fruit")
	tree.Insert(art.Key("date"), "A sweet brown fruit")

	// Search for a value by key
	if value, found := tree.Search(art.Key("banana")); found {
		fmt.Println("Found:", value)
	} else {
		fmt.Println("Key not found")
	}

	// Iterate over the tree in ascending order
	fmt.Println("\nAscending order iteration:")
	tree.ForEach(func(node art.Node) bool {
		fmt.Printf("Key: %s, Value: %s\n", node.Key(), node.Value())
		return true
	})

	// Iterate over the tree in descending order using reverse traversal
	fmt.Println("\nDescending order iteration:")
	tree.ForEach(func(node art.Node) bool {
		fmt.Printf("Key: %s, Value: %s\n", node.Key(), node.Value()))
		return true
	}, art.TraverseReverse)

	// Iterate over keys with a specific prefix
	fmt.Println("\nIteration with prefix 'c':")
	tree.ForEachPrefix(art.Key("c"), func(node art.Node) bool {
		fmt.Printf("Key: %s, Value: %s\n", node.Key(), node.Value())
		return true
	})
}

// Expected Output:
// Found: A long yellow fruit
//
// Ascending order iteration:
// Key: apple, Value: A sweet red fruit
// Key: banana, Value: A long yellow fruit
// Key: cherry, Value: A small red fruit
// Key: date, Value: A sweet brown fruit
//
// Descending order iteration:
// Key: date, Value: A sweet brown fruit
// Key: cherry, Value: A small red fruit
// Key: banana, Value: A long yellow fruit
// Key: apple, Value: A sweet red fruit
//
// Iteration with prefix 'c':
// Key: cherry, Value: A small red fruit
```

# Documentation

Check out the documentation on [godoc.org](http://godoc.org/github.com/plar/go-adaptive-radix-tree)

# Performance

[plar/go-adaptive-radix-tree](https://github.com/plar/go-adaptive-radix-tree) outperforms [kellydunn/go-art](https://github.com/kellydunn/go-art) by avoiding memory allocations during search operations.
It also provides prefix based iteration over the tree.

Benchmarks were performed on datasets extracted from different projects:
- The "Words" dataset contains a list of 235,886 english words. [2]
- The "UUIDs" dataset contains 100,000 uuids.                   [2]
- The "HSK Words" dataset contains 4,995 words.                 [4]

|**go-adaptive-radix-tree**| #  | Average time      |Bytes per operation|Allocs per operation |
|:-------------------------|---:|------------------:|------------------:|--------------------:|
|       Tree Insert Words  |  9 | 117,888,698 ns/op |  37,942,744  B/op | 1,214,541 allocs/op |
|       Tree Search Words  | 26 |  44,555,608 ns/op |            0 B/op |         0 allocs/op |
|       Tree Insert UUIDs  | 18 |  59,360,135 ns/op |   18,375,723 B/op |   485,057 allocs/op |
|       Tree Search UUIDs  | 54 |  21,265,931 ns/op |            0 B/op |         0 allocs/op |
|**go-art**                |    |                   |                   |                     |
|       Tree Insert Words  |  5 | 272,047,975 ns/op |   81,628,987 B/op | 2,547,316 allocs/op |
|       Tree Search Words  | 10 | 129,011,177 ns/op |   13,272,278 B/op | 1,659,033 allocs/op |
|       Tree Insert UUIDs  | 10 | 140,309,246 ns/op |   33,678,160 B/op |   874,561 allocs/op |
|       Tree Search UUIDs  | 20 |  82,120,943 ns/op |    3,883,131 B/op |   485,391 allocs/op |

To see more benchmarks just run

```
$ ./make qa/benchmarks
```

# References

[1] [The Adaptive Radix Tree: ARTful Indexing for Main-Memory Databases (Specification)](http://www-db.in.tum.de/~leis/papers/ART.pdf)

[2] [C99 implementation of the Adaptive Radix Tree](https://github.com/armon/libart)

[3] [Another Adaptive Radix Tree implementation in Go](https://github.com/kellydunn/go-art)

[4] [HSK Words](http://hskhsk.pythonanywhere.com/hskwords). HSK(Hanyu Shuiping Kaoshi) - Standardized test of Standard Mandarin Chinese proficiency.
