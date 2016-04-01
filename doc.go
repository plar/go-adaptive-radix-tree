/*
Package art implements an adapative radix tree(ART) in pure Go.
Note that this implementation is not thread-safe but it could be really easy to implement.

The design of ART is based on "The Adaptive Radix Tree: ARTful Indexing for Main-Memory Databases" [1]

[1] http://db.in.tum.de/~leis/papers/ART.pdf
*/
package art
