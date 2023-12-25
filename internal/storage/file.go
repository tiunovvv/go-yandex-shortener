package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"go.uber.org/zap"
)

type URLsJSON struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type File struct {
	memory *Memory
	file   *os.File
	logger *zap.Logger
}

func NewFile(filePath string, logger *zap.Logger) (Store, error) {
	memory := &Memory{urls: make(map[string]URLInfo)}
	const perm = 0666

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, perm)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %s, %w", filePath, err)
	}
	f := &File{memory: memory, file: file, logger: logger}
	if err := f.loadURLs(); err != nil {
		logger.Sugar().Info("failed to get data from temp file", err)
	}
	return f, nil
}

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

func (f *File) SaveURL(ctx context.Context, shortURL string, fullURL string, userID string) error {
	if err := f.memory.SaveURL(ctx, shortURL, fullURL, userID); err != nil {
		return fmt.Errorf("failed to save in local memory %w", err)
	}

	f.writeURLInFile(shortURL, fullURL)

	return nil
}

func (f *File) GetFullURL(ctx context.Context, shortURL string) (string, bool, error) {
	return f.memory.GetFullURL(ctx, shortURL)
}

func (f *File) GetShortURL(ctx context.Context, fullURL string) string {
	return f.memory.GetShortURL(ctx, fullURL)
}

func (f *File) SaveURLBatch(ctx context.Context, urls map[string]string, userID string) error {
	if err := f.memory.SaveURLBatch(ctx, urls, userID); err != nil {
		return fmt.Errorf("failed to save URL slice %w", err)
	}

	for k, v := range urls {
		f.writeURLInFile(k, v)
	}

	return nil
}

func (f *File) GetURLByUserID(ctx context.Context, userID string) map[string]string {
	return f.memory.GetURLByUserID(ctx, userID)
}

func (f *File) SetDeletedFlag(ctx context.Context, userID string, shortURL string) error {
	return f.memory.SetDeletedFlag(ctx, userID, shortURL)
}

func (f *File) GetPing(ctx context.Context) error {
	return nil
}

func (f *File) Close() error {
	if err := f.file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}

func (f *File) writeURLInFile(shortURL string, fullURL string) {
	writer := bufio.NewWriter(f.file)

	u := URLsJSON{
		UUID:        strconv.Itoa(len(f.memory.urls)),
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
