package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"github.com/tiunovvv/go-yandex-shortener/internal/middleware"
	"github.com/tiunovvv/go-yandex-shortener/internal/models"
	"github.com/tiunovvv/go-yandex-shortener/internal/shortener"
	"go.uber.org/zap"

	myErrors "github.com/tiunovvv/go-yandex-shortener/internal/errors"
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
	const seconds = 5 * time.Second

	router := gin.New()
	router.Use(middleware.GinGzip(h.logger))
	router.Use(middleware.GinLogger(h.logger))
	router.Use(middleware.GinTimeOut(seconds, "timeout error"))
	router.POST("/", h.PostHandler)
	router.POST("/api/shorten", h.PostAPIHandler)
	router.POST("/api/shorten/batch", h.PostAPIBatch)
	router.GET("/:id", h.GetHandler)
	router.GET("/ping", h.GetPing)
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

	shortURL, err := h.shortener.GetShortURL(c, fullURL)
	fullShortURL := fmt.Sprintf("%s%s%s", h.config.BaseURL, c.Request.URL.RequestURI(), shortURL)

	if errors.Is(err, myErrors.ErrURLAlreadySaved) {
		c.Status(http.StatusConflict)
	} else {
		c.Status(http.StatusCreated)
	}

	if _, err := c.Writer.Write([]byte(fullShortURL)); c.Request.Body == nil && err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		h.logger.Sugar().Errorf("failed to write %s into body: %w", fullShortURL, err)
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

	fullURL, err := h.shortener.GetFullURL(c, shortURL)

	if err != nil {
		newErrorResponce(c, http.StatusBadRequest, err.Error())
		return
	}

	c.Writer.Header().Set("Location", fullURL)
	c.Status(http.StatusTemporaryRedirect)
}

func (h *Handler) GetPing(c *gin.Context) {
	if err := h.shortener.CheckConnect(c); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) PostAPIHandler(c *gin.Context) {
	var req models.ReqAPI
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Sugar().Error("failed to decode request JSON body: %w", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	fullURL := req.URL

	if _, err := url.ParseRequestURI(fullURL); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		h.logger.Sugar().Errorf("%s is not URL", fullURL)
		return
	}

	shortURL, err := h.shortener.GetShortURL(c, fullURL)
	fullShortURL := fmt.Sprintf("%s/%s", h.config.BaseURL, shortURL)
	resp := models.ResAPI{Result: fullShortURL}

	if errors.Is(err, myErrors.ErrURLAlreadySaved) {
		c.AbortWithStatusJSON(http.StatusConflict, resp)
		return
	}

	c.AbortWithStatusJSON(http.StatusCreated, resp)
}

func (h *Handler) PostAPIBatch(c *gin.Context) {
	var fullURLSlice []models.ReqAPIBatch

	if err := c.ShouldBindJSON(&fullURLSlice); err != nil {
		h.logger.Sugar().Error("failed to bind request JSON body: %w", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	shortURLSlice, err := h.shortener.GetShortURLBatch(c, fullURLSlice)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		h.logger.Sugar().Errorf("failed to save list of URLS")
		return
	}

	for i := 0; i < len(shortURLSlice); i++ {
		shortURLSlice[i].ShortURL = fmt.Sprintf("%s/%s", h.config.BaseURL, shortURLSlice[i].ShortURL)
	}

	c.AbortWithStatusJSON(http.StatusCreated, shortURLSlice)
}
