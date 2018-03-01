package cache

type Node struct {
	URL      string
	Children []*Node
	Parents  []*Node
}

func NewNode(url string) *Node {
	return &Node{
		URL:      url,
		Children: make([]*Node, 0),
		Parents:  make([]*Node, 0),
	}
}

type Cache map[string]*Node

func NewCache() Cache {
	return map[string]*Node{}
}

/*
func (c Cache) Add(n *Node) {
	c[n.URL] = n
	for _, child := range n.Children {
		p := child.Parents
		p = append(p, n)
		child.Parents = p
	}
}
*/

func (c Cache) UpstreamDinghyfile(url string) string {
	return ""
}

func (c Cache) Add(url string, depUrls ...string) {
	n := NewNode(url)
	c[url] = n
	for _, depUrl := range depUrls {
		if _, exists := c[depUrl]; !exists {
			depNode := NewNode(depUrl)
			c[depUrl] = depNode
		}
		depNode := c[depUrl]
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

func (c Cache) RootsOf(n *Node) []*Node {
	roots := make([]*Node, 0)
	if n == nil {
		// means n isnt in the cache
		return nil
	}
	vistedSet := map[*Node]struct{}{}
	q := make(chan *Node, len(c))
	q <- n
	for len(q) > 0 {
		curr := <-q
		vistedSet[curr] = struct{}{}
		for _, parent := range curr.Parents {
			if _, visted := vistedSet[parent]; !visted {
				q <- parent
			}
		}
		if len(curr.Parents) == 0 {
			roots = append(roots, curr)
		}
	}
	return roots
}

func (n *Node) String() string {
	return n.URL
}
