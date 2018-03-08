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

// C is the in memory cache (depencency graph) for dinghyfiles and modules
var C Cache

// NewCache initialize a new cache
func NewCache() Cache {
	return map[string]*Node{}
}

// Add adds a node to the cache and updates its links
func (c Cache) Add(url string, depURLs ...string) {
	log.Debug("Adding " + url + " to cache")
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
		if !inSlice(n, depNode.Parents) {
			depNode.Parents = append(depNode.Parents, n)
		}

		// update children of parent
		if !inSlice(depNode, n.Children) {
			n.Children = append(n.Children, depNode)
		}
	}
}

func inSlice(n *Node, slice []*Node) bool {
	for _, i := range slice {
		if i == n {
			return true
		}
	}
	return false
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

// DumpCache prints the cache, used for debugging
func (c Cache) DumpCache() {
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
