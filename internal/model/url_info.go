package model

type URLInfo struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"shorten_url"`
	ID          string `json:"id"`
}
