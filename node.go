package art

// prefix used in the node to store the key prefix.
// it is used to improve leaf key comparison performance.
type prefix [MaxPrefixLen]byte

// node is the base struct for all node types.
type node struct {
	prefix      prefix // prefix of the node
	prefixLen   uint16 // length of the prefix
	childrenLen uint16 // number of children in the node4, node16, node48, node256
}

// String returns string representation of the Kind value.
func (k Kind) String() string {
	return []string{"Leaf", "Node4", "Node16", "Node48", "Node256"}[k]
}

// charAt returns the character at the given index.
// If the index is out of bounds, it returns 0 and false.
func (k Key) charAt(idx int) (byte, bool) {
	if k.isValid(idx) {
		return k[idx], true
	}
	return 0, false
}

// isValid checks if the given index is within the bounds of the key.
func (k Key) isValid(idx int) bool {
	return idx >= 0 && idx < len(k)
}

// Node helpers
func replaceRef(oldNode **nodeRef, newNode *nodeRef) {
	*oldNode = newNode
}

func replaceNode(oldNode *nodeRef, newNode *nodeRef) {
	*oldNode = *newNode
}
