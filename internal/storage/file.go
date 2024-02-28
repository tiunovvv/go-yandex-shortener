package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	myErrors "github.com/tiunovvv/go-yandex-shortener/internal/errors"
	"go.uber.org/zap"
)

// URLsJSON structer of file memmory.
type URLsJSON struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// File struct for file storage.
type File struct {
	memory *Memory
	file   *os.File
	log    *zap.SugaredLogger
}

// NewFile creates new File storage of get URLs from existing file.
func NewFile(filePath string, log *zap.SugaredLogger) (Store, error) {
	memory := &Memory{urls: make(map[string]URLInfo)}
	const perm = 0666

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, perm)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s, %w", filePath, err)
	}
	f := &File{memory: memory, file: file, log: log}
	if err := f.loadURLs(); err != nil {
		f.log.Info("failed to get data from temp file", err)
	}
	return f, nil
}

// loadURLs loads URLs from temp file.
func (f *File) loadURLs() error {
	scanner := bufio.NewScanner(f.file)

	urls := make(map[string]URLInfo)
	for scanner.Scan() {
		urlsJSON := URLsJSON{}
		err := json.Unmarshal(scanner.Bytes(), &urlsJSON)
		if err != nil {
			return fmt.Errorf("failed to unmarshall temp file %w", err)
		}
		urls[urlsJSON.ShortURL] = URLInfo{fullURL: urlsJSON.OriginalURL}
	}
	for k, v := range urls {
		if err := f.memory.SaveURL(context.Background(), k, v.fullURL, v.userID); err != nil {
			return fmt.Errorf("failed to save in local memory %w", err)
		}
	}

	return nil
}

// SaveURL saves all info about new URL in local map and in temp file.
func (f *File) SaveURL(ctx context.Context, shortURL string, fullURL string, userID string) error {
	if err := f.memory.SaveURL(ctx, shortURL, fullURL, userID); err != nil {
		if errors.Is(err, myErrors.ErrURLAlreadySaved) {
			return err
		}
		return fmt.Errorf("failed to save in local memory %w", err)
	}

	f.writeURLInFile(shortURL, fullURL)

	return nil
}

// GetFullURL fullURL by shortURL from in-memmory.
func (f *File) GetFullURL(ctx context.Context, shortURL string) (string, bool, error) {
	return f.memory.GetFullURL(ctx, shortURL)
}

// GetShortURL shortURL by fullURL from in-memmory.
func (f *File) GetShortURL(ctx context.Context, fullURL string) string {
	return f.memory.GetShortURL(ctx, fullURL)
}

// SaveURLBatch saves all info about new URL list in local map and in temp file.
func (f *File) SaveURLBatch(ctx context.Context, urls map[string]string, userID string) error {
	if err := f.memory.SaveURLBatch(ctx, urls, userID); err != nil {
		return fmt.Errorf("failed to save URL slice %w", err)
	}

	for k, v := range urls {
		f.writeURLInFile(k, v)
	}

	return nil
}

// GetURLByUserID returns list of user URLs from in-memmory.
func (f *File) GetURLByUserID(ctx context.Context, userID string) map[string]string {
	return f.memory.GetURLByUserID(ctx, userID)
}

// SetDeletedFlag sets deleted flag for short URL in local map.
func (f *File) SetDeletedFlag(ctx context.Context, userID string, shortURL string) error {
	return f.memory.SetDeletedFlag(ctx, userID, shortURL)
}

// GetPing method needed for implementation of Store interface.
func (f *File) GetPing(ctx context.Context) error {
	return nil
}

// Close closes temp file.
func (f *File) Close() error {
	if err := f.file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}

// writeURLInFile save shortURL and fullURL in temp file.
func (f *File) writeURLInFile(shortURL string, fullURL string) {
	writer := bufio.NewWriter(f.file)

	u := URLsJSON{
		UUID:        strconv.Itoa(len(f.memory.urls)),
		ShortURL:    shortURL,
		OriginalURL: fullURL}

	data, err := json.Marshal(u)
	if err != nil {
		f.log.Errorf("failed to write masrshaling data %w", err)
		return
	}

	if _, err := writer.Write(data); err != nil {
		f.log.Errorf("failed to write data into temp file %w", err)
		return
	}

	if err := writer.WriteByte('\n'); err != nil {
		f.log.Errorf("failed to write newline into temp file %w", err)
		return
	}

	if err := writer.Flush(); err != nil {
		f.log.Errorf("failed to flush temp file %w", err)
		return
	}
}
