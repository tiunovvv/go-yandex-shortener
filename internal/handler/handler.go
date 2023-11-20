package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tiunovvv/go-yandex-shortener/internal/compressor"
	"github.com/tiunovvv/go-yandex-shortener/internal/logger"
	"github.com/tiunovvv/go-yandex-shortener/internal/models"
	"github.com/tiunovvv/go-yandex-shortener/internal/shortener"
)

const isNotURL = "%s is not URL"

type Handler struct {
	shortener  *shortener.Shortener
	logger     *logger.Logger
	compressor *compressor.Compressor
}

func NewHandler(shortener *shortener.Shortener, logger *logger.Logger, compressor *compressor.Compressor) *Handler {
	return &Handler{
		shortener:  shortener,
		logger:     logger,
		compressor: compressor,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	router.Use(h.compressor.GinGzipMiddleware())
	router.Use(h.logger.GinLoggerMiddleware())
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
		h.logger.Sugar().Errorf(isNotURL, fullURL)
		return
	}

	shortURL := h.shortener.GetShortURL(fullURL, c.Request.URL.RequestURI())
	c.Status(http.StatusCreated)

	if _, err := c.Writer.Write([]byte(shortURL)); c.Request.Body == nil && err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		h.logger.Sugar().Errorf("Cant write %s into body", shortURL)
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

	fullURL, err := h.shortener.GetFullURL(shortURL)

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
		h.logger.Sugar().Errorf(isNotURL, fullURL)
		return
	}

	shortURL := h.shortener.GetShortURL(fullURL, "/")
	resp := models.ResponseAPIShorten{Result: shortURL}
	c.AbortWithStatusJSON(http.StatusCreated, resp)
}
