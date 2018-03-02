package cache

// Node represents either a dinghyfile or a module.
// The URL is the github URL for the dinghyfile or module that can optionally
// include the commit hash for versioning purposes.
type Node struct {
	URL      string
	Children []*Node
	Parents  []*Node
}

// NewNode allocates a new node in the cache
func NewNode(url string) *Node {
	return &Node{
		URL:      url,
		Children: make([]*Node, 0),
		Parents:  make([]*Node, 0),
	}
}

// Cache is the datastructure that maintains a mapping of dinghyfiles and their dependencies
type Cache map[string]*Node

// NewCache initialize a new cache
func NewCache() Cache {
	return map[string]*Node{}
}

// Add adds a node to the cache and updates its links
func (c Cache) Add(url string, depURLs ...string) {
	// check if it already exists in cache
	if _, exists := c[url]; !exists {
		c[url] = NewNode(url)
	}
	n := c[url]

	for _, depURL := range depURLs {
		if _, exists := c[depURL]; !exists {
			depNode := NewNode(depURL)
			c[depURL] = depNode
		}
		depNode := c[depURL]

		// update parents of child
		parents := depNode.Parents
		parents = append(parents, n)
		depNode.Parents = parents

		// update children of parent
		children := n.Children
		children = append(children, depNode)
		n.Children = children
	}
}

// UpstreamNodes returns two arrays:
// 1) Array of all upstream nodes from a node
// 2) Array of only the _Root_ nodes (dinghyfiles) for a given node
func (c Cache) UpstreamNodes(n *Node) ([]*Node, []*Node) {
	upstreams := make([]*Node, 0)
	roots := make([]*Node, 0)
	if n == nil {
		// means n isnt in the cache
		return nil, nil
	}
	vistedSet := map[*Node]bool{}
	q := make(chan *Node, len(c))
	q <- n
	for len(q) > 0 {
		curr := <-q
		vistedSet[curr] = true
		if curr != n {
			// don't add self to the list of upstreams or roots
			upstreams = append(upstreams, curr)
			if len(curr.Parents) == 0 {
				roots = append(roots, curr)
			}		
		}

		// enqueue the upstream nodes if not already visited
		for _, parent := range curr.Parents {
			if _, visted := vistedSet[parent]; !visted {
				q <- parent
				vistedSet[parent] = true
			}
		}
	}

	return upstreams, roots
}

func (n *Node) String() string {
	return n.URL
}
