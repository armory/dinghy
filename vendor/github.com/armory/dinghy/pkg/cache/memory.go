/*
* Copyright 2019 Armory, Inc.

* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at

*    http://www.apache.org/licenses/LICENSE-2.0

* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

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

// MemoryCache maintains a mapping of dinghyfiles and their dependencies
type MemoryCache map[string]*Node

// NewMemoryCache initializes a new cache
func NewMemoryCache() MemoryCache {
	return MemoryCache{}
}

func findInSlice(n *Node, slice []*Node) int {
	for i, x := range slice {
		if x.URL == n.URL {
			return i
		}
	}
	return -1
}

// SetDeps sets the dependencies for a parent
func (c MemoryCache) SetDeps(parent string, deps []string) {
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
func (c MemoryCache) UpstreamURLs(url string) ([]string, []string) {
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

// GetRoots returns all roots for a given leaf
func (c MemoryCache) GetRoots(url string) []string {
	_, roots := c.UpstreamURLs(url)
	return roots
}

// Dump prints the cache, used for debugging
func (c MemoryCache) Dump() {
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
