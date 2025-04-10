package lockctx

import "fmt"

// dirgraph is a directed graph for use with dagPolicy.
// Conceptually it is a partial directed graph representation, in that it only represents edge relationships.
// There is no notion of node existence in dirgraph, only the existence of edges.
type dirgraph struct {
	// Maps node to a set of neighbours.
	edges map[string]map[string]struct{}
}

func newDirectedGraph() dirgraph {
	return dirgraph{
		edges: make(map[string]map[string]struct{}),
	}
}

// neighbours returns the neighbours of a node.
// A neighbour relationship is directional. If only the edge a->b exists, then
// neighbours(a) will include b, but neighbours(b) will not include a.
func (d dirgraph) neighbours(node string) map[string]struct{} {
	edges := d.edges[node]
	if edges == nil {
		edges = make(map[string]struct{})
		d.edges[node] = edges
	}
	return edges
}

// addEdge adds an edge between node1 and node2. Edges are directional, so this will
// add node2 to node1's neighbour set, but will not add node1 to node2's neighbour set.
// Self-connections are allowed and will be detected as a cycle.
// This function is idempotent.
func (d dirgraph) addEdge(node1, node2 string) {
	edges := d.neighbours(node1)
	edges[node2] = struct{}{}
}

// hasCycle searches for cycles in the graph.
// If one or more cycles exists, one of the cycles is returned at random.
// If no cycle exists, returns nil, false.
// Returned cycles are in the form of a list of nodes, where subsequent nodes are
// connected in the graph, and the last element is connected to the first element.
func (d dirgraph) hasCycle() ([]string, bool) {
	visited := make(map[string]struct{}, len(d.edges))
	path := make([]string, 0, len(d.edges))
	for node := range d.edges {
		if cycle, ok := d.dfsCycle(node, visited, path); ok {
			return toMinimalCycle(cycle), true
		}
	}
	return nil, false
}

// toMinimalCycle converts a cycle path found by DFS to the minimal cycle,
// removing nodes which are not part of the cycle path.
func toMinimalCycle(cycle []string) []string {
	last := cycle[len(cycle)-1]
	for i, node := range cycle {
		if node == last {
			return cycle[i+1:]
		}
	}
	// it should be impossible to reach this point if the input is an output of dfsCycle.
	panic("unreachable code")
}

// dfsCycle is the recursive step of hasCycle.
func (d dirgraph) dfsCycle(node string, visited map[string]struct{}, path []string) ([]string, bool) {
	if _, ok := visited[node]; ok {
		return nil, false
	}
	visited[node] = struct{}{}
	path = append(path, node)

	for neighbour := range d.neighbours(node) {
		for _, pathNode := range path {
			if pathNode == neighbour {
				cycle := append(path, neighbour)
				return cycle, true
			}
		}
		if cycle, ok := d.dfsCycle(neighbour, visited, path); ok {
			return cycle, true
		}
	}

	return nil, false
}

type DAGPolicyBuilder struct {
	dag dirgraph
}

func (b DAGPolicyBuilder) Add(lock1, lock2 string) DAGPolicyBuilder {
	b.dag.addEdge(lock1, lock2)
	return b
}

func (b DAGPolicyBuilder) Build() Policy {
	if cycle, ok := b.dag.hasCycle(); ok {
		panic(fmt.Sprintf("invalid DAG policy contains cycle: %v", cycle))
	}
	return dagPolicy{
		dag: b.dag,
	}
}

func NewDAGPolicyBuilder() DAGPolicyBuilder {
	return DAGPolicyBuilder{
		dag: newDirectedGraph(),
	}
}

type dagPolicy struct {
	dag dirgraph
}

func (policy dagPolicy) CanAcquire(holding []string, next string) bool {
	if len(holding) == 0 {
		return true
	}
	last := holding[len(holding)-1]
	if _, ok := policy.dag.neighbours(last)[next]; ok {
		return true
	}
	return false
}
