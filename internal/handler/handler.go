package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"github.com/tiunovvv/go-yandex-shortener/internal/middleware"
	"github.com/tiunovvv/go-yandex-shortener/internal/models"
	"github.com/tiunovvv/go-yandex-shortener/internal/shortener"
	"go.uber.org/zap"
)

type Handler struct {
	config    *config.Config
	shortener *shortener.Shortener
	logger    *zap.Logger
}

func NewHandler(config *config.Config, shortener *shortener.Shortener, logger *zap.Logger) *Handler {
	return &Handler{
		config:    config,
		shortener: shortener,
		logger:    logger,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	router.Use(middleware.GinGzip(h.logger))
	router.Use(middleware.GinLogger(h.logger))
	router.POST("/", h.PostHandler)
	router.POST("/api/shorten", h.PostAPIHandler)
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
		h.logger.Sugar().Errorf("%s is not URL", fullURL)
		return
	}

	shortURL := h.shortener.GetShortURL(fullURL)

	fullShortURL := h.config.BaseURL + c.Request.URL.RequestURI() + shortURL
	c.Status(http.StatusCreated)

	if _, err := c.Writer.Write([]byte(fullShortURL)); c.Request.Body == nil && err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		h.logger.Sugar().Errorf("Cant write %s into body", fullShortURL)
		return
	}
}

func (h *Handler) GetHandler(c *gin.Context) {
	path := c.Request.URL.RequestURI()
	shortURL := strings.ReplaceAll(path, "/", "")

	if shortURL == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, "is not a form baseURL/shortURL")
		return
	}

	fullURL, err := h.shortener.Storage.MemoryStore.GetFullURL(shortURL)

	if err != nil {
		newErrorResponce(c, http.StatusBadRequest, err.Error())
		return
	}

	c.Writer.Header().Set("Location", fullURL)
	c.Status(http.StatusTemporaryRedirect)
}

func (h *Handler) PostAPIHandler(c *gin.Context) {
	var req models.RequestAPIShorten
	dec := json.NewDecoder(c.Request.Body)
	if err := dec.Decode(&req); err != nil {
		h.logger.Sugar().Error("cannot decode request JSON body")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	fullURL := req.URL

	if _, err := url.ParseRequestURI(fullURL); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		h.logger.Sugar().Errorf("%s is not URL", fullURL)
		return
	}

	shortURL := h.shortener.GetShortURL(fullURL)
	fullShortURL := h.config.BaseURL + "/" + shortURL
	resp := models.ResponseAPIShorten{Result: fullShortURL}
	c.AbortWithStatusJSON(http.StatusCreated, resp)
}
