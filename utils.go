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
	dst.zeroChild = src.zeroChild

	maxCopyLen := min(int(src.prefixLen), MaxPrefixLen)
	for i := 0; i < maxCopyLen; i++ {
		dst.prefix[i] = src.prefix[i]
	}
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
func nodeMinimum(zeroChild *nodeRef, children []*nodeRef) *leaf {
	if zeroChild != nil {
		return zeroChild.minimum()
	}

	for _, child := range children {
		if child != nil {
			return child.minimum()
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
