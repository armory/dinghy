package cache

import (
	log "github.com/sirupsen/logrus"
)

// Node represents either a dinghyfile or a module.
// The URL is the github URL for the dinghyfile or module that can optionally
// include the commit hash for versioning purposes.
type Node struct {
	URL      string
	Children []*Node
	Parents  []*Node
}

func (n *Node) String() string {
	return n.URL
}

// NewNode allocates a new node in the cache
func NewNode(url string) *Node {
	return &Node{
		URL:      url,
		Children: make([]*Node, 0),
		Parents:  make([]*Node, 0),
	}
}

// MemoryCacheStore maintains a mapping of dinghyfiles and their dependencies
type MemoryCacheStore map[string]*Node

// NewMemoryCacheStore initializes a new cache
func NewMemoryCacheStore() MemoryCacheStore {
	return MemoryCacheStore{}
}

func findInSlice(n *Node, slice []*Node) int {
	for i, x := range slice {
		if x.URL == n.URL {
			return i
		}
	}
	return -1
}

func (c MemoryCacheStore) SetDeps(parent string, deps ...string) {
	if _, exists := c[parent]; !exists {
		c[parent] = NewNode(parent)
	}

	node := c[parent]
	found := make(map[string]*Node, 0)
	removed := make(map[string]*Node, 0)

	for _, child := range node.Children {
		removed[child.URL] = child
	}

	for _, depURL := range deps {
		if depNode, exists := removed[depURL]; exists {
			found[depURL] = depNode
			delete(removed, depURL)
		} else {
			if _, exists := c[depURL]; !exists {
				c[depURL] = NewNode(depURL)
			}
			found[depURL] = c[depURL]
		}
	}

	for _, removedNode := range removed {
		if i := findInSlice(node, removedNode.Parents); i != -1 {
			removedNode.Parents = append(removedNode.Parents[:i], removedNode.Parents[i+1:]...)
		}

		if i := findInSlice(removedNode, node.Children); i != -1 {
			node.Children = append(node.Children[:i], node.Children[i+1:]...)
		}
	}

	for _, foundNode := range found {
		if findInSlice(foundNode, node.Children) == -1 {
			node.Children = append(node.Children, foundNode)
		}

		if findInSlice(node, foundNode.Parents) == -1 {
			foundNode.Parents = append(foundNode.Parents, node)
		}
	}
}

// UpstreamURLs returns two arrays:
// 1) Array of all upstream URLs from a URL
// 2) Array of only the root URLs (dinghyfiles) for a given URL
func (c MemoryCacheStore) UpstreamURLs(url string) ([]string, []string) {
	n, exists := c[url]
	if !exists {
		return nil, nil
	}

	upstreams := make([]*Node, 0)
	roots := make([]*Node, 0)

	visited := map[*Node]bool{}
	q := make(chan *Node, len(c))
	q <- n
	for len(q) > 0 {
		curr := <-q
		visited[curr] = true

		// don't add self to the list of upstreams or roots
		if curr != n {
			upstreams = append(upstreams, curr)
			if len(curr.Parents) == 0 {
				roots = append(roots, curr)
			}
		}

		// enqueue the upstream nodes if not already visited
		for _, parent := range curr.Parents {
			if _, exists := visited[parent]; !exists {
				q <- parent
				visited[parent] = true
			}
		}
	}

	upstreamURLs := make([]string, 0)
	for _, node := range upstreams {
		upstreamURLs = append(upstreamURLs, node.URL)
	}

	rootURLs := make([]string, 0)
	for _, node := range roots {
		rootURLs = append(rootURLs, node.URL)
	}

	return upstreamURLs, rootURLs
}

// Dump prints the cache, used for debugging
func (c MemoryCacheStore) Dump() {
	for k, v := range c {
		log.Debug("-----------")
		log.Debug(k)
		log.Debug("Parents:")
		for _, p := range v.Parents {
			log.Debug("  " + p.URL)
		}
		log.Debug("Children:")
		for _, c := range v.Children {
			log.Debug("  " + c.URL)
		}
		log.Debug("-----------")
	}
}
