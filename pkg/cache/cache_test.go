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

	assert.Contains(t, c["mod1"].Parents, c["df1"])

	assert.Equal(t, "[df1]", fmt.Sprint(c["mod1"].Parents))
	assert.Equal(t, "[df1 df2]", fmt.Sprint(c["mod2"].Parents))
	assert.Equal(t, "[df2 mod2 mod1]", fmt.Sprint(c["mod3"].Parents))
	assert.Equal(t, "[mod1 mod2]", fmt.Sprint(c["df1"].Children))
	assert.Equal(t, "[mod2 mod3]", fmt.Sprint(c["df2"].Children))
	assert.Equal(t, "[mod4]", fmt.Sprint(c["mod3"].Children))
	assert.Equal(t, "[mod3]", fmt.Sprint(c["mod4"].Parents))
	assert.Equal(t, "[mod5 mod6]", fmt.Sprint(c["mod4"].Children))
	assert.Equal(t, "[mod4]", fmt.Sprint(c["mod5"].Parents))
	assert.Equal(t, "[mod4]", fmt.Sprint(c["mod6"].Parents))

	n := c["mod3"]
	fmt.Println(c.UpstreamNodes(n))

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
