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

// Shortener contains bussines logic.
type Shortener struct {
	store storage.Store
	log   *zap.SugaredLogger
}

// NewShortener creates new Shortener.
func NewShortener(store storage.Store, log *zap.SugaredLogger) *Shortener {
	return &Shortener{
		store: store,
		log:   log,
	}
}

// GetShortURL generates and saves new short URL or gets it from storage if exists.
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

// GetShortURLBatch generates and saves new short URL list or gets it from storage if exists.
func (sh *Shortener) GetShortURLBatch(
	ctx context.Context,
	reqSlice []models.ReqAPIBatch,
	userID string,
) ([]models.ResAPIBatch, error) {
	urls := make(map[string]string, len(reqSlice))
	resSlice := make([]models.ResAPIBatch, len(reqSlice))
	i := 0
	for _, req := range reqSlice {
		shortURL := generateShortURL()
		resSlice[i] = models.ResAPIBatch{ID: req.ID, ShortURL: shortURL}
		urls[resSlice[i].ShortURL] = req.FullURL
		i++
	}

	if err := sh.store.SaveURLBatch(ctx, urls, userID); err != nil {
		return nil, fmt.Errorf("failed to save URL Batch: %w", err)
	}

	return resSlice, nil
}

// GetFullURL gets full URL from storage by short URL.
func (sh *Shortener) GetFullURL(ctx context.Context, shortURL string) (string, bool, error) {
	fullURL, deleteFlag, err := sh.store.GetFullURL(ctx, shortURL)
	if err != nil {
		return "", false, fmt.Errorf("failed to get fullURL from filestore: %w", err)
	}
	return fullURL, deleteFlag, nil
}

// GetURLByUserID returns list of user URLs.
func (sh *Shortener) GetURLByUserID(ctx context.Context, baseURL string, userID string) []models.UsersURLs {
	urls := sh.store.GetURLByUserID(ctx, userID)
	userURLs := make([]models.UsersURLs, len(urls))

	i := 0
	for k, v := range urls {
		shortURL := fmt.Sprintf("%s/%s", baseURL, k)
		userURLs[i] = models.UsersURLs{ShortURL: shortURL, OriginalURL: v}
		i++
	}
	return userURLs
}

// CheckConnect checks connection to storage.
func (sh *Shortener) CheckConnect(ctx context.Context) error {
	if err := sh.store.GetPing(ctx); err != nil {
		return fmt.Errorf("failed to connect store: %w", err)
	}
	return nil
}

// SetDeletedFlag acynchronously sets deleted flag in storage for list of short URL in request body.
func (sh *Shortener) SetDeletedFlag(ctx context.Context, userID string, shortURLSlice []string) {
	const countOfWorkers = 3
	jobQueue := make(chan Job, countOfWorkers)
	dispatcher := NewDispatcher(ctx, countOfWorkers, jobQueue, sh.store, sh.log)

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

// generateShortURL generates new short URL from random 8 symbols.
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
