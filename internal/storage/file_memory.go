package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/tiunovvv/go-yandex-shortener/internal/config"
	"github.com/tiunovvv/go-yandex-shortener/internal/generator"
	"github.com/tiunovvv/go-yandex-shortener/internal/models"
	"github.com/tiunovvv/go-yandex-shortener/internal/shortener"
	"go.uber.org/zap"

	myErrors "github.com/tiunovvv/go-yandex-shortener/internal/errors"
)

type URLsJSON struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStore struct {
	inMemoryStore *InMemoryStore
	file          *os.File
	logger        *zap.Logger
	*generator.Generator
}

func NewFileStore(config *config.Config, logger *zap.Logger) shortener.Store {
	inMemoryStore := &InMemoryStore{urls: make(map[string]string)}
	const perm = 0666
	file, err := os.OpenFile(config.FileStoragePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, perm)
	if err != nil {
		logger.Sugar().Errorf("failed to open file: %s, %w", config.FileStoragePath, err)
	}
	f := &FileStore{inMemoryStore: inMemoryStore, file: file, logger: logger}
	if err := f.loadURLs(); err != nil {
		logger.Sugar().Errorf("failed to gett data from temp file", err)
	}
	return f
}

func (f *FileStore) loadURLs() error {
	scanner := bufio.NewScanner(f.file)

	urls := make(map[string]string)
	for scanner.Scan() {
		urlsJSON := URLsJSON{}
		err := json.Unmarshal(scanner.Bytes(), &urlsJSON)
		if err != nil {
			return fmt.Errorf("failed to unmarshall temp file %w", err)
		}
		urls[urlsJSON.ShortURL] = urlsJSON.OriginalURL
	}
	for k, v := range urls {
		if err := f.inMemoryStore.SaveURL(k, v); err != nil {
			return fmt.Errorf("failed to save in local memory %w", err)
		}
	}

	return nil
}

func (f *FileStore) SaveURL(ctx context.Context, shortURL string, fullURL string) error {
	if err := f.inMemoryStore.SaveURL(shortURL, fullURL); err != nil {
		return fmt.Errorf("failed to save in local memory %w", err)
	}

	if f.file != nil {
		writer := bufio.NewWriter(f.file)

		u := URLsJSON{
			UUID:        strconv.Itoa(len(f.inMemoryStore.urls)),
			ShortURL:    shortURL,
			OriginalURL: fullURL}

		data, err := json.Marshal(u)
		if err != nil {
			f.logger.Sugar().Errorf("failed to write masrshaling data %w", err)
			return nil
		}

		if _, err := writer.Write(data); err != nil {
			f.logger.Sugar().Errorf("failed to write data into temp file %w", err)
			return nil
		}

		if err := writer.WriteByte('\n'); err != nil {
			f.logger.Sugar().Errorf("failed to write newline into temp file %w", err)
			return nil
		}

		if err := writer.Flush(); err != nil {
			f.logger.Sugar().Errorf("failed to flush temp file %w", err)
			return nil
		}
	}

	return nil
}

func (f *FileStore) GetFullURL(ctx context.Context, shortURL string) (string, error) {
	return f.inMemoryStore.GetFullURL(shortURL)
}

func (f *FileStore) GetShortURL(ctx context.Context, fullURL string) string {
	return f.inMemoryStore.GetShortURL(fullURL)
}

func (f *FileStore) GetShortURLBatch(ctx context.Context, reqSlice []models.ReqAPIBatch) ([]models.ResAPIBatch, error) {
	var resSlice []models.ResAPIBatch
	for _, req := range reqSlice {
		if shortURL := f.GetShortURL(ctx, req.FullURL); shortURL != "" {
			res := models.ResAPIBatch{ID: req.ID, ShortURL: shortURL}
			resSlice = append(resSlice, res)
			continue
		}
		res := models.ResAPIBatch{ID: req.ID, ShortURL: f.GenerateShortURL()}
		for errors.Is(f.SaveURL(ctx, res.ShortURL, req.FullURL), myErrors.ErrKeyAlreadyExists) {
			res.ShortURL = f.GenerateShortURL()
			continue
		}
		resSlice = append(resSlice, res)
	}
	return resSlice, nil
}

func (f *FileStore) GetPing(ctx context.Context) error {
	return fmt.Errorf("failed to connect to databse")
}

func (f *FileStore) CloseStore() error {
	if err := f.file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}
