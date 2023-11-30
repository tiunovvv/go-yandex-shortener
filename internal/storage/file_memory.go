package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/tiunovvv/go-yandex-shortener/internal/config"
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

func NewFileStore(config *config.Config, logger *zap.Logger) *FileStore {
	inMemoryStore := &InMemoryStore{urls: make(map[string]string)}
	const perm = 0666
	file, err := os.OpenFile(config.FileStoragePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, perm)
	if err != nil {
		logger.Sugar().Errorf("can`t open file: %s, %w", config.FileStoragePath, err)
	}
	f := &FileStore{inMemoryStore: inMemoryStore, file: file, logger: logger}
	if err := f.loadURLs(); err != nil {
		logger.Sugar().Errorf("error getting data from temp file", err)
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
			return fmt.Errorf("error unmarshalling temp file %w", err)
		}
		urls[urlsJSON.ShortURL] = urlsJSON.OriginalURL
	}
	for k, v := range urls {
		if err := f.inMemoryStore.SaveURL(k, v); err != nil {
			return fmt.Errorf("error saving in local memory %w", err)
		}
	}

	return nil
}

func (f *FileStore) SaveURL(shortURL string, fullURL string) error {
	if err := f.inMemoryStore.SaveURL(shortURL, fullURL); err != nil {
		return fmt.Errorf("error saving in local memory %w", err)
	}

	writer := bufio.NewWriter(f.file)

	u := URLsJSON{
		UUID:        strconv.Itoa(len(f.inMemoryStore.urls)),
		ShortURL:    shortURL,
		OriginalURL: fullURL}

	data, err := json.Marshal(u)
	if err != nil {
		f.logger.Sugar().Errorf("error masrshaling data %w", err)
		return nil
	}

	if _, err := writer.Write(data); err != nil {
		f.logger.Sugar().Errorf("error writing data into temp file %w", err)
		return nil
	}

	if err := writer.WriteByte('\n'); err != nil {
		f.logger.Sugar().Errorf("error writing newline into temp file %w", err)
		return nil
	}

	if err := writer.Flush(); err != nil {
		f.logger.Sugar().Errorf("error flushing temp file %w", err)
		return nil
	}

	return nil
}

func (f *FileStore) GetFullURL(shortURL string) (string, error) {
	return f.inMemoryStore.GetFullURL(shortURL)
}

func (f *FileStore) GetShortURL(fullURL string) string {
	return f.inMemoryStore.GetShortURL(fullURL)
}
