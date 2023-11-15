package handler

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tiunovvv/go-yandex-shortener/internal/shortener"
)

type Handler struct {
	shortener *shortener.Shortener
}

func NewHandler(shortener *shortener.Shortener) *Handler {
	return &Handler{
		shortener: shortener,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	router.POST("/", h.PostHandler)
	router.GET("/:id", h.GetHandler)
	return router
}

func (h *Handler) PostHandler(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		newErrorResponce(c, http.StatusBadRequest, err.Error())
		return
	}

	if len(body) == 0 {
		newErrorResponce(c, http.StatusBadRequest, "body is initial")
		return
	}

	fullURL := string(body)

	if _, err := url.ParseRequestURI(fullURL); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Printf("%s is not URL", fullURL)
		return
	}

	shortURL := h.shortener.GetShortURL(fullURL, c.Request.URL.RequestURI())
	c.Status(http.StatusCreated)

	if _, err := c.Writer.Write([]byte(shortURL)); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log.Printf("Cant write %s into body", shortURL)
	}
}

func (h *Handler) GetHandler(c *gin.Context) {
	path := c.Request.URL.RequestURI()
	shortURL := strings.ReplaceAll(path, "/", "")

	if shortURL == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, "is not a form baseURL/shortURL")
		return
	}

	fullURL, err := h.shortener.GetFullURL(shortURL)

	if err != nil {
		newErrorResponce(c, http.StatusBadRequest, err.Error())
		return
	}

	c.Writer.Header().Set("Location", fullURL)
	c.Status(http.StatusTemporaryRedirect)
}
