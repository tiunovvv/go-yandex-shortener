package handler

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tiunovvv/go-yandex-shortener/pkg/storage"
)

const (
	schemeHTTP    = "http://"
	schemeHTTPS   = "https://"
	bodyIsInitial = "Body is initial"
	idIsInitial   = "Id is initial"
	Location      = "Location"
)

type Handler struct {
	storage      *storage.URLShortener
	shortURLBase string
}

func NewHandler(storage *storage.URLShortener, shortURLBase string) *Handler {
	return &Handler{
		storage:      storage,
		shortURLBase: shortURLBase}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	router.POST("/", h.PostHandler)
	router.GET("/:id", h.GetHandler)
	return router
}

func (h *Handler) PostHandler(c *gin.Context) {

	body, err := io.ReadAll(c.Request.Body)

	if len(body) == 0 {
		newErrorResponce(c, http.StatusBadRequest, bodyIsInitial)
		return
	}

	if err != nil {
		newErrorResponce(c, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := url.ParseRequestURI(string(body)); err != nil {
		newErrorResponce(c, http.StatusBadRequest, err.Error())
		return
	}

	shortURL, err := storage.AppendToMap(h.storage, body)

	if err != nil {
		newErrorResponce(c, http.StatusBadRequest, err.Error())
		return
	}

	url := h.shortURLBase + string(shortURL)

	c.Status(http.StatusCreated)
	c.Writer.Write([]byte(url))
}

func (h *Handler) GetHandler(c *gin.Context) {

	path := c.Request.URL.RequestURI()
	shortURL := strings.Replace(path, "/", "", -1)

	if shortURL == "" {
		newErrorResponce(c, http.StatusBadRequest, idIsInitial)
		return
	}

	fullURL, err := storage.GetFullURL(h.storage, shortURL)

	if err != nil {
		newErrorResponce(c, http.StatusBadRequest, err.Error())
		return
	}

	c.Writer.Header().Set(Location, string(fullURL))
	c.Status(http.StatusTemporaryRedirect)
}
