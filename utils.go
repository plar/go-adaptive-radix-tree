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
