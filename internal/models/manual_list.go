package models

// ManualListFilter — query-параметры GET /manuals
type ManualListFilter struct {
	TagID  *int
	Author string
	Q      string
	Sort   string // "" (по умолчанию) или "popular"
	Page   int
	Limit  int
}

// ManualListResult — пагинированный список
type ManualListResult struct {
	Items []Manual `json:"items"`
	Total int      `json:"total"`
	Page  int      `json:"page"`
	Limit int      `json:"limit"`
}
