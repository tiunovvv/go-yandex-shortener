package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/tiunovvv/go-yandex-shortener/internal/models"
	"go.uber.org/zap"
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
}

func NewFileStore(filePath string, logger *zap.Logger) (Store, error) {
	inMemoryStore := &InMemoryStore{urls: make(map[string]string)}
	const perm = 0666

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, perm)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s, %w", filePath, err)
	}
	f := &FileStore{inMemoryStore: inMemoryStore, file: file, logger: logger}
	if err := f.loadURLs(); err != nil {
		logger.Sugar().Info("failed to get data from temp file", err)
	}
	return f, nil
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
		if err := f.inMemoryStore.SaveURL(context.Background(), k, v); err != nil {
			return fmt.Errorf("failed to save in local memory %w", err)
		}
	}

	return nil
}

func (f *FileStore) SaveURL(ctx context.Context, shortURL string, fullURL string) error {
	if err := f.inMemoryStore.SaveURL(ctx, shortURL, fullURL); err != nil {
		return fmt.Errorf("failed to save in local memory %w", err)
	}

	f.writeURLInFile(shortURL, fullURL)

	return nil
}

func (f *FileStore) GetFullURL(ctx context.Context, shortURL string) (string, error) {
	return f.inMemoryStore.GetFullURL(ctx, shortURL)
}

func (f *FileStore) GetShortURL(ctx context.Context, fullURL string) string {
	return f.inMemoryStore.GetShortURL(ctx, fullURL)
}

func (f *FileStore) GetShortURLBatch(ctx context.Context, reqSlice []models.ReqAPIBatch) ([]models.ResAPIBatch, error) {
	resSlice, err := f.inMemoryStore.GetShortURLBatch(ctx, reqSlice)
	if err != nil {
		return nil, fmt.Errorf("failed to save URL slice %w", err)
	}

	for _, res := range resSlice {
		for _, req := range reqSlice {
			if res.ID == req.ID {
				f.writeURLInFile(res.ShortURL, req.FullURL)
				break
			}
		}
	}

	return resSlice, nil
}

func (f *FileStore) GetPing(ctx context.Context) error {
	return fmt.Errorf("failed to connect to database")
}

func (f *FileStore) Close() error {
	if err := f.file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}

func (f *FileStore) GenerateShortURL() string {
	return f.inMemoryStore.GenerateShortURL()
}

func (f *FileStore) writeURLInFile(shortURL string, fullURL string) {
	writer := bufio.NewWriter(f.file)

	u := URLsJSON{
		UUID:        strconv.Itoa(len(f.inMemoryStore.urls)),
		ShortURL:    shortURL,
		OriginalURL: fullURL}

	data, err := json.Marshal(u)
	if err != nil {
		f.logger.Sugar().Errorf("failed to write masrshaling data %w", err)
		return
	}

	if _, err := writer.Write(data); err != nil {
		f.logger.Sugar().Errorf("failed to write data into temp file %w", err)
		return
	}

	if err := writer.WriteByte('\n'); err != nil {
		f.logger.Sugar().Errorf("failed to write newline into temp file %w", err)
		return
	}

	if err := writer.Flush(); err != nil {
		f.logger.Sugar().Errorf("failed to flush temp file %w", err)
		return
	}
}
