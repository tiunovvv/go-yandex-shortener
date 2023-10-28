package handler

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/tiunovvv/go-yandex-shortener/pkg/shortener"
)

const (
	schemeHttp    = "http://"
	schemeHttps   = "https://"
	bodyIsInitial = "Body is initial"
)

type Handler struct {
	shorteners *shortener.URLShortener
}

func NewHandler(shorteners *shortener.URLShortener) *Handler {
	return &Handler{shorteners: shorteners}
}

func (h *Handler) PostHandler(res http.ResponseWriter, req *http.Request) {

	body, err := io.ReadAll(req.Body)

	if len(body) == 0 {
		http.Error(res, bodyIsInitial, http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	bodyUrl, err := url.ParseRequestURI(string(body))
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	shortUrl, err := shortener.AppendToMap(h.shorteners, bodyUrl)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	scheme := schemeHttp
	if req.TLS != nil {
		scheme = schemeHttps
	}

	url := scheme + req.Host + req.URL.RequestURI() + string(shortUrl)

	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(url))
}

func (h *Handler) GetHandler(res http.ResponseWriter, req *http.Request) {
	path := req.URL.RequestURI()
	shortUrl := strings.Replace(path, "/", "", -1)

	fullUrl, _ := shortener.GetFullUrl(h.shorteners, shortUrl)

	// if err != nil {
	// 	http.Error(res, err.Error(), http.StatusBadRequest)
	// 	return
	// }

	res.Header().Set("Location", string(fullUrl))
	res.WriteHeader(http.StatusTemporaryRedirect)
}
