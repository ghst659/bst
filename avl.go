// Package bst implements a word prefix tree.
package bst

import (
	"context"
	"fmt"
	"io"
)

// AVL is a basic unoptimised unbalanced BST.
type AVL struct {
	Key    KeyType
	Value  interface{}
	Parent *AVL
	Child  [2]*AVL // index is oneof {lo, hi}
	Height int
}

func (n *AVL) height() int {
	if n == nil || n.IsSentinel() {
		return -1
	}
	return n.Height
}

func (n *AVL) updateHeight() int {
	if n != nil && !n.IsSentinel() {
		n.Height = 1 + imax(n.Child[lo].height(), n.Child[hi].height())
	}
	return n.height()
}

func (n *AVL) IsSentinel() bool {
	return n != nil && n.Parent == n
}

// NewAVL allocates a new BasiccBST.
func NewAVL() *AVL {
	sentinel := &AVL{}
	sentinel.Parent = sentinel
	return sentinel
}

// Get retrieves a pointer to a AVL node for a given key.
func (n *AVL) Get(k KeyType) *AVL {
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
func (n *AVL) Visit(f func(n *AVL) error) error {
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
func (n *AVL) Viz(iow io.Writer) {
	iow.Write([]byte("digraph treemap {\n"))
	defer iow.Write([]byte("}\n"))
	n.Child[lo].Visit(func(n *AVL) error {
		if n != nil {
			if n.Child[lo] != nil {
				text := fmt.Sprintf("  %s(%d):w -> %s(%d):n [label=\"lo\"];\n",
					n.Key.String(), n.height(),
					n.Child[lo].Key.String(), n.Child[lo].height())
				iow.Write([]byte(text))
			}
			if n.Child[hi] != nil {
				text := fmt.Sprintf("  %s(%d):e -> %s(%d):n [label=\"hi\"];\n",
					n.Key.String(), n.height(),
					n.Child[hi].Key.String(), n.Child[hi].height())
				iow.Write([]byte(text))
			}
		}
		return nil
	})
}

// Keys returns a channel to stream the keys from low to high.
func (n *AVL) Keys(ctx context.Context) chan KeyType {
	keys := make(chan KeyType)
	go func() {
		defer close(keys)
		n.Visit(func(n *AVL) error {
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
func (n *AVL) Check(ctx context.Context) chan *AVL {
	nodes := make(chan *AVL)
	go func() {
		defer close(nodes)
		n.Visit(func(n *AVL) error {
			badLo := (n.Child[lo] != nil && !n.Child[lo].Key.Less(n.Key))
			badHi := (n.Child[hi] != nil && !n.Key.Less(n.Child[hi].Key))
			badBal := (n.Child[lo].height() - n.Child[hi].height())
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
func (n *AVL) Insert(k KeyType, v interface{}) {
	switch {
	case n.IsSentinel() || k.Less(n.Key):
		if n.Child[lo] == nil {
			n.Child[lo] = &AVL{
				Key:    k,
				Value:  v,
				Parent: n,
			}
		} else {
			n.Child[lo].Insert(k, v)
		}
	case n.Key.Less(k):
		if n.Child[hi] == nil {
			n.Child[hi] = &AVL{
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
func (n *AVL) which() int {
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
func (n *AVL) next(d int) *AVL {
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
func (n *AVL) Next() *AVL {
	return n.next(hi)
}

// Prev returns the previous node.
func (n *AVL) Prev() *AVL {
	return n.next(lo)
}

// Delete removes a node from the tree.
func (n *AVL) Delete() {
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

func imax(a, b int) int {
	if b > a {
		return b
	}
	return a
}

func iabs(k int) int {
	if k < 0 {
		return -k
	}
	return k
}
