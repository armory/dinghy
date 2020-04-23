package git


type GitInfo struct {
	RawData                 map[string]interface{}
	Org, Repo, Path, Branch string
}