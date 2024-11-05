package art

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// copy the node from src to dst
func copyNode(dst *node, src *node) {
	if dst == nil || src == nil {
		return
	}
	dst.prefixLen = src.prefixLen
	dst.prefix = src.prefix
}

// find the child node index by key
func findIndex(keys []byte, ch byte) int {
	for i, key := range keys {
		if key == ch {
			return i
		}
	}
	return indexNotFound
}

// findLongestCommonPrefix returns the longest common prefix of key1 and key2
func findLongestCommonPrefix(key1 Key, key2 Key, keyOffset int) int {
	limit := min(len(key1), len(key2))

	idx := keyOffset
	for ; idx < limit; idx++ {
		if key1[idx] != key2[idx] {
			break
		}
	}

	return idx - keyOffset
}

// find the minimum leaf node
func nodeMinimum(children []*nodeRef) *leaf {
	numChildren := len(children)
	if numChildren == 0 {
		return nil
	}

	// zero byte key
	if children[numChildren-1] != nil {
		return children[numChildren-1].minimum()
	}

	for i := 0; i < numChildren-1; i++ {
		if children[i] != nil {
			return children[i].minimum()
		}
	}

	return nil
}

// find the maximum leaf node
func nodeMaximum(children []*nodeRef) *leaf {
	for i := len(children) - 1; i >= 0; i-- {
		if children[i] != nil {
			return children[i].maximum()
		}
	}

	return nil
}
