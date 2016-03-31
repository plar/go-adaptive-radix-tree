package art

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type treeStats struct {
	leafCount    int
	node4Count   int
	node16Count  int
	node48Count  int
	node256Count int
}

var words [][]byte

func loadTestFile(path string) [][]byte {
	file, err := os.Open(path)
	if err != nil {
		panic("Couldn't open " + path)
	}
	defer file.Close()

	var words [][]byte
	reader := bufio.NewReader(file)
	for {
		if line, err := reader.ReadBytes(byte('\n')); err != nil {
			break
		} else {
			if len(line) > 0 {
				words = append(words, line[:len(line)-1])
			}
		}
	}
	return words
}

func TestTreeLongestCommonPrefix(t *testing.T) {
	tree := &tree{}

	l1 := factory.newLeaf([]byte("abcdefg12345678"), "abcdefg12345678").Leaf()
	l2 := factory.newLeaf([]byte("abcdefg!@#$%^&*"), "abcdefg!@#$%^&*").Leaf()
	assert.Equal(t, 7, tree.longestCommonPrefix(l1, l2, 0))
	assert.Equal(t, 3, tree.longestCommonPrefix(l1, l2, 4))

	l1 = factory.newLeaf([]byte("abcdefg12345678"), "abcdefg12345678").Leaf()
	l2 = factory.newLeaf([]byte("defg!@#$%^&*"), "defg!@#$%^&*").Leaf()
	assert.Equal(t, 0, tree.longestCommonPrefix(l1, l2, 0))
}

func TestTreeInit(t *testing.T) {
	tree := New()
	assert.NotNil(t, tree)
}

func TestPoolFactory(t *testing.T) {
	factory := newPoolObjFactory()
	n4 := factory.newNode48()
	assert.NotNil(t, n4)
	factory.releaseNode(n4)

	n4v2 := factory.newNode48()
	assert.True(t, n4 == n4v2)

	n4v3 := factory.newNode48()
	assert.NotEqual(t, n4v2, n4v3)
}

func TestObjFactory(t *testing.T) {
	factory := newObjFactory()
	n4 := factory.newNode48()
	assert.NotNil(t, n4)
	n4v2 := factory.newNode48()
	assert.True(t, n4 != n4v2)
}

func TestTreeUpdate(t *testing.T) {
	tree := newTree()

	key := []byte("key")

	ov, updated := tree.Insert(key, "value")
	assert.Nil(t, ov)
	assert.False(t, updated)
	assert.Equal(t, 1, tree.size)
	assert.Equal(t, Kind(NODE_LEAF), tree.root.kind)

	v, found := tree.Search(key)
	assert.True(t, found)
	assert.Equal(t, "value", v.(string))

	ov, updated = tree.Insert(key, "otherValue")
	assert.Equal(t, "value", ov.(string))
	assert.True(t, updated)
	assert.Equal(t, 1, tree.size)

	v, found = tree.Search(key)
	assert.True(t, found)
	assert.Equal(t, "otherValue", v.(string))
}

func TestTreeInsertSimilarPrefix(t *testing.T) {
	tree := newTree()

	tree.Insert([]byte{1}, 1)
	tree.Insert([]byte{1, 1}, 11)

	v, found := tree.Search([]byte{1, 1})
	assert.True(t, found)
	assert.Equal(t, 11, v.(int))

}

// An Art Node with a similar prefix should be split into new nodes accordingly
// And should be searchable as intended.
func TestTreeInsert3AndSearchWords(t *testing.T) {
	tree := newTree()

	searchTerms := []string{"A", "a", "aa"}

	for _, term := range searchTerms {
		tree.Insert([]byte(term), term)
	}

	for _, term := range searchTerms {
		v, found := tree.Search([]byte(term))
		assert.True(t, found)
		assert.Equal(t, term, v.(string))
	}
}

func TestTreeInsertAndGrowToBiggerNode(t *testing.T) {
	var testData = []struct {
		totalNodes byte
		expected   Kind
	}{
		{5, NODE_16},
		{17, NODE_48},
		{49, NODE_256},
	}

	for _, data := range testData {
		tree := newTree()
		for i := byte(0); i < data.totalNodes; i++ {
			tree.Insert([]byte{i}, i)
		}
		assert.Equal(t, int(data.totalNodes), tree.size)
		assert.Equal(t, data.expected, tree.root.kind)
	}
}

