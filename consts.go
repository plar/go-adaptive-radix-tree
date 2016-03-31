package art

// node types
const (
	NODE_4    = kind(1)
	NODE_16   = kind(2)
	NODE_48   = kind(3)
	NODE_256  = kind(4)
	NODE_LEAF = kind(5)
)

// node constraits
const (
	NODE_4_SHRINK = 2
	NODE_4_MIN    = 2
	NODE_4_MAX    = 4

	NODE_16_SHRINK = 5 // 3
	NODE_16_MIN    = 5
	NODE_16_MAX    = 16

	NODE_48_SHRINK = 17 // 12
	NODE_48_MIN    = 17
	NODE_48_MAX    = 48

	NODE_256_SHRINK = 49 // 37
	NODE_256_MIN    = 49
	NODE_256_MAX    = 256
)

const (
	MAX_PREFIX_LENGTH = 10
)
