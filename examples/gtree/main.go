package main

import (
	"fmt"

	art "github.com/plar/go-adaptive-radix-tree/v2"
)

func main() {
	// Create a new generic tree with int as Key and string as Value.
	tree := NewGTree[int, string]()

	// Insert some values into the tree.
	tree.Insert(1, "one")
	tree.Insert(2, "two")
	tree.Insert(3, "three")

	// Search for a value.
	if value, found := tree.Search(2); found {
		fmt.Printf("Found: %s\n", value) // Output: Found: two
	} else {
		fmt.Println("Not found")
	}

	// Delete a value.
	if value, deleted := tree.Delete(3); deleted {
		fmt.Printf("Deleted: %s\n", value) // Output: Deleted: three
	} else {
		fmt.Println("Not found for deletion")
	}

	// Check the size of the tree.
	fmt.Printf("Tree Size: %d\n", tree.Size()) // Output: Tree Size: 2

	// Traverse the tree using ForEach.
	tree.ForEach(func(node art.Node) bool {
		fmt.Printf("Node Key: %s, Node Value: %s\n", string(node.Key()), node.Value().(string))
		return true // Continue iteration
	}, art.TraverseLeaf)
}
