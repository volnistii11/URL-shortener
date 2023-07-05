package model

type URL struct {
	ID            uint   `json:"uuid,string,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
	ShortURL      string `json:"short_url,omitempty"`
	OriginalURL   string `json:"original_url,omitempty"`
	UserID        int    `json:"user_id,omitempty"`
}

type URLS []*URL
