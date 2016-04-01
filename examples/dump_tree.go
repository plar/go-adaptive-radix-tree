package main

import (
	"fmt"

	"github.com/plar/go-adaptive-radix-tree"
)

func main() {
	tree := art.New()
	terms := []string{"A", "a", "aa"}
	for _, term := range terms {
		tree.Insert(art.Key(term), term)
	}
	fmt.Println(tree)
}
