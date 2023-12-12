package models


type Vacancy struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Company     string `json:"company"`
	Location    string `json:"location"`
	Description string `json:"description"`
}

type SearchHistory struct {
	ID        int    `json:"id"`
	Query     string `json:"query"`
	Timestamp string `json:"timestamp"`
}