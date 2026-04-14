package model

type FileStorage struct {
	UUID     int    `json:"uuid"`
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
}
