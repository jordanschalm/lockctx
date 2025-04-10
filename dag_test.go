package lockctx

import (
	"slices"
	"testing"

	"github.com/jordanschalm/lockctx/internal"
)

func TestDAGConstructor(t *testing.T) {
	dag := newDirectedGraph()
	internal.AssertTrue(t, len(dag.edges) == 0)
	_, ok := dag.hasCycle()
	internal.AssertFalse(t, ok)
}

func TestDAGAddEdge(t *testing.T) {
	t.Run("should have 0 neighbours initially", func(t *testing.T) {
		dag := newDirectedGraph()
		internal.AssertTrue(t, len(dag.neighbours("a")) == 0)
	})

	t.Run("edges should be directed", func(t *testing.T) {
		dag := newDirectedGraph()
		dag.addEdge("a", "b")
		internal.AssertTrue(t, len(dag.neighbours("a")) == 1)
		internal.AssertTrue(t, len(dag.neighbours("b")) == 0)
	})

	t.Run("adding edges should be idempotent", func(t *testing.T) {
		dag := newDirectedGraph()
		dag.addEdge("a", "b")
		dag.addEdge("a", "b")
		internal.AssertTrue(t, len(dag.neighbours("a")) == 1)
		internal.AssertTrue(t, len(dag.neighbours("b")) == 0)
	})

	t.Run("can create multiple neighbours", func(t *testing.T) {
		dag := newDirectedGraph()
		dag.addEdge("a", "b")
		internal.AssertTrue(t, len(dag.neighbours("a")) == 1)
		internal.AssertTrue(t, len(dag.neighbours("b")) == 0)
		dag.addEdge("a", "c")
		internal.AssertTrue(t, len(dag.neighbours("a")) == 2)
		internal.AssertTrue(t, len(dag.neighbours("b")) == 0)
		internal.AssertTrue(t, len(dag.neighbours("c")) == 0)
		dag.addEdge("a", "d")
		internal.AssertTrue(t, len(dag.neighbours("a")) == 3)
		internal.AssertTrue(t, len(dag.neighbours("b")) == 0)
		internal.AssertTrue(t, len(dag.neighbours("c")) == 0)
		internal.AssertTrue(t, len(dag.neighbours("d")) == 0)
	})

	t.Run("can create a cycle", func(t *testing.T) {
		dag := newDirectedGraph()
		dag.addEdge("a", "b")
		dag.addEdge("b", "a")
		internal.AssertTrue(t, len(dag.neighbours("a")) == 1)
		internal.AssertTrue(t, len(dag.neighbours("b")) == 1)
	})

	t.Run("self-connection", func(t *testing.T) {
		dag := newDirectedGraph()
		dag.addEdge("a", "a")
		internal.AssertTrue(t, len(dag.neighbours("a")) == 1)
	})

	t.Run("disconnected", func(t *testing.T) {
		dag := newDirectedGraph()
		dag.addEdge("a", "b")
		dag.addEdge("b", "c")
		dag.addEdge("c", "d")
		// no connection from e->f
		dag.addEdge("e", "f")
		dag.addEdge("f", "g")
		internal.AssertTrue(t, len(dag.neighbours("a")) == 1)
		internal.AssertTrue(t, len(dag.neighbours("b")) == 1)
		internal.AssertTrue(t, len(dag.neighbours("c")) == 1)
		internal.AssertTrue(t, len(dag.neighbours("d")) == 0)
		internal.AssertTrue(t, len(dag.neighbours("e")) == 1)
		internal.AssertTrue(t, len(dag.neighbours("f")) == 1)
		internal.AssertTrue(t, len(dag.neighbours("g")) == 0)
	})
}

func TestDAGHasCycle(t *testing.T) {
	t.Run("no cycle", func(t *testing.T) {
		dag := newDirectedGraph()
		dag.addEdge("a", "b")
		dag.addEdge("b", "c")
		dag.addEdge("c", "d")
		dag.addEdge("d", "e")
		_, ok := dag.hasCycle()
		internal.AssertFalse(t, ok)
	})

	t.Run("self-connection", func(t *testing.T) {
		dag := newDirectedGraph()
		dag.addEdge("a", "a")
		cycle, ok := dag.hasCycle()
		internal.AssertTrue(t, ok)
		internal.AssertTrue(t, slices.Equal([]string{"a"}, cycle))
	})

	t.Run("minimal cycle", func(t *testing.T) {
		dag := newDirectedGraph()
		dag.addEdge("a", "b")
		dag.addEdge("b", "a")
		cycle, ok := dag.hasCycle()
		slices.Sort(cycle)
		internal.AssertTrue(t, ok)
		internal.AssertTrue(t, slices.Equal([]string{"a", "b"}, cycle))
	})

	t.Run("multiple cycles", func(t *testing.T) {
		dag := newDirectedGraph()
		dag.addEdge("a", "b")
		dag.addEdge("b", "a") // cycle: a->b
		dag.addEdge("c", "d")
		dag.addEdge("d", "b") // cycle: b->c->d

		cycle, ok := dag.hasCycle()
		slices.Sort(cycle)
		internal.AssertTrue(t, ok)
		internal.AssertTrue(t, slices.Equal([]string{"a", "b"}, cycle) || slices.Equal([]string{"b", "c", "d"}, cycle))
	})
}
