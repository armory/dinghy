package cache

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRoots(t *testing.T) {
	dinghyfiles := map[string][]string{
		"df1": []string{"mod1", "mod2"},
		"df2": []string{"mod2", "mod3"},
	}
	modules := map[string][]string{
		"mod1": []string{"mod3"},
		"mod2": []string{"mod3"},
		"mod3": []string{"mod4"},
		"mod4": []string{"mod5", "mod6"},
	}
	c := NewCache()
	for dinghyfile, deps := range dinghyfiles {
		c.Add(dinghyfile, deps...)
	}
	for module, deps := range modules {
		c.Add(module, deps...)
	}
	n := c["mod5"]
	roots := c.RootsOf(n)
	fmt.Println(roots)
	expected1 := c["df1"]
	expected2 := c["df2"]
	assert.Contains(t, roots, expected1)
	assert.Contains(t, roots, expected2)

}

func TestCircularDep(t *testing.T) {

}

/*
	q := make(chan *Node, 100) // 100 is a magic number
	for dinghyfile, deps := range dinghyfiles {
		n := NewNode(dinghyfile)
		for _, dep := range deps {
			d := NewNode(dep)
			n.Children = append(n.Children, d)
			q <- d
		}
		cache.Add(n)
	}
	for len(q) > 0 {
		mod := <-q
		modDeps, ok := modules[mod.URL]
		if ok {
			for _, dep := range modDeps {
				d := NewNode(dep)
				mod.Children = append(mod.Children, d)
				q <- d
			}
			cache.Add(mod)
		}
	}
*/
