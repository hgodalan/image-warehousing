package models

// SearchRequest represents a search query from the user
type SearchRequest struct {
	Query  string `json:"query"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// SearchResult represents a single search result with relevance score
type SearchResult struct {
	ImageID        string  `json:"image_id"`
	RelevanceScore float64 `json:"relevance_score"`
	Reason         string  `json:"reason,omitempty"`
	Image          *Image  `json:"image,omitempty"`
}

// SearchResponse represents the complete search results
type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
	Query   string         `json:"query"`
}
