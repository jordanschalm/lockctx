package graph

// Graph is a directed graph for use with the DAG policy.
// Graph only represents edge relationships. There is no notion of node existence.
type Graph struct {
	// Maps node to a set of neighbours.
	edges map[string]map[string]struct{}
}

func NewGraph() Graph {
	return Graph{
		edges: make(map[string]map[string]struct{}),
	}
}

// Neighbours returns the neighbours of a node.
// A neighbour relationship is directional. If only the edge a->b exists, then
// Neighbours(a) will include b, but Neighbours(b) will not include a.
func (d Graph) Neighbours(node string) map[string]struct{} {
	edges := d.edges[node]
	if edges == nil {
		edges = make(map[string]struct{})
		d.edges[node] = edges
	}
	return edges
}

// AddEdge adds an edge between node1 and node2. Edges are directional, so this will
// add node2 to node1's neighbour set, but will not add node1 to node2's neighbour set.
// Self-connections are allowed and will be detected as a cycle.
// This function is idempotent.
func (d Graph) AddEdge(node1, node2 string) {
	edges := d.Neighbours(node1)
	edges[node2] = struct{}{}
}

// HasCycle searches for cycles in the graph.
// If one or more cycles exists, one of the cycles is returned at random.
// If no cycle exists, returns nil, false.
// Returned cycles are in the form of a list of nodes, where subsequent nodes are
// connected in the graph, and the last element is connected to the first element.
func (d Graph) HasCycle() ([]string, bool) {
	visited := make(map[string]struct{}, len(d.edges))
	path := make([]string, 0, len(d.edges))
	for node := range d.edges {
		if cycle, ok := d.dfsCycle(node, visited, path); ok {
			return toMinimalCycle(cycle), true
		}
	}
	return nil, false
}

// dfsCycle is the recursive step of HasCycle.
func (d Graph) dfsCycle(node string, visited map[string]struct{}, path []string) ([]string, bool) {
	if _, ok := visited[node]; ok {
		return nil, false
	}
	visited[node] = struct{}{}
	path = append(path, node)

	for neighbour := range d.Neighbours(node) {
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
