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
	Config      *config.Config
	MemoryStore *MemoryStore
	logger      *zap.Logger
	file        *os.File
}

func NewFileStore(config *config.Config, logger *zap.Logger) *FileStore {
	memoryStore := &MemoryStore{Urls: make(map[string]string)}
	const perm = 0666
	file, err := os.OpenFile(config.FileStoragePath, os.O_CREATE|os.O_RDONLY, perm)
	if err != nil {
		logger.Sugar().Errorf("can`t open file: %s, %w", config.FileStoragePath, err)
	}
	f := &FileStore{MemoryStore: memoryStore, logger: logger, file: file}
	if err := f.LoadURLs(config.FileStoragePath); err != nil {
		logger.Sugar().Errorf("error getting data from temp file", err)
	}
	return f
}

func (f *FileStore) LoadURLs(filename string) error {
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

	// if err := f.file.Close(); err != nil {
	// 	return fmt.Errorf("error closing temp file %w", err)
	// }

	for k, v := range urls {
		if err := f.MemoryStore.SaveURL(v, k); err != nil {
			return fmt.Errorf("error saving in local memory %w", err)
		}
	}

	return nil
}

func (f *FileStore) SaveURLInFile(shortURL string, fullURL string) error {
	writer := bufio.NewWriter(f.file)

	u := URLsJSON{
		UUID:        strconv.Itoa(len(f.MemoryStore.Urls)),
		ShortURL:    shortURL,
		OriginalURL: fullURL}

	data, err := json.Marshal(u)
	if err != nil {
		return fmt.Errorf("error masrshaling data %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("error writing data into temp file %w", err)
	}

	if err := writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("error writing newline into temp file %w", err)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("error flushing temp file %w", err)
	}

	return nil
}
