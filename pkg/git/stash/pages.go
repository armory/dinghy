package stash

// PagedAPIResponse contains the fields needed for paged Stash APIs, described here:
// https://docs.atlassian.com/bitbucket-server/rest/5.8.0/bitbucket-rest.html#paging-params
type PagedAPIResponse struct {
	IsLastPage    bool `json:"isLastPage"`
	Start         int  `json:"start"`
	NextPageStart int  `json:"nextPageStart"`
}
