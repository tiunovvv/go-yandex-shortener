package shortener

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	myErrors "github.com/tiunovvv/go-yandex-shortener/internal/errors"
	"github.com/tiunovvv/go-yandex-shortener/internal/models"
	"github.com/tiunovvv/go-yandex-shortener/internal/storage"
)

type Shortener struct {
	store storage.Store
}

func NewShortener(store storage.Store) *Shortener {
	return &Shortener{
		store: store,
	}
}

func (sh *Shortener) GetShortURL(ctx context.Context, fullURL string, userID string) (string, error) {
	if shortURL := sh.store.GetShortURL(ctx, fullURL); len(shortURL) != 0 {
		return shortURL, myErrors.ErrURLAlreadySaved
	}
	shortURL := generateShortURL()
	for errors.Is(sh.store.SaveURL(ctx, shortURL, fullURL, userID), myErrors.ErrKeyAlreadyExists) {
		shortURL = generateShortURL()
	}

	return shortURL, nil
}

func (sh *Shortener) GetShortURLBatch(
	ctx context.Context,
	reqSlice []models.ReqAPIBatch,
	userID string) ([]models.ResAPIBatch, error) {
	urls := make(map[string]string)
	resSlice := make([]models.ResAPIBatch, 0, len(reqSlice))
	for _, req := range reqSlice {
		res := models.ResAPIBatch{ID: req.ID, ShortURL: generateShortURL()}
		resSlice = append(resSlice, res)
		urls[res.ShortURL] = req.FullURL
	}

	if err := sh.store.SaveURLBatch(ctx, urls, userID); err != nil {
		return nil, fmt.Errorf("failed to get short URLs: %w", err)
	}

	return resSlice, nil
}

func (sh *Shortener) GetFullURL(ctx context.Context, shortURL string) (string, error) {
	fullURL, err := sh.store.GetFullURL(ctx, shortURL)
	if err != nil {
		return "", fmt.Errorf("failed to get fullURL from filestore: %w", err)
	}
	return fullURL, nil
}

func (sh *Shortener) GetURLByUserID(ctx context.Context, baseURL string, userID string) []models.UsersURLs {
	urls := sh.store.GetURLByUserID(ctx, userID)
	userURLs := make([]models.UsersURLs, 0, len(urls))
	for k, v := range urls {
		userURL := models.UsersURLs{ShortURL: fmt.Sprintf("%s/%s", baseURL, k), OriginalURL: v}
		userURLs = append(userURLs, userURL)
	}
	return userURLs
}

func (sh *Shortener) CheckConnect(ctx context.Context) error {
	if err := sh.store.GetPing(ctx); err != nil {
		return fmt.Errorf("failed to connect store: %w", err)
	}
	return nil
}

func generateShortURL() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	const length = 8
	str := make([]byte, length)

	charset := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	for i := range str {
		str[i] = charset[rand.Intn(len(charset))]
	}

	return string(str)
}