func TestTreeInsertWordsAndMinMax(t *testing.T) {
	words := loadTestFile("test/assets/words.txt")
	tree := newTree()
	for _, w := range words {
		tree.Insert(w, w)
	}

	for _, w := range words {
		v, found := tree.Search(w)
		assert.Equal(t, w, v)
		assert.True(t, found)
	}

	minimum := tree.root.Minimum()
	assert.Equal(t, []byte("A"), minimum.value.([]byte))
	maximum := tree.root.Maximum()
	assert.Equal(t, []byte("zythum"), maximum.value.([]byte))
}

func TestTreeInsertUUIDsAndMinMax(t *testing.T) {
	words := loadTestFile("test/assets/uuid.txt")
	tree := newTree()
	for _, w := range words {
		tree.Insert(w, w)
	}

	for _, w := range words {
		v, found := tree.Search(w)
		assert.Equal(t, w, v)
		assert.True(t, found)
	}

	minimum := tree.root.Minimum()
	assert.Equal(t, []byte("00026bda-e0ea-4cda-8245-522764e9f325"), minimum.value.([]byte))
	maximum := tree.root.Maximum()
	assert.Equal(t, []byte("ffffcb46-a92e-4822-82af-a7190f9c1ec5"), maximum.value.([]byte))
}

