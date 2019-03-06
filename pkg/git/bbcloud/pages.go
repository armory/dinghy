package bbcloud

// PagedAPIResponse contains the fields needed for paged APIs
type PagedAPIResponse struct {
	PageLength    int `json:"pagelen"`
	CurrentPage   int `json:"page"`
	NumberOfPages int `json:"size"`
}
