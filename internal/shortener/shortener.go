package shortener

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	myErrors "github.com/tiunovvv/go-yandex-shortener/internal/errors"
	"github.com/tiunovvv/go-yandex-shortener/internal/models"
	"github.com/tiunovvv/go-yandex-shortener/internal/storage"
	"go.uber.org/zap"
)

type Shortener struct {
	store  storage.Store
	logger *zap.Logger
}

func NewShortener(store storage.Store, logger *zap.Logger) *Shortener {
	return &Shortener{
		store:  store,
		logger: logger,
	}
}

func (sh *Shortener) GetShortURL(ctx context.Context, fullURL string, userID string) (string, error) {
	shortURL := generateShortURL()
	err := sh.store.SaveURL(ctx, shortURL, fullURL, userID)
	if errors.Is(err, myErrors.ErrURLAlreadySaved) {
		shortURL := sh.store.GetShortURL(ctx, fullURL)
		return shortURL, myErrors.ErrURLAlreadySaved
	}

	for errors.Is(err, myErrors.ErrKeyAlreadyExists) {
		shortURL = generateShortURL()
		err = sh.store.SaveURL(ctx, shortURL, fullURL, userID)
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
		return nil, fmt.Errorf("failed to save URL Batch: %w", err)
	}

	return resSlice, nil
}

func (sh *Shortener) GetFullURL(ctx context.Context, shortURL string) (string, bool, error) {
	fullURL, deleteFlag, err := sh.store.GetFullURL(ctx, shortURL)
	if err != nil {
		return "", false, fmt.Errorf("failed to get fullURL from filestore: %w", err)
	}
	return fullURL, deleteFlag, nil
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

func (sh *Shortener) SetDeletedFlag(ctx context.Context, userID string, shortURLSlice []string) {
	jobQueue := make(chan Job, len(shortURLSlice))
	const countOfWorkers = 3
	dispatcher := NewDispatcher(ctx, countOfWorkers, jobQueue, sh.store, sh.logger)

	go func() {
		var wg sync.WaitGroup
		for _, shortURL := range shortURLSlice {
			dispatcher.jobQueue <- Job{userID: userID, shortURL: shortURL}
		}
		close(dispatcher.jobQueue)
		for _, worker := range dispatcher.workerPool {
			wg.Add(1)
			go func(w *Worker) {
				defer wg.Done()
				w.Start(ctx)
			}(worker)
		}
		wg.Wait()
	}()
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
