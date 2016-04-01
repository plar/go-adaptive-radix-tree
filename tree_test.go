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

	l1 := factory.newLeaf(Key("abcdefg12345678"), "abcdefg12345678").leaf()
	l2 := factory.newLeaf(Key("abcdefg!@#$%^&*"), "abcdefg!@#$%^&*").leaf()
	assert.Equal(t, 7, tree.longestCommonPrefix(l1, l2, 0))
	assert.Equal(t, 3, tree.longestCommonPrefix(l1, l2, 4))

	l1 = factory.newLeaf(Key("abcdefg12345678"), "abcdefg12345678").leaf()
	l2 = factory.newLeaf(Key("defg!@#$%^&*"), "defg!@#$%^&*").leaf()
	assert.Equal(t, 0, tree.longestCommonPrefix(l1, l2, 0))
}

func TestTreeInit(t *testing.T) {
	tree := New()
	assert.NotNil(t, tree)
}

// func TestPoolFactory(t *testing.T) {
// 	t.SkipNow()
// 	factory := newPoolObjFactory()
// 	n4 := factory.newNode48()
// 	assert.NotNil(t, n4)
// 	factory.releaseNode(n4)

// 	n4v2 := factory.newNode48()
// 	assert.True(t, n4 == n4v2)

// 	n4v3 := factory.newNode48()
// 	assert.NotEqual(t, n4v2, n4v3)
// }

func TestObjFactory(t *testing.T) {
	factory := newObjFactory()
	n4 := factory.newNode48()
	assert.NotNil(t, n4)
	n4v2 := factory.newNode48()
	assert.True(t, n4 != n4v2)
}

func TestTreeUpdate(t *testing.T) {
	tree := newTree()

	key := Key("key")

	ov, updated := tree.Insert(key, "value")
	assert.Nil(t, ov)
	assert.False(t, updated)
	assert.Equal(t, 1, tree.size)
	assert.Equal(t, Leaf, tree.root.kind)

	v, found := tree.Search(key)
	assert.True(t, found)
	assert.Equal(t, "value", v)

	ov, updated = tree.Insert(key, "otherValue")
	assert.Equal(t, "value", ov)
	assert.True(t, updated)
	assert.Equal(t, 1, tree.size)

	v, found = tree.Search(key)
	assert.True(t, found)
	assert.Equal(t, "otherValue", v)
}

func TestTreeInsertSimilarPrefix(t *testing.T) {
	tree := newTree()

	tree.Insert(Key{1}, 1)
	tree.Insert(Key{1, 1}, 11)

	v, found := tree.Search(Key{1, 1})
	assert.True(t, found)
	assert.Equal(t, 11, v)

}

// An Art Node with a similar prefix should be split into new nodes accordingly
// And should be searchable as intended.
func TestTreeInsert3AndSearchWords(t *testing.T) {
	tree := newTree()

	searchTerms := []string{"A", "a", "aa"}

	for _, term := range searchTerms {
		tree.Insert(Key(term), term)
	}

	for _, term := range searchTerms {
		v, found := tree.Search(Key(term))
		assert.True(t, found)
		assert.Equal(t, term, v)
	}
}

