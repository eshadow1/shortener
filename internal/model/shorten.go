// Package model определяет основные модели данных, бизнес-ошибки и ключи контекста,
// используемые во всех слоях приложения (handler, service, repository).
package model

import "encoding/json"

// OriginalInfo представляет информацию об оригинальном URL,
// включая сам URL и опциональный идентификатор корреляции для пакетных запросов.
type OriginalInfo struct {
	// OriginalURL содержит оригинальный URL, который необходимо сократить.
	OriginalURL string `json:"url"`
	// CorrelationID содержит уникальный идентификатор,
	// используемый для связывания запроса и ответа в пакетной обработке.
	CorrelationID string `json:"correlation_id,omitempty"`
}

// UnmarshalJSON реализует кастомную десериализацию JSON для структуры OriginalInfo.
func (o *OriginalInfo) UnmarshalJSON(data []byte) error {
	var raw struct {
		OriginalURL   *string `json:"original_url"`
		URL           *string `json:"url"`
		CorrelationID *string `json:"correlation_id"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if raw.OriginalURL != nil {
		o.OriginalURL = *raw.OriginalURL
	} else if raw.URL != nil {
		o.OriginalURL = *raw.URL
	}

	if raw.CorrelationID != nil {
		o.CorrelationID = *raw.CorrelationID
	}

	return nil
}

// ShortenInfo представляет информацию о сокращенном URL,
// включая сам короткий URL и опциональный идентификатор корреляции.
type ShortenInfo struct {
	// ShortURL содержит сгенерированный короткий URL.
	ShortURL string `json:"short_url"`
	// CorrelationID содержит уникальный идентификатор,
	// используемый для связывания запроса и ответа в пакетной обработке.
	CorrelationID string `json:"correlation_id,omitempty"`
}

// UnmarshalJSON реализует кастомную десериализацию JSON для структуры ShortenInfo.
func (o *ShortenInfo) UnmarshalJSON(data []byte) error {
	var raw struct {
		ShortURL      *string `json:"short_url"`
		URL           *string `json:"result"`
		CorrelationID *string `json:"correlation_id"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if raw.ShortURL != nil {
		o.ShortURL = *raw.ShortURL
	} else if raw.URL != nil {
		o.ShortURL = *raw.URL
	}

	if raw.CorrelationID != nil {
		o.CorrelationID = *raw.CorrelationID
	}

	return nil
}

// UserURL представляет пару URL (оригинальный и короткий),
// принадлежащих одному пользователю, с указанием статуса удаления.
type UserURL struct {
	// OriginalURL содержит оригинальный URL.
	OriginalURL string `json:"original_url"`
	// ShortURL содержит соответствующий короткий URL.
	ShortURL string `json:"short_url"`
	// IsDeleted указывает, был ли данный URL удален пользователем.
	IsDeleted bool `json:"is_deleted"`
}

// DeleteInfo содержит информацию, необходимую для операции массового удаления URL
type DeleteInfo struct {
	// UserID содержит идентификатор пользователя, инициировавшего удаление.
	UserID string
	// URLs содержит список коротких URL, подлежащих удалению.
	URLs []string
}
