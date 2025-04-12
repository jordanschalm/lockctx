package lockctx

import (
	"fmt"

	"github.com/jordanschalm/lockctx/internal/graph"
)

// statelessPolicy is a type to use for a stateless Policy defined as a function.
// This just makes the Policy function satisfy the Policy interface.
type statelessPolicy func([]string, string) bool

// CanAcquire calls the receiver function.
func (policy statelessPolicy) CanAcquire(holding []string, next string) bool {
	return policy(holding, next)
}

// StringOrderPolicy enforces that locks are acquired in lexicographic sort order.
// This Policy guarantees deadlock-free operation.
var StringOrderPolicy statelessPolicy = func(holding []string, next string) bool {
	if len(holding) == 0 {
		return true
	}
	last := holding[len(holding)-1]
	// next lock ID must sort after last acquired lock
	return last < next
}

// NoPolicy enforces no constraints on lock ordering.
var NoPolicy statelessPolicy = func(holding []string, next string) bool {
	return true
}

// DAGPolicyBuilder is used to construct a DAG policy.
// A DAG policy uses a directed acyclic graph, where graph nodes are lock IDs,
// to define when locks may be acquired. If an edge exists from A->B, then
// if I have most recently acquired lock A, I am allowed to acquire lock B next.
// Edges are added at construction time with Add. When all edges have been defined,
// the Policy can be created with Build.
// A DAG Policy returned from Build guarantees deadlock-free operation.
type DAGPolicyBuilder struct {
	dag graph.Graph
}

// NewDAGPolicyBuilder returns a DAGPolicyBuilder with an empty graph.
func NewDAGPolicyBuilder() DAGPolicyBuilder {
	return DAGPolicyBuilder{
		dag: graph.NewGraph(),
	}
}

// Add defines the Policy by adding a lock acquisition allowance (an edge in the graph).
// The resulting Policy will allow threads to acquire lock2 if they have just acquired lock1.
func (b DAGPolicyBuilder) Add(lock1, lock2 string) DAGPolicyBuilder {
	b.dag.AddEdge(lock1, lock2)
	return b
}

// Build validates that the constructed graph is acyclic.
// If the constructed graph is cyclic, this function will panic. DAGPolicyBuilder (and policies in general)
// are intended to be called at startup with statically defined parameters, hence the use of panic here.
// If the constructed graph is acyclic, a dagPolicy using the constructed graph is returned.
func (b DAGPolicyBuilder) Build() Policy {
	if cycle, ok := b.dag.HasCycle(); ok {
		panic(fmt.Sprintf("invalid DAG policy contains cycle: %v", cycle))
	}
	return dagPolicy{
		dag: b.dag,
	}
}

type dagPolicy struct {
	dag graph.Graph
}

// CanAcquire returns true if the caller is allowed to acquire the next lock N.
// Let L be the lock the caller most recently acquired (last element in holding).
// The caller can acquire N if N is there exists an edge L->N in the DAG.
func (policy dagPolicy) CanAcquire(holding []string, next string) bool {
	if len(holding) == 0 {
		return true
	}
	last := holding[len(holding)-1]
	if _, ok := policy.dag.Neighbours(last)[next]; ok {
		return true
	}
	return false
}
