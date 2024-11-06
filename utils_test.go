package art

import (
	"bufio"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testDataset defines a dataset for testing the tree.
type testDataset struct {
	name         string
	insertItems  interface{}
	deleteItems  interface{}
	expectedSize int
	expectedRoot interface{}
	deleteStatus bool
}

type testDatasetBuilder func(data *testDataset, tree *tree)

func (ds *testDataset) build(_ *testing.T, tree *tree) {
	switch insertData := ds.insertItems.(type) {
	case []string:
		for _, term := range insertData {
			tree.Insert(Key(term), term)
		}
	case testDatasetBuilder:
		insertData(ds, tree)
	}
}

func (ds *testDataset) process(t *testing.T, tree *tree) {
	t.Helper()

	switch deleteData := ds.deleteItems.(type) {
	case []string:
		ds.processAsStrings(t, tree, deleteData)
	case []byte:
		ds.processAsBytes(t, tree, deleteData)
	case testDatasetBuilder:
		deleteData(ds, tree)
	}
}

func (ds *testDataset) processAsStrings(t *testing.T, tree *tree, stringData []string) {
	t.Helper()

	for _, strVal := range stringData {
		ds.processSingleItem(t, tree, Key(strVal), strVal)
	}
}

func (ds *testDataset) processAsBytes(t *testing.T, tree *tree, bytesData []byte) {
	t.Helper()

	for _, byteVal := range bytesData {
		ds.processSingleItem(t, tree, Key{byteVal}, []byte{byteVal})
	}
}

func (ds *testDataset) processSingleItem(t *testing.T, tree *tree, key Key, expectedVal interface{}) {
	t.Helper()

	val, deleted := tree.Delete(key)
	assert.Equal(t, ds.deleteStatus, deleted, ds.name)

	if deleted {
		assert.Equal(t, expectedVal, val, ds.name)
	}

	_, found := tree.Search(key)
	assert.False(t, found, ds.name)
}

func (ds *testDataset) assert(t *testing.T, tree *tree) {
	t.Helper()
	assert.Equal(t, ds.expectedSize, tree.size, ds.name)

	switch root := ds.expectedRoot.(type) {
	case Kind:
		assert.Equal(t, root, tree.root.kind, ds.name)
	case *nodeRef:
		assert.Equal(t, root, tree.root, ds.name)
	case nil:
		assert.Nil(t, tree.root, ds.name)
	}
}

// treeStats defines the statistics of the tree.
type treeStats struct {
	leafCount    int
	node4Count   int
	node16Count  int
	node48Count  int
	node256Count int
}

// processStats processes the node statistics.
func (stats *treeStats) processStats(node Node) bool {
	switch node.Kind() {
	case Node4:
		stats.node4Count++
	case Node16:
		stats.node16Count++
	case Node48:
		stats.node48Count++
	case Node256:
		stats.node256Count++
	case Leaf:
		stats.leafCount++
	}

	return true
}

// collectStats collects the statistics of the tree.
func collectStats(it Iterator) treeStats {
	stats := treeStats{}

	for it.HasNext() {
		node, _ := it.Next()
		stats.processStats(node)
	}

	return stats
}

// loadTestFile loads the test file from the given path.
func loadTestFile(path string) [][]byte {
	var words [][]byte

	file, err := os.Open(path)
	if err != nil {
		panic("Couldn't open " + path)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		if line, err := reader.ReadBytes(byte('\n')); err != nil {
			break
		} else if len(line) > 0 {
			words = append(words, line[:len(line)-1])
		}
	}

	return words
}

// treeWithData creates a tree with the data from the given file.
func treeWithData(filePath string) (*tree, [][]byte) {
	tree := newTree()

	data := loadTestFile(filePath)
	for _, item := range data {
		tree.Insert(item, item)
	}

	return tree, data
}
