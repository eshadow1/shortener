package model

import "encoding/json"

type OriginalInfo struct {
	OriginalURL   string `json:"url"`
	CorrelationID string `json:"correlation_id,omitempty"`
}

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

type ShortenInfo struct {
	ShortURL      string `json:"short_url"`
	CorrelationID string `json:"correlation_id,omitempty"`
}

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
