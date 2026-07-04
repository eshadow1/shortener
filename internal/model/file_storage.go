package model

// FileStorage представляет модель хранения данных об URL в хранилище,
// содержащую уникальные идентификаторы, ссылки и метаданные пользователя.
type FileStorage struct {
	// UUID содержит уникальный идентификатор записи.
	UUID string `json:"uuid"`
	// Short содержит короткую версию URL.
	Short string `json:"short_url"`
	// Original содержит оригинальный (длинный) URL.
	Original string `json:"original_url"`
	// UserID содержит идентификатор пользователя, создавшего запись (может быть пустым).
	UserID string `json:"user_id,omitempty"`
	// IsDelete указывает, была ли запись удалена.
	IsDelete bool `json:"is_delete,omitempty"`
}

// MemoryStorage представляет модель хранения данных об URL в памяти.
type MemoryStorage struct {
	// Original содержит оригинальный (длинный) URL.
	Original string `json:"original_url"`
	// IsDelete указывает, была ли запись удалена.
	IsDelete bool `json:"is_delete"`
}
