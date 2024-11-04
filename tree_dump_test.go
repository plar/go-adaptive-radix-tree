package art

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var updateGolden = flag.Bool("update-golden", false, "update .golden files")

func TestTreeStringer(t *testing.T) {
	tests := []struct {
		name   string
		tree   func() *tree
		golden string
	}{
		{
			name: "Dump4",
			tree: func() *tree {
				n4 := factory.newNode4()
				n4leaf := factory.newLeaf([]byte("key4"), "value4")
				n4.addChild('k', true, n4leaf)
				return &tree{root: n4}
			},
			golden: "test/stringer/dump4.golden",
		},
		{
			name: "Dump4BinaryValue",
			tree: func() *tree {
				n4 := factory.newNode4()
				n4leaf := factory.newLeaf([]byte("key4"), []byte("value4"))
				n4.addChild('k', true, n4leaf)
				return &tree{root: n4}
			},
			golden: "test/stringer/dump4_binary_value.golden",
		},
		{
			name: "Dump4Int",
			tree: func() *tree {
				n4 := factory.newNode4()
				n4leaf := factory.newLeaf([]byte("key4"), 4)
				n4.addChild('k', true, n4leaf)
				return &tree{root: n4}
			},
			golden: "test/stringer/dump4_int.golden",
		},
		{
			name: "Dump16IntValue",
			tree: func() *tree {
				n16 := factory.newNode16()
				n16_2 := factory.newNode16()
				n16_2leaf := factory.newLeaf([]byte("4yek"), 4)
				n16_2.addChild('z', true, n16_2leaf)

				n16leaf := factory.newLeaf([]byte("key4"), 4)
				c4leaf := factory.newLeaf([]byte("cey4"), 44)
				n16.addChild('k', true, n16leaf)
				n16.addChild('c', true, c4leaf)
				n16.addChild('z', true, n16_2)
				return &tree{root: n16}
			},
			golden: "test/stringer/dump16_int_value.golden",
		},
		{
			name: "Dump16",
			tree: func() *tree {
				n16 := factory.newNode16()
				n16leaf := factory.newLeaf([]byte("key16"), "value16")
				n16.addChild('k', true, n16leaf)
				return &tree{root: n16}
			},
			golden: "test/stringer/dump16.golden",
		},
		{
			name: "Dump48",
			tree: func() *tree {
				n48 := factory.newNode48()
				n48leaf := factory.newLeaf([]byte("key48"), "value48")
				n48.addChild('k', true, n48leaf)
				return &tree{root: n48}
			},
			golden: "test/stringer/dump48.golden",
		},
		{
			name: "Dump256",
			tree: func() *tree {
				n256 := factory.newNode256()
				n256leaf := factory.newLeaf([]byte("key256"), "value256")
				n256.addChild('k', true, n256leaf)
				return &tree{root: n256}
			},
			golden: "test/stringer/dump256.golden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualOut := TreeStringer(tt.tree())

			if *updateGolden {
				t.Logf("%s: updating golden file %s...", tt.name, tt.golden)
				err := os.WriteFile(tt.golden, []byte(actualOut), 0644)
				t.Logf("%s: updated golden file %s: err=%v", tt.name, tt.golden, err)
				require.NoError(t, err)
			}

			goldenOut, err := os.ReadFile(tt.golden)
			require.NoError(t, err)
			assert.Equal(t, string(goldenOut), actualOut)
		})
	}
}