func TestTreeInsertAndDeleteOperations(t *testing.T) {
	type testData struct {
		message string
		insert  interface{}
		delete  interface{}
		size    int
		root    interface{}
	}
	type treeTestCustom func(data *testData, tree *tree)

	var data = []testData{
		{
			"Insert 1 Delete 1",
			[]string{"test"},
			[]string{"test"},
			0,
			nil},
		{
			"Insert 2 Delete 1",
			[]string{"test1", "test2"},
			[]string{"test2"},
			1,
			NODE_LEAF},
		{
			"Insert 2 Delete 2",
			[]string{"test1", "test2"},
			[]string{"test1", "test2"},
			0,
			nil},
		{
			"Insert 5 Delete 1",
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1"},
			4,
			NODE_4},
		{
			"Insert 5 Delete 5",
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1", "2", "3", "4", "5"},
			0,
			nil},
		{
			"Insert 17 Delete 1",
			[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17"},
			[]string{"2"},
			16,
			NODE_16},
		{
			"Insert 17 Delete 17",
			[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17"},
			[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17"},
			0,
			nil},
		{
			"Insert 49 Delete 1",
			treeTestCustom(func(data *testData, tree *tree) {
				for i := byte(0); i < 49; i++ {
					tree.Insert([]byte{i}, []byte{i})
				}
			}),
			[]byte{byte(2)},
			48,
			NODE_48},
		{
			"Insert 49 Delete 49",
			treeTestCustom(func(data *testData, tree *tree) {
				for i := byte(0); i < 49; i++ {
					tree.Insert([]byte{i}, []byte{i})
				}
			}),
			treeTestCustom(func(data *testData, tree *tree) {
				for i := byte(0); i < 49; i++ {
					term := []byte{i}
					val, deleted := tree.Delete(term)
					assert.True(t, deleted, data.message)
					assert.Equal(t, term, val)
				}
			}),
			0,
			nil},
		{
			"Insert 256 Delete 1",
			treeTestCustom(func(data *testData, tree *tree) {
				for i := 0; i < 256; i++ {
					term := bytes.NewBuffer([]byte{})
					term.WriteByte(byte(i))
					tree.Insert(term.Bytes(), term.Bytes())
				}
			}),
			[]byte{2},
			255,
			NODE_256},
		{
			"Insert 256 Delete 256",
			treeTestCustom(func(data *testData, tree *tree) {
				for i := 0; i < 256; i++ {
					term := fmt.Sprintf("%d", i)
					tree.Insert([]byte(term), term)
				}
			}),
			treeTestCustom(func(data *testData, tree *tree) {
				for i := 0; i < 256; i++ {
					term := fmt.Sprintf("%d", i)
					val, deleted := tree.Delete([]byte(term))
					assert.True(t, deleted, data.message)
					assert.Equal(t, term, val.(string))
				}
			}),
			0,
			nil},
	}

	for _, ds := range data {
		tree := newTree()

		// insert test data...
		if strs, ok := ds.insert.([]string); ok {
			for _, term := range strs {
				tree.Insert([]byte(term), term)
			}
		} else if builder, ok := ds.insert.(treeTestCustom); ok {
			builder(&ds, tree)
		}

		// delete test data and check...
		if strs, ok := ds.delete.([]string); ok {
			for _, term := range strs {
				val, deleted := tree.Delete([]byte(term))
				assert.True(t, deleted, ds.message)

				if s, ok := val.(string); ok {
					assert.Equal(t, term, s, ds.message)
				}

				_, found := tree.Search([]byte(term))
				assert.False(t, found, ds.message)
			}
		} else if bts, ok := ds.delete.([]byte); ok {
			for _, term := range bts {

				val, deleted := tree.Delete([]byte{term})
				assert.True(t, deleted, ds.message)

				assert.Equal(t, []byte{term}, val, ds.message)

				_, found := tree.Search([]byte{term})
				assert.False(t, found, ds.message)
			}
		} else if builder, ok := ds.delete.(treeTestCustom); ok {
			builder(&ds, tree)
		}

		// general assertions...
		assert.Equal(t, ds.size, tree.size, ds.message)

		if ds.root == nil {
			assert.Nil(t, tree.root, ds.message)
		} else if k, ok := ds.root.(Kind); ok {
			assert.Equal(t, k, tree.root.kind, ds.message)
		} else if an, ok := ds.root.(*artNode); ok {
			assert.Equal(t, an, tree.root, ds.message)
		}

		//fmt.Println("===================================\n", DumpNode(tree.root))
	}
}

// Inserting a single value into the tree and removing it should result in a nil tree root.
func TestInsertAndDeleteOne(t *testing.T) {
	tree := newTree()
	tree.Insert([]byte("test"), []byte("data"))
	v, deleted := tree.Delete([]byte("test"))
	assert.True(t, deleted)
	assert.Equal(t, []byte("data"), v.([]byte))
	assert.Equal(t, 0, tree.size)
	assert.Nil(t, tree.root)
}

func TestInsertTwoAndDeleteOne(t *testing.T) {
	tree := newTree()
	tree.Insert([]byte("2"), []byte{2})
	tree.Insert([]byte("1"), []byte{1})

	_, found := tree.Search([]byte("2"))
	assert.True(t, found)
	_, found = tree.Search([]byte("1"))
	assert.True(t, found)

	v, deleted := tree.Delete([]byte("2"))
	assert.True(t, deleted)
	if deleted {
		assert.Equal(t, []byte{2}, v.([]byte))
	}

	assert.Equal(t, 1, tree.size)
	assert.Equal(t, NODE_LEAF, tree.root.kind)
}

func TestInsertTwoAndDeleteTwo(t *testing.T) {
	tree := newTree()
	tree.Insert([]byte("2"), []byte{2})
	tree.Insert([]byte("1"), []byte{1})

	_, found := tree.Search([]byte("2"))
	assert.True(t, found)
	_, found = tree.Search([]byte("1"))
	assert.True(t, found)

	v, deleted := tree.Delete([]byte("2"))
	assert.True(t, deleted)
	assert.Equal(t, []byte{2}, v.([]byte))

	v, deleted = tree.Delete([]byte("1"))
	assert.True(t, deleted)
	assert.Equal(t, []byte{1}, v.([]byte))

	assert.Equal(t, 0, tree.size)
	assert.Nil(t, tree.root)
}

func TestTreeTraversalPreordered(t *testing.T) {
	tree := newTree()

	tree.Insert([]byte("1"), 1)
	tree.Insert([]byte("2"), 2)

	traversal := []Node{}
	tree.ForEach(func(node Node) bool {
		traversal = append(traversal, node)
		return true
	})

	assert.Equal(t, 2, tree.size)
	assert.Equal(t, traversal[0].(*artNode), tree.root)
	assert.Nil(t, traversal[0].Key())
	assert.Nil(t, traversal[0].Value())
	assert.NotEqual(t, NODE_LEAF, traversal[0].Kind())

	assert.Equal(t, traversal[1].Key(), Key("1"))
	assert.Equal(t, traversal[1].Value().(int), 1)
	assert.Equal(t, NODE_LEAF, traversal[1].Kind())

	assert.Equal(t, traversal[2].Key(), Key("2"))
	assert.Equal(t, traversal[2].Value().(int), 2)
	assert.Equal(t, NODE_LEAF, traversal[2].Kind())
}

func TestTreeTraversalNode48(t *testing.T) {
	tree := newTree()

	for i := 48; i > 0; i-- {
		tree.Insert([]byte{byte(i)}, i)
	}

	traversal := []Node{}
	tree.ForEach(func(node Node) bool {
		traversal = append(traversal, node)
		return true
	})

	// Order should be Node48, then the rest of the keys in sorted order
	assert.Equal(t, 48, tree.size)
	assert.Equal(t, traversal[0].(*artNode), tree.root)
	assert.Equal(t, NODE_48, traversal[0].Kind())

	for i := 1; i < 48; i++ {
		assert.Equal(t, traversal[i].Key(), Key([]byte{byte(i)}))
		assert.Equal(t, NODE_LEAF, traversal[i].Kind())
	}
}

func TestTreeTraversalWordsStats(t *testing.T) {
	words := loadTestFile("test/assets/words.txt")
	tree := New()
	for _, w := range words {
		tree.Insert(w, w)
	}

	stats := treeStats{}
	tree.ForEach(func(node Node) bool {
		switch node.Kind() {
		case NODE_4:
			stats.node4Count++
		case NODE_16:
			stats.node16Count++
		case NODE_48:
			stats.node48Count++
		case NODE_256:
			stats.node256Count++
		case NODE_LEAF:
			stats.leafCount++
		}
		return true
	})

	assert.Equal(t, treeStats{235886, 111616, 12181, 458, 1}, stats)
}

func TestTreeIterator(t *testing.T) {
	tree := newTree()
	tree.Insert([]byte("2"), []byte{2})
	tree.Insert([]byte("1"), []byte{1})

	it := tree.Iterator()
	assert.NotNil(t, it)
	assert.True(t, it.HasNext())
	n4, err := it.Next()
	assert.NoError(t, err)
	assert.Equal(t, NODE_4, n4.Kind())

	assert.True(t, it.HasNext())
	v1, err := it.Next()
	assert.NoError(t, err)
	assert.Equal(t, v1.Value().([]byte), []byte{1})

	assert.True(t, it.HasNext())
	v2, err := it.Next()
	assert.NoError(t, err)
	assert.Equal(t, v2.Value().([]byte), []byte{2})

	assert.False(t, it.HasNext())
	bad, err := it.Next()
	assert.Nil(t, bad)
	assert.Equal(t, "There are no more nodes in the tree", err.Error())

}

func TestTreeIteratorConcurrentModification(t *testing.T) {
	tree := newTree()
	tree.Insert([]byte("2"), []byte{2})
	tree.Insert([]byte("1"), []byte{1})

	it1 := tree.Iterator()
	assert.NotNil(t, it1)
	assert.True(t, it1.HasNext())

	// simulate concurrent modification
	tree.Insert([]byte("3"), []byte{3})
	bad, err := it1.Next()
	assert.Nil(t, bad)
	assert.Equal(t, "Concurrent modification has been detected", err.Error())

	it2 := tree.Iterator()
	assert.NotNil(t, it2)
	assert.True(t, it2.HasNext())

	tree.Delete([]byte("3"))
	bad, err = it2.Next()
	assert.Nil(t, bad)
	assert.Equal(t, "Concurrent modification has been detected", err.Error())
}

func TestTreeIterateWordsStats(t *testing.T) {
	words := loadTestFile("test/assets/words.txt")
	tree := New()
	for _, w := range words {
		tree.Insert(w, w)
	}

	stats := treeStats{}

	for it := tree.Iterator(); it.HasNext(); {
		node, _ := it.Next()
		switch node.Kind() {
		case NODE_4:
			stats.node4Count++
		case NODE_16:
			stats.node16Count++
		case NODE_48:
			stats.node48Count++
		case NODE_256:
			stats.node256Count++
		case NODE_LEAF:
			stats.leafCount++
		}
	}

	assert.Equal(t, treeStats{235886, 111616, 12181, 458, 1}, stats)
}

func TestTreeInsertAndDeleteAllWords(t *testing.T) {
	words := loadTestFile("test/assets/words.txt")
	tree := newTree()
	for _, w := range words {
		tree.Insert(w, w)
	}

	for _, w := range words {
		v, deleted := tree.Delete(w)
		assert.True(t, deleted)
		assert.Equal(t, w, v.([]byte))
	}

	assert.Equal(t, 0, tree.size)
	assert.Nil(t, tree.root)
}

func TestTreeInsertAndDeleteAllUUIDs(t *testing.T) {
	uuids := loadTestFile("test/assets/uuid.txt")
	tree := newTree()
	for _, w := range uuids {
		tree.Insert(w, w)
	}

	for _, w := range uuids {
		v, deleted := tree.Delete(w)
		assert.True(t, deleted)
		assert.Equal(t, w, v.([]byte))
	}

	assert.Equal(t, 0, tree.size)
	assert.Nil(t, tree.root)
}

func BenchmarkTreeInsertWords(b *testing.B) {
	b.StopTimer()
	words := loadTestFile("test/assets/words.txt")
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		tree := New()
		for _, w := range words {
			tree.Insert(w, w)
		}
	}
}

func BenchmarkTreeSearchWords(b *testing.B) {
	b.StopTimer()
	words := loadTestFile("test/assets/words.txt")
	tree := New()
	for _, w := range words {
		tree.Insert(w, w)
	}
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		for _, w := range words {
			tree.Search(w)
		}
	}
}

func BenchmarkTreeIteratorWords(b *testing.B) {
	words := loadTestFile("test/assets/words.txt")
	tree := New()
	for _, w := range words {
		tree.Insert(w, w)
	}
	b.ResetTimer()

	stats := treeStats{}
	for it := tree.Iterator(); it.HasNext(); {
		node, _ := it.Next()
		switch node.Kind() {
		case NODE_4:
			stats.node4Count++
		case NODE_16:
			stats.node16Count++
		case NODE_48:
			stats.node48Count++
		case NODE_256:
			stats.node256Count++
		case NODE_LEAF:
			stats.leafCount++
		}
	}
	assert.Equal(b, treeStats{235886, 111616, 12181, 458, 1}, stats)
}

func BenchmarkTreeForEachWords(b *testing.B) {
	words := loadTestFile("test/assets/words.txt")
	tree := New()
	for _, w := range words {
		tree.Insert(w, w)
	}
	b.ResetTimer()

	stats := treeStats{}
	tree.ForEach(func(node Node) bool {
		switch node.Kind() {
		case NODE_4:
			stats.node4Count++
		case NODE_16:
			stats.node16Count++
		case NODE_48:
			stats.node48Count++
		case NODE_256:
			stats.node256Count++
		case NODE_LEAF:
			stats.leafCount++
		}
		return true
	})
	assert.Equal(b, treeStats{235886, 111616, 12181, 458, 1}, stats)
}

func BenchmarkTreeInsertUUIDs(b *testing.B) {
	words := loadTestFile("test/assets/uuid.txt")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		tree := New()
		for _, w := range words {
			tree.Insert(w, w)
		}
	}
}

