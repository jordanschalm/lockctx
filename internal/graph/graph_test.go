package graph

import (
	"slices"
	"testing"

	"github.com/jordanschalm/lockctx/internal/assert"
)

func TestGraphConstructor(t *testing.T) {
	graph := NewGraph()
	assert.True(t, len(graph.edges) == 0)
	_, ok := graph.HasCycle()
	assert.False(t, ok)
}

func TestGraphAddEdge(t *testing.T) {
	t.Run("should have 0 Neighbours initially", func(t *testing.T) {
		graph := NewGraph()
		assert.True(t, len(graph.Neighbours("a")) == 0)
	})

	t.Run("edges should be directed", func(t *testing.T) {
		graph := NewGraph()
		graph.AddEdge("a", "b")
		assert.True(t, len(graph.Neighbours("a")) == 1)
		assert.True(t, len(graph.Neighbours("b")) == 0)
	})

	t.Run("adding edges should be idempotent", func(t *testing.T) {
		graph := NewGraph()
		graph.AddEdge("a", "b")
		graph.AddEdge("a", "b")
		assert.True(t, len(graph.Neighbours("a")) == 1)
		assert.True(t, len(graph.Neighbours("b")) == 0)
	})

	t.Run("can create multiple Neighbours", func(t *testing.T) {
		graph := NewGraph()
		graph.AddEdge("a", "b")
		assert.True(t, len(graph.Neighbours("a")) == 1)
		assert.True(t, len(graph.Neighbours("b")) == 0)
		graph.AddEdge("a", "c")
		assert.True(t, len(graph.Neighbours("a")) == 2)
		assert.True(t, len(graph.Neighbours("b")) == 0)
		assert.True(t, len(graph.Neighbours("c")) == 0)
		graph.AddEdge("a", "d")
		assert.True(t, len(graph.Neighbours("a")) == 3)
		assert.True(t, len(graph.Neighbours("b")) == 0)
		assert.True(t, len(graph.Neighbours("c")) == 0)
		assert.True(t, len(graph.Neighbours("d")) == 0)
	})

	t.Run("can create a cycle", func(t *testing.T) {
		graph := NewGraph()
		graph.AddEdge("a", "b")
		graph.AddEdge("b", "a")
		assert.True(t, len(graph.Neighbours("a")) == 1)
		assert.True(t, len(graph.Neighbours("b")) == 1)
	})

	t.Run("self-connection", func(t *testing.T) {
		graph := NewGraph()
		graph.AddEdge("a", "a")
		assert.True(t, len(graph.Neighbours("a")) == 1)
	})

	t.Run("disconnected", func(t *testing.T) {
		graph := NewGraph()
		graph.AddEdge("a", "b")
		graph.AddEdge("b", "c")
		graph.AddEdge("c", "d")
		// no connection from e->f
		graph.AddEdge("e", "f")
		graph.AddEdge("f", "g")
		assert.True(t, len(graph.Neighbours("a")) == 1)
		assert.True(t, len(graph.Neighbours("b")) == 1)
		assert.True(t, len(graph.Neighbours("c")) == 1)
		assert.True(t, len(graph.Neighbours("d")) == 0)
		assert.True(t, len(graph.Neighbours("e")) == 1)
		assert.True(t, len(graph.Neighbours("f")) == 1)
		assert.True(t, len(graph.Neighbours("g")) == 0)
	})
}

func TestGraphHasCycle(t *testing.T) {
	t.Run("no cycle", func(t *testing.T) {
		graph := NewGraph()
		graph.AddEdge("a", "b")
		graph.AddEdge("b", "c")
		graph.AddEdge("c", "d")
		graph.AddEdge("d", "e")
		_, ok := graph.HasCycle()
		assert.False(t, ok)
	})

	t.Run("self-connection", func(t *testing.T) {
		graph := NewGraph()
		graph.AddEdge("a", "a")
		cycle, ok := graph.HasCycle()
		assert.True(t, ok)
		assert.True(t, slices.Equal([]string{"a"}, cycle))
	})

	t.Run("minimal cycle", func(t *testing.T) {
		graph := NewGraph()
		graph.AddEdge("a", "b")
		graph.AddEdge("b", "a")
		cycle, ok := graph.HasCycle()
		slices.Sort(cycle)
		assert.True(t, ok)
		assert.True(t, slices.Equal([]string{"a", "b"}, cycle))
	})

	t.Run("multiple cycles", func(t *testing.T) {
		graph := NewGraph()
		graph.AddEdge("a", "b")
		graph.AddEdge("b", "a") // cycle: a->b
		graph.AddEdge("c", "d")
		graph.AddEdge("d", "b") // cycle: b->c->d

		cycle, ok := graph.HasCycle()
		slices.Sort(cycle)
		assert.True(t, ok)
		assert.True(t, slices.Equal([]string{"a", "b"}, cycle) || slices.Equal([]string{"b", "c", "d"}, cycle))
	})
}
