package cache

// CacheStore exposes an interface for new stores to implement
type CacheStore interface {
	SetDeps(parent string, deps ...string)
	UpstreamURLs(url string) (upstreams, roots []string)
	Dump()
}

// C is the in memory cache (depencency graph) for dinghyfiles and modules
var C CacheStore
