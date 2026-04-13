package model

type OriginalInfo struct {
	OriginalURL string `json:"url"`
}

type ShortenInfo struct {
	ShortURL string `json:"result"`
}
