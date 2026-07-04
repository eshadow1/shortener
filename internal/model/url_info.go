package model

// URLInfo представляет агрегированную информацию о сокращённом URL,
// содержащую оригинальную ссылку, её короткий вариант и уникальный идентификатор.
type URLInfo struct {
	// OriginalURL содержит оригинальный (длинный) URL.
	OriginalURL string `json:"original_url"`
	// ShortURL содержит соответствующий короткий URL.
	ShortURL string `json:"shorten_url"`
	// ID содержит уникальный идентификатор записи.
	ID string `json:"id"`
}
