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

func (sh *Shortener) GetShortURL(ctx context.Context, fullURL string) (string, error) {
	if shortURL := sh.store.GetShortURL(ctx, fullURL); len(shortURL) != 0 {
		return shortURL, myErrors.ErrURLAlreadySaved
	}
	shortURL := sh.GenerateShortURL()
	for errors.Is(sh.store.SaveURL(ctx, shortURL, fullURL), myErrors.ErrKeyAlreadyExists) {
		shortURL = sh.GenerateShortURL()
	}

	return shortURL, nil
}

func (sh *Shortener) GetShortURLBatch(
	ctx context.Context,
	reqSlice []models.ReqAPIBatch) ([]models.ResAPIBatch, error) {
	var cancelCtx context.CancelFunc
	const seconds = time.Second * 10
	ctx, cancelCtx = context.WithTimeout(ctx, seconds)
	defer cancelCtx()

	urls := make(map[string]string)
	resSlice := make([]models.ResAPIBatch, 0, len(reqSlice))
	for _, req := range reqSlice {
		res := models.ResAPIBatch{ID: req.ID, ShortURL: sh.GenerateShortURL()}
		resSlice = append(resSlice, res)
		urls[res.ShortURL] = req.FullURL
	}

	if err := sh.store.SaveURLBatch(ctx, urls); err != nil {
		return nil, fmt.Errorf("failed to get short URLs: %w", err)
	}

	return resSlice, nil
}

func (sh *Shortener) GetFullURL(ctx context.Context, shortURL string) (string, error) {
	var cancelCtx context.CancelFunc
	const seconds = time.Second * 10
	ctx, cancelCtx = context.WithTimeout(ctx, seconds)
	defer cancelCtx()
	fullURL, err := sh.store.GetFullURL(ctx, shortURL)
	if err != nil {
		return "", fmt.Errorf("failed to get fullURL from filestore: %w", err)
	}
	return fullURL, nil
}

func (sh *Shortener) CheckConnect(ctx context.Context) error {
	if err := sh.store.GetPing(ctx); err != nil {
		return fmt.Errorf("failed to connect store: %w", err)
	}
	return nil
}

func (sh *Shortener) GenerateShortURL() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	const length = 8
	str := make([]byte, length)

	charset := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	for i := range str {
		str[i] = charset[rand.Intn(len(charset))]
	}

	return string(str)
}
