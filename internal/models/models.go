package models

// ReqAPI struct for "/api/shorten" handler request.
type ReqAPI struct {
	URL string `json:"url"`
}

// ResAPI struct for "/api/shorten" handler response.
type ResAPI struct {
	Result string `json:"result"`
}

// ReqAPIBatch struct for "/api/shorten/batch" handler request.
type ReqAPIBatch struct {
	ID      string `json:"correlation_id"`
	FullURL string `json:"original_url"`
}

// ResAPIBatch struct for "/api/shorten/batch" handler response.
type ResAPIBatch struct {
	ID       string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}

// UsersURLs struct for "/api/user/urls" handler response.
type UsersURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
