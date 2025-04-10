package graph

import (
	"slices"
	"testing"

	"github.com/jordanschalm/lockctx/internal/assert"
)

func TestDAGConstructor(t *testing.T) {
	dag := NewGraph()
	assert.True(t, len(dag.edges) == 0)
	_, ok := dag.HasCycle()
	assert.False(t, ok)
}

func TestDAGAddEdge(t *testing.T) {
	t.Run("should have 0 Neighbours initially", func(t *testing.T) {
		dag := NewGraph()
		assert.True(t, len(dag.Neighbours("a")) == 0)
	})

	t.Run("edges should be directed", func(t *testing.T) {
		dag := NewGraph()
		dag.AddEdge("a", "b")
		assert.True(t, len(dag.Neighbours("a")) == 1)
		assert.True(t, len(dag.Neighbours("b")) == 0)
	})

	t.Run("adding edges should be idempotent", func(t *testing.T) {
		dag := NewGraph()
		dag.AddEdge("a", "b")
		dag.AddEdge("a", "b")
		assert.True(t, len(dag.Neighbours("a")) == 1)
		assert.True(t, len(dag.Neighbours("b")) == 0)
	})

	t.Run("can create multiple Neighbours", func(t *testing.T) {
		dag := NewGraph()
		dag.AddEdge("a", "b")
		assert.True(t, len(dag.Neighbours("a")) == 1)
		assert.True(t, len(dag.Neighbours("b")) == 0)
		dag.AddEdge("a", "c")
		assert.True(t, len(dag.Neighbours("a")) == 2)
		assert.True(t, len(dag.Neighbours("b")) == 0)
		assert.True(t, len(dag.Neighbours("c")) == 0)
		dag.AddEdge("a", "d")
		assert.True(t, len(dag.Neighbours("a")) == 3)
		assert.True(t, len(dag.Neighbours("b")) == 0)
		assert.True(t, len(dag.Neighbours("c")) == 0)
		assert.True(t, len(dag.Neighbours("d")) == 0)
	})

	t.Run("can create a cycle", func(t *testing.T) {
		dag := NewGraph()
		dag.AddEdge("a", "b")
		dag.AddEdge("b", "a")
		assert.True(t, len(dag.Neighbours("a")) == 1)
		assert.True(t, len(dag.Neighbours("b")) == 1)
	})

	t.Run("self-connection", func(t *testing.T) {
		dag := NewGraph()
		dag.AddEdge("a", "a")
		assert.True(t, len(dag.Neighbours("a")) == 1)
	})

	t.Run("disconnected", func(t *testing.T) {
		dag := NewGraph()
		dag.AddEdge("a", "b")
		dag.AddEdge("b", "c")
		dag.AddEdge("c", "d")
		// no connection from e->f
		dag.AddEdge("e", "f")
		dag.AddEdge("f", "g")
		assert.True(t, len(dag.Neighbours("a")) == 1)
		assert.True(t, len(dag.Neighbours("b")) == 1)
		assert.True(t, len(dag.Neighbours("c")) == 1)
		assert.True(t, len(dag.Neighbours("d")) == 0)
		assert.True(t, len(dag.Neighbours("e")) == 1)
		assert.True(t, len(dag.Neighbours("f")) == 1)
		assert.True(t, len(dag.Neighbours("g")) == 0)
	})
}

func TestDAGHasCycle(t *testing.T) {
	t.Run("no cycle", func(t *testing.T) {
		dag := NewGraph()
		dag.AddEdge("a", "b")
		dag.AddEdge("b", "c")
		dag.AddEdge("c", "d")
		dag.AddEdge("d", "e")
		_, ok := dag.HasCycle()
		assert.False(t, ok)
	})

	t.Run("self-connection", func(t *testing.T) {
		dag := NewGraph()
		dag.AddEdge("a", "a")
		cycle, ok := dag.HasCycle()
		assert.True(t, ok)
		assert.True(t, slices.Equal([]string{"a"}, cycle))
	})

	t.Run("minimal cycle", func(t *testing.T) {
		dag := NewGraph()
		dag.AddEdge("a", "b")
		dag.AddEdge("b", "a")
		cycle, ok := dag.HasCycle()
		slices.Sort(cycle)
		assert.True(t, ok)
		assert.True(t, slices.Equal([]string{"a", "b"}, cycle))
	})

	t.Run("multiple cycles", func(t *testing.T) {
		dag := NewGraph()
		dag.AddEdge("a", "b")
		dag.AddEdge("b", "a") // cycle: a->b
		dag.AddEdge("c", "d")
		dag.AddEdge("d", "b") // cycle: b->c->d

		cycle, ok := dag.HasCycle()
		slices.Sort(cycle)
		assert.True(t, ok)
		assert.True(t, slices.Equal([]string{"a", "b"}, cycle) || slices.Equal([]string{"b", "c", "d"}, cycle))
	})
}
