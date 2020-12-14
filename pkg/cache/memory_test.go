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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdempotency(t *testing.T) {
	c := NewMemoryCache()

	c.SetDeps("foo", []string{"c1", "c2"})
	assert.ElementsMatchf(t, c["foo"].Children, []*Node{c["c1"], c["c2"]}, "foo should have c1 and c2 as children")
	assert.ElementsMatchf(t, c["foo"].Parents, []*Node{}, "foo should not have any parents")

	c.SetDeps("foo", []string{"c2", "c3"})
	assert.ElementsMatchf(t, c["foo"].Children, []*Node{c["c2"], c["c3"]}, "foo should have c2 and c3 as children, and not c1")
	assert.ElementsMatchf(t, c["foo"].Parents, []*Node{}, "foo should not have any parents")

	c.SetDeps("foo", []string{})
	assert.ElementsMatchf(t, c["foo"].Children, []*Node{}, "foo should not have any children")
	assert.ElementsMatchf(t, c["foo"].Parents, []*Node{}, "foo should not have any parents")
}

func TestCache(t *testing.T) {
	c := createCache()
	assert.ElementsMatchf(t, c["df1"].Parents, []*Node{}, "df1 should not have any parents")
	assert.ElementsMatchf(t, c["df2"].Parents, []*Node{}, "df2 should not have any parents")
	assert.ElementsMatchf(t, c["df1"].Children, []*Node{c["mod1"], c["mod2"]}, "df1's children list is wrong")
	assert.ElementsMatchf(t, c["df2"].Children, []*Node{c["mod2"], c["mod3"]}, "df2's children list is wrong")
	assert.ElementsMatchf(t, c["mod1"].Parents, []*Node{c["df1"]}, "mod1's parent list is wrong")
	assert.ElementsMatchf(t, c["mod2"].Parents, []*Node{c["df1"], c["df2"]}, "mod2's parent list is wrong")
	assert.ElementsMatchf(t, c["mod3"].Parents, []*Node{c["mod1"], c["mod2"], c["df2"]}, "mod3's parent list is wrong")
	assert.ElementsMatchf(t, c["mod4"].Parents, []*Node{c["mod3"]}, "mod4's parent list is wrong")
	assert.ElementsMatchf(t, c["mod5"].Parents, []*Node{c["mod4"]}, "mod5's parent list is wrong")
	assert.ElementsMatchf(t, c["mod6"].Parents, []*Node{c["mod4"]}, "mod6's parent list is wrong")
	assert.ElementsMatchf(t, c["mod3"].Children, []*Node{c["mod4"]}, "mod4's children list is wrong")
	assert.ElementsMatchf(t, c["mod4"].Children, []*Node{c["mod5"], c["mod6"]}, "mod6's children list is wrong")
	assert.ElementsMatchf(t, c["mod5"].Children, []*Node{}, "mod5 should not have any parents")
	assert.ElementsMatchf(t, c["mod6"].Children, []*Node{}, "mod6 should not have any parents")
}

func TestDeletingDependency(t *testing.T) {
	c := NewMemoryCache()

	c.SetDeps("df1", []string{"mod1", "mod2"})
	c.SetDeps("df2", []string{"mod2", "mod3"})

	_, roots := c.UpstreamURLs("mod2")
	assert.ElementsMatchf(t, roots, []string{"df1", "df2"}, "mod2 should have df1 and df2 as parents")

	// remove mod2 as a dependency from df2
	c.SetDeps("df2", []string{"mod3", "mod4"})

	_, roots = c.UpstreamURLs("mod2")
	assert.ElementsMatchf(t, roots, []string{"df1"}, "mod2 should only have one parent")
	assert.ElementsMatchf(t, c["df2"].Children, []*Node{c["mod3"], c["mod4"]}, "df2's children list is wrong")
}

func TestUpstreamsAndRoots(t *testing.T) {
	c := createCache()
	up, roots := c.UpstreamURLs("mod3")
	assert.ElementsMatchf(t, up, []string{"mod1", "mod2", "df1", "df2"}, "mod3's upstream nodes aren't quite right!")
	assert.ElementsMatchf(t, roots, []string{"df1", "df2"}, "mod3's root nodes aren't quite right!")

	up, roots = c.UpstreamURLs("mod6")
	assert.ElementsMatchf(t, up, []string{"mod1", "mod2", "mod3", "mod4", "df1", "df2"}, "mod6's upstream nodes aren't quite right!")
	assert.ElementsMatchf(t, roots, []string{"df1", "df2"}, "mod6's root nodes aren't quite right!")
}

/* The test dependency graph we are working with
   looks like this:

   df1    df2
    /\    /\
   /  \  /  \
mod1  mod2   \
   \      `\  |
    `------ mod3
             |
             |
            mod4
             /\
            /  \
         mod5  mod6
*/

func createCache() MemoryCache {
	c := NewMemoryCache()

	c.SetDeps("df1", []string{"mod1", "mod2"})
	c.SetDeps("df2", []string{"mod2", "mod3"})

	c.SetDeps("mod1", []string{"mod3"})
	c.SetDeps("mod2", []string{"mod3"})
	c.SetDeps("mod3", []string{"mod4"})
	c.SetDeps("mod4", []string{"mod5", "mod6"})

	return c
}
