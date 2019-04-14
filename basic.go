// Package bst implements a word prefix tree.
package bst

import (
	"context"
	"fmt"
	"io"
)

// KeyType is the interface required from BST keys.
type KeyType interface {
	Equal(KeyType) bool
	Less(KeyType) bool
	String() string
}

// enums for the left and right sides of the tree.
const (
	lo = iota
	hi
)

// BasicBST is a basic unoptimised unbalanced BST.
type BasicBST struct {
	Key    KeyType
	Value  interface{}
	Parent *BasicBST
	Child  [2]*BasicBST // index is oneof {lo, hi}
}

func (n *BasicBST) IsSentinel() bool {
	return n != nil && n.Parent == n
}

// NewBasic allocates a new BasiccBST.
func NewBasic() *BasicBST {
	sentinel := &BasicBST{}
	sentinel.Parent = sentinel
	return sentinel
}

// Get retrieves a pointer to a BasicBST node for a given key.
func (n *BasicBST) Get(k KeyType) *BasicBST {
	switch {
	case n == nil:
		return nil
	case n.IsSentinel() || k.Less(n.Key):
		return n.Child[lo].Get(k)
	case n.Key.Less(k):
		return n.Child[hi].Get(k)
	default:
		return n
	}
}

// Visit visits the BST nodes in tree order.
func (n *BasicBST) Visit(f func(n *BasicBST) error) error {
	if n == nil {
		return nil
	}
	if n.IsSentinel() {
		return n.Child[lo].Visit(f)
	}
	if n.Child[lo] != nil {
		if err := n.Child[lo].Visit(f); err != nil {
			return err
		}
	}
	if err := f(n); err != nil {
		return err
	}
	if n.Child[hi] != nil {
		if err := n.Child[hi].Visit(f); err != nil {
			return err
		}
	}
	return nil
}

// Viz writes a DOT visualisation of the graph to an io.Writer
func (n *BasicBST) Viz(iow io.Writer) {
	iow.Write([]byte("digraph treemap {\n"))
	defer iow.Write([]byte("}\n"))
	n.Child[lo].Visit(func(n *BasicBST) error {
		if n != nil {
			if n.Child[lo] != nil {
				text := fmt.Sprintf("  %s:w -> %s:n [label=\"lo\"];\n",
					n.Key.String(), n.Child[lo].Key.String())
				iow.Write([]byte(text))
			}
			if n.Child[hi] != nil {
				text := fmt.Sprintf("  %s:e -> %s:n [label=\"hi\"];\n",
					n.Key.String(), n.Child[hi].Key.String())
				iow.Write([]byte(text))
			}
			// if !n.Parent.IsSentinel() {
			// 	text := fmt.Sprintf("  %s -> %s [label=\"parent\", style=dashed];\n",
			// 		n.Key.String(), n.Parent.Key.String())
			// 	iow.Write([]byte(text))
			// }
		}
		return nil
	})
}

// Keys returns a channel to stream the keys from low to high.
func (n *BasicBST) Keys(ctx context.Context) chan KeyType {
	keys := make(chan KeyType)
	go func() {
		defer close(keys)
		n.Visit(func(n *BasicBST) error {
			select {
			case keys <- n.Key:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()
	return keys
}

// Check returns a channel of nodes violating the BST condition.
func (n *BasicBST) Check(ctx context.Context) chan *BasicBST {
	nodes := make(chan *BasicBST)
	go func() {
		defer close(nodes)
		n.Visit(func(n *BasicBST) error {
			badLo := (n.Child[lo] != nil && !n.Child[lo].Key.Less(n.Key))
			badHi := (n.Child[hi] != nil && !n.Key.Less(n.Child[hi].Key))
			if badLo || badHi {
				select {
				case nodes <- n:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}()
	return nodes
}

// Insert inserts a key, value pair into the BST.
func (n *BasicBST) Insert(k KeyType, v interface{}) {
	switch {
	case n.IsSentinel() || k.Less(n.Key):
		if n.Child[lo] == nil {
			n.Child[lo] = &BasicBST{
				Key:    k,
				Value:  v,
				Parent: n,
			}
		} else {
			n.Child[lo].Insert(k, v)
		}
	case n.Key.Less(k):
		if n.Child[hi] == nil {
			n.Child[hi] = &BasicBST{
				Key:    k,
				Value:  v,
				Parent: n,
			}
		} else {
			n.Child[hi].Insert(k, v)
		}
	default:
		n.Value = v
	}
}

// Which returns the node's index from its parent.
func (n *BasicBST) Which() int {
	switch p := n.Parent; {
	case n == p.Child[lo]:
		return lo
	case n == p.Child[hi]:
		return hi
	default:
		return -1
	}
}

// Next returns the next node.
func (n *BasicBST) Next() *BasicBST {
	if n.Child[hi] != nil {
		cur := n.Child[hi]
		for cur.Child[lo] != nil {
			cur = cur.Child[lo]
		}
		return cur
	}
	cur := n
	for cur.Which() == hi {
		cur = cur.Parent
	}
	if cur.Parent.IsSentinel() {
		return nil
	}
	return cur.Parent
}

// Delete removes a node from the tree.
func (n *BasicBST) Delete() {
	switch {
	case n.IsSentinel():
		return
	case n == nil:
		return
	default:
		// TODO
	}
}
