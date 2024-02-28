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

// Handler handle requests using config logger and bussines logic.
type Handler struct {
	cfg *config.Config
	sh  *shortener.Shortener
	log *zap.SugaredLogger
}

// NewHandler creates and returns new Handler.
func NewHandler(cfg *config.Config, sh *shortener.Shortener, log *zap.SugaredLogger) *Handler {
	return &Handler{
		cfg: cfg,
		sh:  sh,
		log: log,
	}
}

// InitRoutes initializes and returns router.
func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.Use(middleware.GinGzip(h.log))
	router.Use(middleware.GinLogger(h.log))

	const seconds = 5 * time.Second
	router.Use(middleware.GinTimeOut(seconds, "timeout error"))

	const keyLength = 32
	var cookieStore = cookie.NewStore(securecookie.GenerateRandomKey(keyLength))
	router.Use(sessions.Sessions("mysession", cookieStore))
	router.Use(middleware.SetCookie(h.log))

	router.POST("/", h.PostHandler)
	router.POST("/api/shorten", h.PostAPI)
	router.POST("/api/shorten/batch", h.PostAPIBatch)
	router.GET("/api/user/urls", h.PostAPIUserURLs)
	router.GET("/:id", h.GetHandler)
	router.GET("/ping", h.GetPing)
	router.DELETE("/api/user/urls", h.SetDeletedFlag)
	return router
}

// PostHandler creates short URL from full URL in request body.
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
		h.log.Errorf("%s is not URL", fullURL)
		return
	}

	userID := h.getUserID(c)
	if len(userID) == 0 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	shortURL, err := h.sh.GetShortURL(c, fullURL, userID)

	if errors.Is(err, myErrors.ErrURLAlreadySaved) {
		c.Status(http.StatusConflict)
	} else {
		c.Status(http.StatusCreated)
	}

	fullShortURL, err := url.JoinPath(h.cfg.BaseURL, c.Request.URL.RequestURI(), shortURL)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		h.log.Errorf("fialed to join path: %s %s %s", h.cfg.BaseURL, c.Request.URL.RequestURI(), shortURL)
		return
	}

	if _, err := c.Writer.Write([]byte(fullShortURL)); c.Request.Body == nil && err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		h.log.Errorf("failed to write %s into body: %w", fullShortURL, err)
		return
	}
}

// GetHandler returns full URL by short URL like http://localhost:8080/KMlLwVjl.
func (h *Handler) GetHandler(c *gin.Context) {
	path := c.Request.URL.RequestURI()
	shortURL := strings.ReplaceAll(path, "/", "")

	if shortURL == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, "is not a form baseURL/shortURL")
		return
	}

	fullURL, deletedFlag, err := h.sh.GetFullURL(c, shortURL)

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

// GetPing returns 200 if server is up and running.
func (h *Handler) GetPing(c *gin.Context) {
	if err := h.sh.CheckConnect(c); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	c.AbortWithStatus(http.StatusOK)
}

// PostAPI returns short URL from storage for full URL in request body.
func (h *Handler) PostAPI(c *gin.Context) {
	var req models.ReqAPI
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Infof("failed to decode request JSON body: %w", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	fullURL := req.URL

	if _, err := url.ParseRequestURI(fullURL); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		h.log.Infof("%s is not URL", fullURL)
		return
	}

	userID := h.getUserID(c)
	if len(userID) == 0 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	shortURL, err := h.sh.GetShortURL(c, fullURL, userID)
	fullShortURL := fmt.Sprintf("%s/%s", h.cfg.BaseURL, shortURL)
	resp := models.ResAPI{Result: fullShortURL}

	if errors.Is(err, myErrors.ErrURLAlreadySaved) {
		c.AbortWithStatusJSON(http.StatusConflict, resp)
		return
	}

	c.AbortWithStatusJSON(http.StatusCreated, resp)
}

// PostAPIBatch generates short URL list for list of full URL in request body.
func (h *Handler) PostAPIBatch(c *gin.Context) {
	var fullURLSlice []models.ReqAPIBatch

	if err := c.ShouldBindJSON(&fullURLSlice); err != nil {
		h.log.Error("failed to bind request JSON body: %w", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	userID := h.getUserID(c)
	if len(userID) == 0 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	shortURLSlice, err := h.sh.GetShortURLBatch(c, fullURLSlice, userID)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		h.log.Error("failed to save list of URLs: %w", err)
		return
	}

	for i := 0; i < len(shortURLSlice); i++ {
		shortURLSlice[i].ShortURL = fmt.Sprintf("%s/%s", h.cfg.BaseURL, shortURLSlice[i].ShortURL)
	}

	c.AbortWithStatusJSON(http.StatusCreated, shortURLSlice)
}

// PostAPIUserURLs returns list of short and full URLs for current user.
func (h *Handler) PostAPIUserURLs(c *gin.Context) {
	userID := h.getUserID(c)
	if len(userID) == 0 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	usersURLs := h.sh.GetURLByUserID(c, h.cfg.BaseURL, userID)

	if len(usersURLs) == 0 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.AbortWithStatusJSON(http.StatusOK, usersURLs)
}

// SetDeletedFlag sets deleted flag in storagefor list of short URL in request body.
func (h *Handler) SetDeletedFlag(c *gin.Context) {
	userID := h.getUserID(c)
	if len(userID) == 0 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var shortURLSlice []string

	if err := c.ShouldBindJSON(&shortURLSlice); err != nil {
		h.log.Error("failed to bind request JSON body: %w", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	h.sh.SetDeletedFlag(c, userID, shortURLSlice)

	c.AbortWithStatus(http.StatusAccepted)
}
