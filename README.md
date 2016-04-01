An Adaptive Radix Tree Implementation in Go

# What

This library provides a Go implementation of the Adaptive Radix Tree (ART).


Its lookup performance surpasses highly tuned, read-only search trees,
while supporting very efficient insertions and deletions as well.
At the same time, ART is very space efficient and solves the problem of excessive
worst-case space consumption, which plagues most radix trees, by adaptively
choosing compact and efficient data structures for internal nodes.
Even though ART’s performance is comparable to hash tables, it maintains the data in sorted
order, which enables additional operations like range scan and prefix lookup.

# References

http://db.in.tum.de/~leis/papers/ART.pdf


benchmark

====================================================================================================
go-art:
BenchmarkTreeInsertWords-8         5     265,113,838 ns/op    81,629,144 B/op    2,547,335 allocs/op
BenchmarkTreeSearchWords-8        10     128,191,977 ns/op    13,272,281 B/op    1,659,034 allocs/op
BenchmarkTreeInsertUUIDs-8        10     142,224,638 ns/op    33,677,929 B/op      874,534 allocs/op
BenchmarkTreeSearchUUIDs-8        20      80,662,462 ns/op     3,883,116 B/op      485,389 allocs/op

====================================================================================================
my own:
BenchmarkTreeInsertWords-8        10     168,163,741 ns/op    40,316,990 B/op    1,218,302 allocs/op
BenchmarkTreeSearchWords-8        30      48,040,850 ns/op             0 B/op            0 allocs/op
BenchmarkTreeInsertUUIDs-8        10     100,045,425 ns/op    18,974,657 B/op      485,106 allocs/op
BenchmarkTreeSearchUUIDs-8        30      39,716,041 ns/op             0 B/op            0 allocs/op

// Adjusted SHRINK
BenchmarkTreeInsertWords-8        10     173,875,402 ns/op    40,317,054 B/op    1,218,310 allocs/op
BenchmarkTreeSearchWords-8        30      51,761,378 ns/op             0 B/op            0 allocs/op
BenchmarkTreeInsertUUIDs-8        10     101,109,316 ns/op    18,974,790 B/op      485,101 allocs/op
BenchmarkTreeSearchUUIDs-8        30      39,084,196 ns/op             0 B/op            0 allocs/op

my own (pre-pool)
BenchmarkTreeInsertWords-8         3     391626106 ns/op    40335234 B/op    1218341 allocs/op
BenchmarkTreeSearchWords-8        30      48167925 ns/op           0 B/op          0 allocs/op
BenchmarkTreeInsertUUIDs-8        10     183417698 ns/op    18981409 B/op     485104 allocs/op
BenchmarkTreeSearchUUIDs-8        30      42070837 ns/op           0 B/op          0 allocs/op


BenchmarkTreeInsertWords-8        10     169009795 ns/op    40317076 B/op    1218314 allocs/op
BenchmarkTreeSearchWords-8        30      48046529 ns/op           0 B/op          0 allocs/op
BenchmarkTreeInsertUUIDs-8        20      99827708 ns/op    18974729 B/op     485104 allocs/op
BenchmarkTreeSearchUUIDs-8        30      39526190 ns/op           0 B/op          0 allocs/op

// change prefix struct
BenchmarkTreeInsertWords-8        10     160488513 ns/op    38328852 B/op    1218299 allocs/op
BenchmarkTreeSearchWords-8        30      47648962 ns/op           0 B/op          0 allocs/op
BenchmarkTreeInsertUUIDs-8        20      97998414 ns/op    18376099 B/op     485103 allocs/op
BenchmarkTreeSearchUUIDs-8        50      36681493 ns/op           0 B/op          0 allocs/op


BenchmarkTreeInsertWords-8        10     166066047 ns/op    38126688 B/op    1218309 allocs/op
BenchmarkTreeSearchWords-8        30      47627667 ns/op           0 B/op          0 allocs/op
BenchmarkTreeInsertUUIDs-8        20      98856077 ns/op    18294329 B/op     485109 allocs/op
BenchmarkTreeSearchUUIDs-8        30      36800596 ns/op           0 B/op          0 allocs/op

// SSE
BenchmarkTreeInsertWords-8         5     292191220 ns/op    40317161 B/op    1218319 allocs/op
BenchmarkTreeSearchWords-8        10     169534853 ns/op           0 B/op          0 allocs/op
BenchmarkTreeInsertUUIDs-8        10     168494986 ns/op    18974712 B/op     485109 allocs/op
BenchmarkTreeSearchUUIDs-8        10     112584337 ns/op           0 B/op          0 allocs/op


