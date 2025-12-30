package upload

type Response struct {
	TotalCount      int     `json:"total_count"`
	DuplicatesCount int     `json:"duplicates_count"`
	TotalItems      int     `json:"total_items"`
	TotalCategories int     `json:"total_categories"`
	TotalPrice      float64 `json:"total_price"`
}
