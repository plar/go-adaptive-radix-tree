/*
Package art implements an Adapative Radix Tree(ART) in pure Go.
Note that this implementation is not thread-safe but it could be really easy to implement.

The design of ART is based on "The Adaptive Radix Tree: ARTful Indexing for Main-Memory Databases" [1]

Also the current implementation was inspired by [2] and [3]

[1] http://db.in.tum.de/~leis/papers/ART.pdf

[2] https://github.com/armon/libart

[3] https://github.com/kellydunn/go-art
*/
package art
