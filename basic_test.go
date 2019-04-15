package bst

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"testing"
)

type iKey int

func (a iKey) Equal(b KeyType) bool {
	return a == b.(iKey)
}

func (a iKey) Less(b KeyType) bool {
	return a < b.(iKey)
}

func (a iKey) String() string {
	return strconv.Itoa(int(a))
}

func TestBasicEmpty(t *testing.T) {
	s := NewBasic()
	if n := s.Get(iKey(5)); n != nil {
		t.Errorf("unexpected content in empty graph:  %v", *n)
	}
}

const arySize = 64

func TestRandomInsert(t *testing.T) {
	kvs := make([]int, 0, 64)
	for i := 0; i < arySize; i++ {
		kvs = append(kvs, i)
	}
	rand.Shuffle(len(kvs), func(i, j int) {
		kvs[i], kvs[j] = kvs[j], kvs[i]
	})
	s := NewBasic()
	for _, k := range kvs {
		v := -k
		s.Insert(iKey(k), v)
	}
	t.Run("Viz", func(t *testing.T) {
		ofp, err := os.Create("/tmp/shuf.dot")
		if err != nil {
			t.Error("cannot create file")
		}
		defer ofp.Close()
		s.Viz(ofp)
	})
	t.Run("Next", func(t *testing.T) {
		for i := 0; i < arySize; i++ {
			b := s.Get(iKey(i))
			got := b.Next()
			if i == arySize-1 {
				if got != nil {
					t.Errorf("unexpected non-nil next: %+v", *got)
				}
				t.Logf("Next(%d) == nil", i)
			} else {
				if got == nil {
					t.Errorf("unexpected nil Next(%d)", i)
				}
				t.Logf("Next(%d) == %+v", i, *got)
				v := got.Value.(int)
				want := -(i + 1)
				if v != want {
					t.Errorf("bad Next(%d).Value: got %d, want %d", i, v, want)
				}
			}
		}
	})
}

func TestBasicInsertAndVisit(t *testing.T) {
	s := NewBasic()
	kvs := []struct {
		k iKey
		v int
	}{
		{k: 3, v: -3},
		{k: 1, v: -1},
		{k: 5, v: -5},
		{k: 0, v: 0},
		{k: 2, v: -2},
		{k: 4, v: -4},
		{k: 6, v: -6},
	}
	for _, p := range kvs {
		s.Insert(p.k, p.v)
	}
	t.Run("Viz", func(t *testing.T) {
		ofp, err := os.Create("/tmp/bst.dot")
		if err != nil {
			t.Error("cannot create file")
		}
		defer ofp.Close()
		s.Viz(ofp)
	})
	ctx, cancel := context.WithCancel(context.Background())
	t.Run("Check", func(t *testing.T) {
		violations := 0
		for n := range s.Check(ctx) {
			t.Logf("violating node: %+v", *n)
			violations++
		}
		if violations != 0 {
			t.Errorf("check found %d violations", violations)
		}
	})
	t.Run("KeysAndCancel", func(t *testing.T) {
		want := 0
		for got := range s.Keys(ctx) {
			igot := int(got.(iKey))
			if igot != want {
				t.Errorf("bad key: got %d, want %d", igot, want)
			}
			// t.Logf("key check OK: %d", want)
			want++
			if want > 4 {
				cancel()
			}
		}
	})
	t.Run("Get", func(t *testing.T) {
		for i := len(kvs) - 1; i >= 0; i-- {
			p := kvs[i]
			n := s.Get(p.k)
			if n == nil {
				t.Errorf("missing key: %d", int(p.k))
			}
			if n.Value.(int) != p.v {
				t.Errorf("bad value at %d", int(p.k))
			}
		}
	})
	t.Run("Next", func(t *testing.T) {
		for i := 0; i < 6; i++ {
			b := s.Get(iKey(i))
			got := b.Next()
			if got == nil {
				t.Errorf("unexpected nil Next(%d)", i)
			}
			t.Logf("Next(%d) == %+v", i, *got)
			v := got.Value.(int)
			want := -(i + 1)
			if v != want {
				t.Errorf("bad Next(%d).Value: got %d, want %d", i, v, want)
			}
		}
	})
	t.Run("Delete", func(t *testing.T) {
		d := s.Get(iKey(3))
		d.Delete()
		v := 0
		for n := range s.Check(ctx) {
			t.Logf("violation after delete: %+v", *n)
			v++
		}
		if v > 0 {
			t.Errorf("delete violations: %d", v)
		}
	})
}
