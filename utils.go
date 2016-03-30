package art

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func memcpy(dest []byte, src []byte, numBytes int) {
	for i := 0; i < numBytes && i < len(src) && i < len(dest); i++ {
		dest[i] = src[i]
	}
}

func keyChar(key Key, pos int) byte {
	if pos >= 0 && pos < len(key) {
		return key[pos]
	} else {
		return 0
	}
}

func prefixChar(prefix [MAX_PREFIX_LENGTH]byte, pos int) byte {
	if pos >= 0 && pos < len(prefix) {
		return prefix[pos]
	} else {
		return 0
	}
}

func copyMeta(dst, src *artNode) *artNode {
	if dst == nil || src == nil {
		return dst
	}

	d := dst.BaseNode()
	s := src.BaseNode()

	d.numChildren = s.numChildren
	d.prefixLen = s.prefixLen

	for i, limit := 0, min(MAX_PREFIX_LENGTH, s.prefixLen); i < limit; i++ {
		d.prefix[i] = s.prefix[i]
	}

	return dst
}
