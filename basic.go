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

// enums for the left and right sides of the tree.ba
const (
	lo = iota
	hi
)

func imax(a b int) int {
	if a < b {
		return b
	}
	return a
}

// BasicBST is a basic unoptimised unbalanced BST.
type BasicBST struct {
	Key    KeyType
	Value  interface{}
	Parent *BasicBST
	Child  [2]*BasicBST // index is oneof {lo, hi}
	Height int
}

func (n *BasicBST) IsSentinel() bool {
	return n != nil && n.Parent == n
}

func (n *BasicBST) height() int {
	if n == nil || n.IsSentinel() {
		return -1
	}
	return n.Height
}

func (n *BasicBST) calcHeight() int {
	if n != nil {
		n.Height = 1 + imax(n.Child[lo].height(), n.Child[hi].height())
	}
	return n.height()
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

// which returns the node's index from its parent.
func (n *BasicBST) which() int {
	switch p := n.Parent; {
	case n == p.Child[lo]:
		return lo
	case n == p.Child[hi]:
		return hi
	default:
		return -1
	}
}

// opposite reverses a direction.
func opposite(d int) int {
	return (d + 1) % 2
}

// next returns the next tree node in the given direction.
func (n *BasicBST) next(d int) *BasicBST {
	r := opposite(d)
	if n.Child[d] != nil {
		cur := n.Child[d]
		for cur.Child[r] != nil {
			cur = cur.Child[r]
		}
		return cur
	}
	cur := n
	for cur.which() == d {
		cur = cur.Parent
	}
	if cur.Parent.IsSentinel() {
		return nil
	}
	return cur.Parent
}

// Next returns the next node.
func (n *BasicBST) Next() *BasicBST {
	return n.next(hi)
}

// Prev returns the previous node.
func (n *BasicBST) Prev() *BasicBST {
	return n.next(lo)
}

// Delete removes a node from the tree.
func (n *BasicBST) Delete() {
	switch {
	case n.IsSentinel():
		return
	case n == nil:
		return
	case n.Child[hi] == nil:
		n.Parent.Child[n.which()] = n.Child[lo]
	case n.Child[lo] == nil:
		n.Parent.Child[n.which()] = n.Child[hi]
	default:
		cur := n.Child[hi]
		for cur.Child[lo] != nil {
			cur = cur.Child[lo]
		}
		n.Key = cur.Key
		n.Value = cur.Value
		cur.Delete()
	}
}
