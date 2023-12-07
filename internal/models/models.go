package models

type ReqAPI struct {
	URL string `json:"url"`
}

type ResAPI struct {
	Result string `json:"result"`
}

type ReqAPIBatch struct {
	ID      string `json:"correlation_id"`
	FullURL string `json:"original_url"`
}

type ResAPIBatch struct {
	ID       string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}
