package art

// prefix used in the node to store the key prefix.
// it is used to improve leaf key comparison performance.
type prefix [maxPrefixLen]byte

// node is the base struct for all node types.
// it contains the common fields for all nodeX types.
type node struct {
	prefix      prefix // prefix of the node
	prefixLen   uint16 // length of the prefix
	childrenLen uint16 // number of children in the node4, node16, node48, node256
}

// replaceRef is used to replace node in-place by updating the reference.
func replaceRef(oldNode **nodeRef, newNode *nodeRef) {
	*oldNode = newNode
}

// replaceNode is used to replace node in-place by updating the node.
func replaceNode(oldNode *nodeRef, newNode *nodeRef) {
	*oldNode = *newNode
}
