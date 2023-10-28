package handler

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/tiunovvv/go-yandex-shortener/pkg/shortener"
)

const (
	schemeHTTP    = `http://`
	schemeHTTPS   = `https://`
	bodyIsInitial = `Body is initial`
	Location      = `Location`
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

	bodyURL, err := url.ParseRequestURI(string(body))
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL, err := shortener.AppendToMap(h.shorteners, bodyURL)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	scheme := schemeHTTP
	if req.TLS != nil {
		scheme = schemeHTTPS
	}

	url := scheme + req.Host + req.URL.RequestURI() + string(shortURL)

	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(url))
}

func (h *Handler) GetHandler(res http.ResponseWriter, req *http.Request) {
	path := req.URL.RequestURI()
	shortURL := strings.Replace(path, `/`, ``, -1)

	fullURL, err := shortener.GetFullUrl(h.shorteners, shortURL)

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.Header().Set(Location, string(fullURL))
	res.WriteHeader(http.StatusTemporaryRedirect)
}
