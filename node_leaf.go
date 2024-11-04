package art

import "bytes"

// Leaf node with a variable key length.
type leaf struct {
	key   Key
	value interface{}
}

// match returns true if the leaf node's key matches the given key.
func (l *leaf) match(key Key) bool {
	return len(l.key) == len(key) && bytes.Equal(l.key, key)
}

// prefixMatch returns true if the leaf node's key has the given key as a prefix.
func (l *leaf) prefixMatch(key Key) bool {
	if key == nil || len(l.key) < len(key) {
		return false
	}

	return bytes.Equal(l.key[:len(key)], key)
}