func BenchmarkTreeSearchUUIDs(b *testing.B) {
	words := loadTestFile("test/assets/uuid.txt")
	tree := New()
	for _, w := range words {
		tree.Insert(w, w)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, w := range words {
			tree.Search(w)
		}
	}
}

func BenchmarkTreeIteratorUUIDs(b *testing.B) {
	words := loadTestFile("test/assets/uuid.txt")
	tree := New()
	for _, w := range words {
		tree.Insert(w, w)
	}
	b.ResetTimer()

	stats := treeStats{}
	for it := tree.Iterator(); it.HasNext(); {
		node, _ := it.Next()
		switch node.Kind() {
		case NODE_4:
			stats.node4Count++
		case NODE_16:
			stats.node16Count++
		case NODE_48:
			stats.node48Count++
		case NODE_256:
			stats.node256Count++
		case NODE_LEAF:
			stats.leafCount++
		}
	}
	assert.Equal(b, treeStats{100000, 32288, 5120, 0, 0}, stats)
}

func BenchmarkTreeForEachUUIDs(b *testing.B) {
	words := loadTestFile("test/assets/uuid.txt")
	tree := New()
	for _, w := range words {
		tree.Insert(w, w)
	}
	b.ResetTimer()

	stats := treeStats{}
	tree.ForEach(func(node Node) bool {
		switch node.Kind() {
		case NODE_4:
			stats.node4Count++
		case NODE_16:
			stats.node16Count++
		case NODE_48:
			stats.node48Count++
		case NODE_256:
			stats.node256Count++
		case NODE_LEAF:
			stats.leafCount++
		}
		return true
	})
	assert.Equal(b, treeStats{100000, 32288, 5120, 0, 0}, stats)
}
