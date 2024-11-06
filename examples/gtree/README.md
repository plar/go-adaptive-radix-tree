# Example of Adaptive Radix Tree with Generics

This directory contains an example of using a generic wrapper around an Adaptive Radix Tree (ART) in Go. 
The implementation demonstrates how to use generics for flexible key and value support.

## Files 

- `gtree.go`: Contains the generic wrapper for the Adaptive Radix Tree (ART) implementation.
- `main.go`: Demonstrates the usage of the generic wrapper with different key and value types.

## Usage

1. Clone the repository:

```bash
$ git clone https://github.com/plar/go-adaptive-radix-tree.git
$ cd go-adaptive-radix-tree/examples/gtree
```

2. Run the example:

```bash
$ go run .
Found: two
Deleted: three
Tree Size: 2
Node Key: 1, Node Value: one
Node Key: 2, Node Value: two
```

## Customizing the Example

- **Key Types**: You can adapt the `convertKeyToBytes` function within `gtree.go` to support different key types beyond `int`.
- **Value Types**: By changing the generics parameters on `GTree` initialization, you can support any value types as needed.

