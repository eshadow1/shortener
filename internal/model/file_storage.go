package model

type FileStorage struct {
	UUID     string `json:"uuid"`
	Short    string `json:"short_url"`
	Original string `json:"original_url"`
	UserID   string `json:"user_id,omitempty"`
}
