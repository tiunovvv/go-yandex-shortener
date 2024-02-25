package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
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
	router := gin.New()

	router.Use(middleware.GinGzip(h.logger))
	router.Use(middleware.GinLogger(h.logger))

	const seconds = 5 * time.Second
	router.Use(middleware.GinTimeOut(seconds, "timeout error"))

	const keyLength = 32
	var cookieStore = cookie.NewStore(securecookie.GenerateRandomKey(keyLength))
	router.Use(sessions.Sessions("mysession", cookieStore))
	router.Use(middleware.SetCookie(h.logger))

	router.POST("/", h.PostHandler)
	router.POST("/api/shorten", h.PostAPI)
	router.POST("/api/shorten/batch", h.PostAPIBatch)
	router.GET("/api/user/urls", h.PostAPIUserURLs)
	router.GET("/:id", h.GetHandler)
	router.GET("/ping", h.GetPing)
	router.DELETE("/api/user/urls", h.SetDeletedFlag)
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

	userID, status := h.getUserID(c)
	if len(userID) == 0 {
		c.AbortWithStatus(status)
	}

	shortURL, err := h.shortener.GetShortURL(c, fullURL, userID)

	if errors.Is(err, myErrors.ErrURLAlreadySaved) {
		c.Status(http.StatusConflict)
	} else {
		c.Status(http.StatusCreated)
	}

	fullShortURL, err := url.JoinPath(h.config.BaseURL, c.Request.URL.RequestURI(), shortURL)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		h.logger.Sugar().Errorf("fialed to join path: %s %s %s", h.config.BaseURL, c.Request.URL.RequestURI(), shortURL)
		return
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

	fullURL, deletedFlag, err := h.shortener.GetFullURL(c, shortURL)

	if err != nil {
		newErrorResponce(c, http.StatusBadRequest, err.Error())
		return
	}

	if deletedFlag {
		c.AbortWithStatus(http.StatusGone)
		return
	}

	c.Writer.Header().Set("Location", fullURL)
	c.AbortWithStatus(http.StatusTemporaryRedirect)
}

func (h *Handler) GetPing(c *gin.Context) {
	if err := h.shortener.CheckConnect(c); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	c.AbortWithStatus(http.StatusOK)
}

func (h *Handler) PostAPI(c *gin.Context) {
	var req models.ReqAPI
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Sugar().Infof("failed to decode request JSON body: %w", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	fullURL := req.URL

	if _, err := url.ParseRequestURI(fullURL); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		h.logger.Sugar().Infof("%s is not URL", fullURL)
		return
	}

	userID, status := h.getUserID(c)
	if len(userID) == 0 {
		c.AbortWithStatus(status)
	}

	shortURL, err := h.shortener.GetShortURL(c, fullURL, userID)
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

	userID, status := h.getUserID(c)
	if len(userID) == 0 {
		c.AbortWithStatus(status)
	}

	shortURLSlice, err := h.shortener.GetShortURLBatch(c, fullURLSlice, userID)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		h.logger.Sugar().Error("failed to save list of URLs: %w", err)
		return
	}

	for i := 0; i < len(shortURLSlice); i++ {
		shortURLSlice[i].ShortURL = fmt.Sprintf("%s/%s", h.config.BaseURL, shortURLSlice[i].ShortURL)
	}

	c.AbortWithStatusJSON(http.StatusCreated, shortURLSlice)
}

func (h *Handler) PostAPIUserURLs(c *gin.Context) {
	userID, status := h.getUserID(c)
	if len(userID) == 0 {
		c.AbortWithStatus(status)
	}

	usersURLs := h.shortener.GetURLByUserID(c, h.config.BaseURL, userID)

	if len(usersURLs) == 0 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.AbortWithStatusJSON(http.StatusOK, usersURLs)
}

func (h *Handler) SetDeletedFlag(c *gin.Context) {
	userID, status := h.getUserID(c)
	if len(userID) == 0 {
		c.AbortWithStatus(status)
	}

	var shortURLSlice []string

	if err := c.ShouldBindJSON(&shortURLSlice); err != nil {
		h.logger.Sugar().Error("failed to bind request JSON body: %w", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	h.shortener.SetDeletedFlag(c, userID, shortURLSlice)

	c.AbortWithStatus(http.StatusAccepted)
}

func (h *Handler) getUserID(c *gin.Context) (string, int) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		return "", http.StatusUnauthorized
	}

	userID, ok := userIDInterface.(string)

	if !ok {
		h.logger.Sugar().Errorf("failed to get userID from %v", userIDInterface)
		return "", http.StatusInternalServerError
	}

	return userID, http.StatusOK
}