func TestTreeInsertAndGrowToBiggerNode(t *testing.T) {
	var testData = []struct {
		totalNodes byte
		expected   Kind
	}{
		{5, Node16},
		{17, Node48},
		{49, Node256},
	}

	for _, data := range testData {
		tree := newTree()
		for i := byte(0); i < data.totalNodes; i++ {
			tree.Insert(Key{i}, i)
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
		assert.Equal(t, w, v, string(w))
		assert.True(t, found)
	}

	minimum := tree.root.minimum()
	assert.Equal(t, []byte("A"), minimum.value)
	maximum := tree.root.maximum()
	assert.Equal(t, []byte("zythum"), maximum.value)
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

	minimum := tree.root.minimum()
	assert.Equal(t, []byte("00026bda-e0ea-4cda-8245-522764e9f325"), minimum.value)
	maximum := tree.root.maximum()
	assert.Equal(t, []byte("ffffcb46-a92e-4822-82af-a7190f9c1ec5"), maximum.value)
}

func TestTreeInsertAndDeleteOperations(t *testing.T) {
	type testData struct {
		message      string
		insert       interface{}
		delete       interface{}
		size         int
		root         interface{}
		deleteStatus bool
	}
	type treeTestCustom func(data *testData, tree *tree)

	var data = []testData{
		{
			"Insert 1 Delete 1",
			[]string{"test"},
			[]string{"test"},
			0,
			nil,
			true},
		{
			"Insert 2 Delete 1",
			[]string{"test1", "test2"},
			[]string{"test2"},
			1,
			Leaf,
			true},
		{
			"Insert 2 Delete 2",
			[]string{"test1", "test2"},
			[]string{"test1", "test2"},
			0,
			nil,
			true},
		{
			"Insert 5 Delete 1",
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1"},
			4,
			Node4,
			true},
		{
			"Insert 5 Try to delete 1 wrong",
			[]string{"1", "2", "3", "4", "5"},
			[]string{"123"},
			5,
			Node16,
			false},
		{
			"Insert 5 Delete 5",
			[]string{"1", "2", "3", "4", "5"},
			[]string{"1", "2", "3", "4", "5"},
			0,
			nil,
			true},
		{
			"Insert 17 Delete 1",
			[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17"},
			[]string{"2"},
			16,
			Node16,
			true},
		{
			"Insert 17 Delete 17",
			[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17"},
			[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17"},
			0,
			nil,
			true},
		{
			"Insert 49 Delete 0",
			treeTestCustom(func(data *testData, tree *tree) {
				for i := byte(0); i < 49; i++ {
					tree.Insert(Key{i}, []byte{i})
				}
			}),
			[]byte{byte(123)},
			49,
			Node256,
			false},
		{
			"Insert 49 Delete 1",
			treeTestCustom(func(data *testData, tree *tree) {
				for i := byte(0); i < 49; i++ {
					tree.Insert(Key{i}, []byte{i})
				}
			}),
			[]byte{byte(2)},
			48,
			Node48,
			true},
		{
			"Insert 49 Delete 49",
			treeTestCustom(func(data *testData, tree *tree) {
				for i := byte(0); i < 49; i++ {
					tree.Insert(Key{i}, []byte{i})
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
			nil,
			true},
		{
			"Insert 49 Delete 49",
			treeTestCustom(func(data *testData, tree *tree) {
				for i := byte(0); i < 49; i++ {
					tree.Insert(Key{i}, []byte{i})
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
			nil,
			true},
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
			Node256,
			true},
		{
			"Insert 256 Delete 256",
			treeTestCustom(func(data *testData, tree *tree) {
				for i := 0; i < 256; i++ {
					term := fmt.Sprintf("%d", i)
					tree.Insert(Key(term), term)
				}
			}),
			treeTestCustom(func(data *testData, tree *tree) {
				for i := 0; i < 256; i++ {
					term := fmt.Sprintf("%d", i)
					val, deleted := tree.Delete(Key(term))
					assert.Equal(t, data.deleteStatus, deleted)
					assert.Equal(t, term, val)
				}
			}),
			0,
			nil,
			true},
	}

	for _, ds := range data {
		tree := newTree()

		// insert test data...
		if strs, ok := ds.insert.([]string); ok {
			for _, term := range strs {
				tree.Insert(Key(term), term)
			}
		} else if builder, ok := ds.insert.(treeTestCustom); ok {
			builder(&ds, tree)
		}

		// delete test data and check...
		if strs, ok := ds.delete.([]string); ok {
			for _, term := range strs {
				val, deleted := tree.Delete(Key(term))
				assert.Equal(t, ds.deleteStatus, deleted, ds.message)

				if s, ok := val.(string); ok {
					assert.Equal(t, term, s, ds.message)
				}

				_, found := tree.Search(Key(term))
				assert.False(t, found, ds.message)
			}
		} else if bts, ok := ds.delete.([]byte); ok {
			for _, term := range bts {

				val, deleted := tree.Delete(Key{term})
				assert.Equal(t, ds.deleteStatus, deleted, ds.message)

				if deleted {
					assert.Equal(t, []byte{term}, val, ds.message)
				}

				_, found := tree.Search(Key{term})
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

func TestDeleteOneWithSimilarButShorterPrefix(t *testing.T) {
	tree := newTree()
	tree.Insert(Key("keyb::"), "0")
	tree.Insert(Key("keyb::1"), "1")
	tree.Insert(Key("keyb::2"), "2")

	v, deleted := tree.Delete(Key("keyb:"))
	assert.Nil(t, v)
	assert.False(t, deleted)
}

// Inserting a single value into the tree and removing it should result in a nil tree root.
func TestInsertAndDeleteOne(t *testing.T) {
	tree := newTree()
	tree.Insert(Key("test"), "data")
	v, deleted := tree.Delete(Key("test"))
	assert.True(t, deleted)
	assert.Equal(t, "data", v)
	assert.Equal(t, 0, tree.size)
	assert.Nil(t, tree.root)

	tree = newTree()
	tree.Insert(Key("test"), "data")
	v, deleted = tree.Delete(Key("wrong-key"))
	assert.Nil(t, v)
	assert.False(t, deleted)
}

func TestInsertTwoAndDeleteOne(t *testing.T) {
	tree := newTree()
	tree.Insert(Key("2"), 2)
	tree.Insert(Key("1"), 1)

	_, found := tree.Search(Key("2"))
	assert.True(t, found)
	_, found = tree.Search(Key("1"))
	assert.True(t, found)

	v, deleted := tree.Delete(Key("2"))
	assert.True(t, deleted)
	if deleted {
		assert.Equal(t, 2, v)
	}

	assert.Equal(t, 1, tree.size)
	assert.Equal(t, Leaf, tree.root.kind)
}

func TestInsertTwoAndDeleteTwo(t *testing.T) {
	tree := newTree()
	tree.Insert(Key("2"), 2)
	tree.Insert(Key("1"), 1)

	_, found := tree.Search(Key("2"))
	assert.True(t, found)
	_, found = tree.Search(Key("1"))
	assert.True(t, found)

	v, deleted := tree.Delete(Key("2"))
	assert.True(t, deleted)
	assert.Equal(t, 2, v)

	v, deleted = tree.Delete(Key("1"))
	assert.True(t, deleted)
	assert.Equal(t, 1, v)

	assert.Equal(t, 0, tree.size)
	assert.Nil(t, tree.root)
}

func TestTreeTraversalPreordered(t *testing.T) {
	tree := newTree()

	tree.Insert(Key("1"), 1)
	tree.Insert(Key("2"), 2)

	traversal := []Node{}
	tree.ForEach(func(node Node) bool {
		traversal = append(traversal, node)
		return true
	}, TraverseAll)

	assert.Equal(t, 2, tree.size)
	assert.Equal(t, traversal[0], tree.root)
	assert.Nil(t, traversal[0].Key())
	assert.Nil(t, traversal[0].Value())
	assert.NotEqual(t, Leaf, traversal[0].Kind())

	assert.Equal(t, traversal[1].Key(), Key("1"))
	assert.Equal(t, traversal[1].Value(), 1)
	assert.Equal(t, Leaf, traversal[1].Kind())

	assert.Equal(t, traversal[2].Key(), Key("2"))
	assert.Equal(t, traversal[2].Value(), 2)
	assert.Equal(t, Leaf, traversal[2].Kind())

	tree.ForEach(func(node Node) bool {
		assert.Equal(t, Node4, node.Kind())
		return true
	}, TraverseNode)

}

func TestTreeTraversalNode48(t *testing.T) {
	tree := newTree()

	for i := 48; i > 0; i-- {
		tree.Insert(Key{byte(i)}, i)
	}

	traversal := []Node{}
	tree.ForEach(func(node Node) bool {
		traversal = append(traversal, node)
		return true
	}, TraverseAll)

	// Order should be Node48, then the rest of the keys in sorted order
	assert.Equal(t, 48, tree.size)
	assert.Equal(t, traversal[0], tree.root)
	assert.Equal(t, Node48, traversal[0].Kind())

	for i := 1; i < 48; i++ {
		assert.Equal(t, traversal[i].Key(), Key{byte(i)})
		assert.Equal(t, Leaf, traversal[i].Kind())
	}
}

func TestTreeTraversalCancel(t *testing.T) {
	tree := newTree()
	for i := 0; i < 10; i++ {
		tree.Insert(Key{byte(i)}, i)
	}
	assert.Equal(t, 10, tree.Size())

	// delete 5 nodes and cancel
	curNode := 0
	nodes := make([]Node, 0, 5)
	tree.ForEach(func(node Node) bool {
		if curNode >= 5 {
			return false
		}
		nodes = append(nodes, node)
		curNode++
		return true
	}, TraverseAll)

	assert.Equal(t, 5, len(nodes))

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
	}, TraverseAll)

	assert.Equal(t, treeStats{235886, 111616, 12181, 458, 1}, stats)
}

func TestTreeTraversalPrefix(t *testing.T) {
	dataSet := []struct {
		keyPrefix string
		keys      []string
		expected  []string
	}{
		{
			"",
			[]string{},
			[]string{},
		},
		{
			"api",
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "abc.123.456", "api.foo", "api"},
			[]string{"api", "api.foe.fum", "api.foo", "api.foo.bar", "api.foo.baz"},
		}, {
			"a",
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "abc.123.456", "api.foo", "api"},
			[]string{"abc.123.456", "api", "api.foe.fum", "api.foo", "api.foo.bar", "api.foo.baz"},
		}, {
			"b",
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "abc.123.456", "api.foo", "api"},
			[]string{},
		},
		{
			"api.",
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "abc.123.456", "api.foo", "api"},
			[]string{"api.foe.fum", "api.foo", "api.foo.bar", "api.foo.baz"},
		},
		{
			"api.foo.bar",
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "abc.123.456", "api.foo", "api"},
			[]string{"api.foo.bar"},
		},
		{
			"api.end",
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "abc.123.456", "api.foo", "api"},
			[]string{},
		}, {
			"",
			[]string{"api.foo.bar", "api.foo.baz", "api.foe.fum", "abc.123.456", "api.foo", "api"},
			[]string{"abc.123.456", "api", "api.foe.fum", "api.foo", "api.foo.bar", "api.foo.baz"},
		}, {
			"this:key:has",
			[]string{
				"this:key:has:a:long:prefix:3",
				"this:key:has:a:long:common:prefix:2",
				"this:key:has:a:long:common:prefix:1",
			},
			[]string{
				"this:key:has:a:long:common:prefix:1",
				"this:key:has:a:long:common:prefix:2",
				"this:key:has:a:long:prefix:3",
			},
		},
	}

	for _, d := range dataSet {
		tree := New()
		for _, k := range d.keys {
			tree.Insert(Key(k), string(k))
		}

		actual := []string{}
		tree.ForEachPrefix(Key(d.keyPrefix), func(node Node) bool {
			if node.Kind() != Leaf {
				return true
			}
			actual = append(actual, string(node.Key()))
			return true
		})

		assert.Equal(t, d.expected, actual, d.keyPrefix)
	}

}

func TestTreeTraversalForEachPrefixWithSimilarKey(t *testing.T) {
	tree := newTree()
	tree.Insert(Key("abc0"), "0")
	tree.Insert(Key("abc1"), "1")
	tree.Insert(Key("abc2"), "2")

	tree.ForEachPrefix(Key("abc3"), func(node Node) bool {
		return true
	})
}

func TestTreeTraversalPrefixWords(t *testing.T) {
	t.SkipNow()
	words := loadTestFile("test/assets/words.txt")
	tree := New()
	for _, w := range words {
		tree.Insert(w, string(w))
	}

	tree.ForEachPrefix(Key("antis"), func(node Node) bool {
		if node.Kind() != Leaf {
			return true
		}

		fmt.Println(string(node.Key()))
		// if node.Value() != nil {
		// 	fmt.Println("\tvalue=", node.Value().(string))
		// }
		return true
	})
}

func TestTreeIterator(t *testing.T) {
	tree := newTree()
	tree.Insert(Key("2"), []byte{2})
	tree.Insert(Key("1"), []byte{1})

	it := tree.Iterator(TraverseAll)
	assert.NotNil(t, it)
	assert.True(t, it.HasNext())
	n4, err := it.Next()
	assert.NoError(t, err)
	assert.Equal(t, Node4, n4.Kind())

	assert.True(t, it.HasNext())
	v1, err := it.Next()
	assert.NoError(t, err)
	assert.Equal(t, v1.Value(), []byte{1})

	assert.True(t, it.HasNext())
	v2, err := it.Next()
	assert.NoError(t, err)
	assert.Equal(t, v2.Value(), []byte{2})

	assert.False(t, it.HasNext())
	bad, err := it.Next()
	assert.Nil(t, bad)
	assert.Equal(t, ErrNoMoreNodes, err)

}

func TestTreeIteratorConcurrentModification(t *testing.T) {
	tree := newTree()
	tree.Insert(Key("2"), []byte{2})
	tree.Insert(Key("1"), []byte{1})

	it1 := tree.Iterator(TraverseAll)
	assert.NotNil(t, it1)
	assert.True(t, it1.HasNext())

	// simulate concurrent modification
	tree.Insert(Key("3"), []byte{3})
	bad, err := it1.Next()
	assert.Nil(t, bad)
	assert.Equal(t, ErrConcurrentModification, err)

	it2 := tree.Iterator(TraverseAll)
	assert.NotNil(t, it2)
	assert.True(t, it2.HasNext())

	tree.Delete([]byte("3"))
	bad, err = it2.Next()
	assert.Nil(t, bad)
	assert.Equal(t, ErrConcurrentModification, err)

	// test buffered ConcurrentModification
	it3 := tree.Iterator(TraverseNode)
	assert.NotNil(t, it3)
	tree.Insert(Key("3"), []byte{3})
	assert.True(t, it3.HasNext())
	bad, err = it3.Next()
	assert.Nil(t, bad)
	assert.Equal(t, ErrConcurrentModification, err)
}

func TestTreeIterateWordsStats(t *testing.T) {
	words := loadTestFile("test/assets/words.txt")
	tree := New()
	for _, w := range words {
		tree.Insert(w, w)
	}

	stats := treeStats{}

	for it := tree.Iterator(TraverseAll); it.HasNext(); {
		node, _ := it.Next()
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
	}

	assert.Equal(t, treeStats{235886, 111616, 12181, 458, 1}, stats)

	stats = treeStats{}
	for it := tree.Iterator(); it.HasNext(); {
		node, _ := it.Next()
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
	}
	assert.Equal(t, treeStats{235886, 0, 0, 0, 0}, stats)

	stats = treeStats{}
	for it := tree.Iterator(TraverseNode); it.HasNext(); {
		node, _ := it.Next()
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
	}
	assert.Equal(t, treeStats{0, 111616, 12181, 458, 1}, stats)
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
		assert.Equal(t, w, v)
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
		assert.Equal(t, w, v)
	}

	assert.Equal(t, 0, tree.size)
	assert.Nil(t, tree.root)
}

func TestTreeAPI(t *testing.T) {
	// test empty tree
	tree := New()
	assert.NotNil(t, tree)
	assert.Equal(t, 0, tree.Size())

	oldValue, deleted := tree.Delete(Key("non existent key"))
	assert.Nil(t, oldValue)
	assert.False(t, deleted)

	minValue, found := tree.Minimum()
	assert.Nil(t, minValue)
	assert.False(t, found)

	maxValue, found := tree.Maximum()
	assert.Nil(t, maxValue)
	assert.False(t, found)

	value, found := tree.Search(Key("non existent key"))
	assert.Nil(t, value)
	assert.False(t, found)

	it := tree.Iterator(TraverseAll)
	assert.NotNil(t, it)
	assert.False(t, it.HasNext())
	_, err := it.Next()
	assert.Error(t, err)

	tree.ForEach(func(node Node) bool {
		assert.Fail(t, "Should be never called")
		return true
	})

	// test tree with data
	tree = New()
	oldValue, updated := tree.Insert(Key("Hi, I'm Key"), "Nice to meet you, I'm Value")
	assert.Nil(t, oldValue)
	assert.False(t, updated)
	assert.Equal(t, 1, tree.Size())

	value, found = tree.Search(Key("Hi, I'm Key"))
	assert.True(t, found)
	assert.Equal(t, "Nice to meet you, I'm Value", value)

	oldValue, updated = tree.Insert(Key("Hi, I'm Key"), "Ha-ha, Again? Go away!")
	assert.Equal(t, "Nice to meet you, I'm Value", oldValue)
	assert.True(t, updated)
	assert.Equal(t, 1, tree.Size())

	value, found = tree.Search(Key("Hi, I'm Key"))
	assert.True(t, found)
	assert.Equal(t, "Ha-ha, Again? Go away!", value)

	tree.ForEach(func(node Node) bool {
		assert.Equal(t, Key("Hi, I'm Key"), node.Key())
		assert.Equal(t, "Ha-ha, Again? Go away!", node.Value())
		return true
	})

	it = tree.Iterator(TraverseAll)
	assert.NotNil(t, it)
	assert.True(t, it.HasNext())
	next, err := it.Next()
	assert.NoError(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, "Ha-ha, Again? Go away!", next.Value())

	assert.False(t, it.HasNext())
	next, err = it.Next()
	assert.Nil(t, next)
	assert.Error(t, err)

	oldValue, updated = tree.Insert(Key("Hi, I'm Value"), "Surprise!")
	assert.Nil(t, oldValue)
	assert.False(t, updated)

	tree.Insert(Key("Now I know..."), "What?")
	tree.Insert(Key("ABC"), "ABC")
	tree.Insert(Key("DEF"), "DEF")
	tree.Insert(Key("XYZ"), "XYZ")

	minValue, found = tree.Minimum()
	assert.Equal(t, "ABC", minValue)
	assert.True(t, found)

	maxValue, found = tree.Maximum()
	assert.Equal(t, "XYZ", maxValue)
	assert.True(t, found)
}

//
// Benchmarks
//
func BenchmarkTreeInsertWords(b *testing.B) {
	words := loadTestFile("test/assets/words.txt")
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		tree := New()
		for _, w := range words {
			tree.Insert(w, w)
		}
	}
}

func BenchmarkTreeSearchWords(b *testing.B) {
	words := loadTestFile("test/assets/words.txt")
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

func BenchmarkTreeIteratorWords(b *testing.B) {
	words := loadTestFile("test/assets/words.txt")
	tree := New()
	for _, w := range words {
		tree.Insert(w, w)
	}
	b.ResetTimer()

	stats := treeStats{}
	for it := tree.Iterator(TraverseAll); it.HasNext(); {
		node, _ := it.Next()
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
	}, TraverseAll)
	assert.Equal(b, treeStats{235886, 111616, 12181, 458, 1}, stats)

	stats = treeStats{}
	tree.ForEach(func(node Node) bool {
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
	}, TraverseLeaf)
	assert.Equal(b, treeStats{235886, 0, 0, 0, 0}, stats)

	stats = treeStats{}
	tree.ForEach(func(node Node) bool {
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
	}, TraverseNode)
	assert.Equal(b, treeStats{0, 111616, 12181, 458, 1}, stats)
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
	for it := tree.Iterator(TraverseAll); it.HasNext(); {
		node, _ := it.Next()
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
	}, TraverseAll)
	assert.Equal(b, treeStats{100000, 32288, 5120, 0, 0}, stats)
}
