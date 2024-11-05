package art

// insertRecursively inserts a new key-value pair into the tree.
// nrp means Node Reference Pointer
func (tr *tree) insertRecursively(nrp **nodeRef, key Key, value Value, keyOffset int) (Value, treeOpResult) {
	nr := *nrp
	if nr == nil {
		return tr.insertNewLeaf(nrp, key, value)
	}

	if nr.isLeaf() {
		return tr.handleLeafInsertion(nrp, key, value, keyOffset)
	}

	return tr.handleNodeInsertion(nrp, key, value, keyOffset)
}

func (tr *tree) insertNewLeaf(nrp **nodeRef, key Key, value Value) (Value, treeOpResult) {
	replaceRef(nrp, factory.newLeaf(key, value))
	return nil, treeOpInserted
}

func (tr *tree) handleLeafInsertion(nrp **nodeRef, key Key, value Value, keyOffset int) (Value, treeOpResult) {
	nr := *nrp

	if leaf := nr.leaf(); leaf.match(key) {
		oldValue := leaf.value
		leaf.value = value
		return oldValue, treeOpUpdated
	}

	// Insert a new leaf by splitting
	// the old leaf to a node4 and adding the new leaf
	return tr.splitLeaf(nrp, key, value, keyOffset)
}

func (tr *tree) splitLeaf(nrpCurLeaf **nodeRef, key Key, value Value, keyOffset int) (Value, treeOpResult) {
	nrCurLeaf := *nrpCurLeaf
	curLeaf := nrCurLeaf.leaf()

	keysLCP := findLongestCommonPrefix(curLeaf.key, key, keyOffset)

	// Create a new node4 with the longest common prefix
	// between the old leaf and the new leaf key.
	nr4 := factory.newNode4()
	nr4.setPrefix(key[keyOffset:], keysLCP)
	keyOffset += keysLCP

	// branch by the first differing character
	// add the old leaf and the new leaf as children
	// to a newly created node4.
	nr4.addChild(curLeaf.key.charAt(keyOffset), nrCurLeaf)           // old leaf
	nr4.addChild(key.charAt(keyOffset), factory.newLeaf(key, value)) // new leaf

	// replace the old leaf with the new node4
	replaceRef(nrpCurLeaf, nr4)
	return nil, treeOpInserted
}

func (tr *tree) handleNodeInsertion(nrp **nodeRef, key Key, value Value, keyOffset int) (Value, treeOpResult) {
	nr := *nrp
	n := nr.node()
	if n.prefixLen > 0 {
		prefixMismatchIdx := nr.matchDeep(key, keyOffset)
		if prefixMismatchIdx < int(n.prefixLen) {
			return tr.splitNode(nrp, key, value, keyOffset, prefixMismatchIdx)
		}
		keyOffset += int(n.prefixLen)
	}
	return tr.continueInsertion(nrp, key, value, keyOffset)
}

func (tr *tree) splitNode(nrp **nodeRef, key Key, value Value, keyOffset int, mismatchIdx int) (Value, treeOpResult) {
	nr := *nrp
	n := nr.node()

	nr4 := factory.newNode4()
	nr4.setPrefix(n.prefix[:], mismatchIdx)

	tr.reassignPrefix(nr4, nr, key, value, keyOffset, mismatchIdx)

	replaceRef(nrp, nr4)
	return nil, treeOpInserted
}

func (tr *tree) reassignPrefix(newNRP *nodeRef, curNRP *nodeRef, key Key, value Value, keyOffset int, mismatchIdx int) {
	curNode := curNRP.node()
	curNode.prefixLen -= uint16(mismatchIdx + 1)

	idx := keyOffset + mismatchIdx

	// Adjust prefix and add children
	leaf := curNRP.minimum()
	newNRP.addChild(leaf.key.charAt(idx), curNRP)

	for i := 0; i < min(int(curNode.prefixLen), MaxPrefixLen); i++ {
		curNode.prefix[i] = leaf.key[keyOffset+mismatchIdx+i+1]
	}

	// Insert the new leaf
	newNRP.addChild(key.charAt(idx), factory.newLeaf(key, value))
}

func (tr *tree) continueInsertion(nrp **nodeRef, key Key, value Value, keyOffset int) (Value, treeOpResult) {
	nr := *nrp
	nextNRP := nr.findChildByKey(key, keyOffset)
	if *nextNRP != nil {
		// Found a partial match, continue inserting
		return tr.insertRecursively(nextNRP, key, value, keyOffset+1)
	}

	// No child found, create a new leaf node
	nr.addChild(key.charAt(int(keyOffset)), factory.newLeaf(key, value))
	return nil, treeOpInserted
}
