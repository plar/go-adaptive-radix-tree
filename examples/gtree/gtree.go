package main

import (
	"errors"
	"fmt"

	art "github.com/plar/go-adaptive-radix-tree/v2"
)

// GTree is a generic tree that supports any type for keys and values.
type GTree[K comparable, V any] struct {
	tree art.Tree
}

// NewGTree creates a new generic adaptive radix tree.
func NewGTree[K comparable, V any]() *GTree[K, V] {
	return &GTree[K, V]{
		tree: art.New(),
	}
}

// Insert a new key-value pair into the tree.
func (gt *GTree[K, V]) Insert(key K, value V) (oldValue V, updated bool, err error) {
	// Convert key to []byte
	keyBytes, err := convertKeyToBytes(key)
	if err != nil {
		return
	}

	oldVal, updated := gt.tree.Insert(art.Key(keyBytes), value)

	if oldVal != nil {
		oldValue = oldVal.(V)
	}
	return
}

// Delete removes a key from the tree.
func (gt *GTree[K, V]) Delete(key K) (value V, deleted bool) {
	keyBytes, err := convertKeyToBytes(key)
	if err != nil {
		return
	}

	val, deleted := gt.tree.Delete(art.Key(keyBytes))
	if val != nil {
		value = val.(V)
	}
	return
}

// Search for a key in the tree.
func (gt *GTree[K, V]) Search(key K) (value V, found bool) {
	keyBytes, err := convertKeyToBytes(key)
	if err != nil {
		return
	}

	val, found := gt.tree.Search(art.Key(keyBytes))
	if val != nil {
		value = val.(V)
	}
	return
}

// Size returns the number of elements in the tree.
func (gt *GTree[K, V]) Size() int {
	return gt.tree.Size()
}

// ForEach performs the given callback on each node.
func (gt *GTree[K, V]) ForEach(callback func(node art.Node) bool, options ...int) {
	gt.tree.ForEach(callback, options...)
}

// Helper function to convert an integer key to a byte slice.
func convertKeyToBytes[K comparable](key K) ([]byte, error) {
	switch v := any(key).(type) {
	case int:
		return []byte(fmt.Sprintf("%d", v)), nil // Simple conversion
	default:
		return nil, errors.New("unsupported key type")
	}
}
