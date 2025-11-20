package art

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Benchmarks for the tree implementation.
func BenchmarkWordsTreeInsert(b *testing.B) {
	words := loadTestFile("test/assets/words.txt")

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		tree := New()
		for _, w := range words {
			tree.Insert(w, w)
		}
	}
}

func BenchmarkWordsTreeSearch(b *testing.B) {
	tree := New()

	words := loadTestFile("test/assets/words.txt")
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

func BenchmarkWordsTreeIterator(b *testing.B) {
	tree := New()

	words := loadTestFile("test/assets/words.txt")
	for _, w := range words {
		tree.Insert(w, w)
	}

	b.ResetTimer()

	stats := collectStats(tree.Iterator(TraverseAll))
	assert.Equal(b, treeStats{235886, 113419, 10433, 403, 1}, stats)
}

func BenchmarkWordsTreeForEach(b *testing.B) {
	tree := New()

	words := loadTestFile("test/assets/words.txt")
	for _, w := range words {
		tree.Insert(w, w)
	}

	b.ResetTimer()

	stats := treeStats{}
	tree.ForEach(stats.processStats, TraverseAll)
	assert.Equal(b, treeStats{235886, 113419, 10433, 403, 1}, stats)

	stats = treeStats{}
	tree.ForEach(stats.processStats, TraverseLeaf)
	assert.Equal(b, treeStats{235886, 0, 0, 0, 0}, stats)

	stats = treeStats{}
	tree.ForEach(stats.processStats, TraverseNode)
	assert.Equal(b, treeStats{0, 113419, 10433, 403, 1}, stats)
}

func BenchmarkUUIDsTreeInsert(b *testing.B) {
	words := loadTestFile("test/assets/uuid.txt")

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		tree := New()
		for _, w := range words {
			tree.Insert(w, w)
		}
	}
}

func BenchmarkUUIDsTreeSearch(b *testing.B) {
	tree := New()

	words := loadTestFile("test/assets/uuid.txt")
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

func BenchmarkUUIDsTreeIterator(b *testing.B) {
	tree := New()

	words := loadTestFile("test/assets/uuid.txt")
	for _, w := range words {
		tree.Insert(w, w)
	}

	b.ResetTimer()

	stats := collectStats(tree.Iterator(TraverseAll))
	assert.Equal(b, treeStats{100000, 32288, 5120, 0, 0}, stats)
}

func BenchmarkUUIDsTreeForEach(b *testing.B) {
	tree := New()

	words := loadTestFile("test/assets/uuid.txt")
	for _, w := range words {
		tree.Insert(w, w)
	}

	b.ResetTimer()

	stats := collectStats(tree.Iterator(TraverseAll))
	assert.Equal(b, treeStats{100000, 32288, 5120, 0, 0}, stats)
}

func BenchmarkHSKTreeInsert(b *testing.B) {
	words := loadTestFile("test/assets/hsk_words.txt")

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		tree := New()
		for _, w := range words {
			tree.Insert(w, w)
		}
	}
}

func BenchmarkHSKTreeSearch(b *testing.B) {
	tree := New()

	words := loadTestFile("test/assets/hsk_words.txt")
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

func BenchmarkHSKTreeIterator(b *testing.B) {
	tree := New()

	words := loadTestFile("test/assets/hsk_words.txt")
	for _, w := range words {
		tree.Insert(w, w)
	}

	b.ResetTimer()

	stats := collectStats(tree.Iterator(TraverseAll))
	assert.Equal(b, treeStats{4995, 1630, 276, 21, 4}, stats)
}

func BenchmarkHSKTreeForEach(b *testing.B) {
	tree := New()

	words := loadTestFile("test/assets/hsk_words.txt")
	for _, w := range words {
		tree.Insert(w, w)
	}

	b.ResetTimer()

	stats := collectStats(tree.Iterator(TraverseAll))
	assert.Equal(b, treeStats{4995, 1630, 276, 21, 4}, stats)
}
