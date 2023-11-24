package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

const (
	perm       = 0666
	errorOpen  = "error opening temp file %w"
	errorClose = "error closing temp file %w"
)

type URLsJSON struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorage struct {
	memoryStorage *MemoryStorage
	filename      string
}

func NewFileStorage(filename string, memoryStorage *MemoryStorage) *FileStorage {
	return &FileStorage{
		filename:      filename,
		memoryStorage: memoryStorage,
	}
}

func (f *FileStorage) LoadURLs() error {
	if f.filename == "" {
		return nil
	}

	file, err := os.OpenFile(f.filename, os.O_CREATE|os.O_RDONLY, perm)
	if err != nil {
		return fmt.Errorf(errorOpen, err)
	}

	scanner := bufio.NewScanner(file)

	urls := make(map[string]string)
	for scanner.Scan() {
		urlsJSON := URLsJSON{}
		err := json.Unmarshal(scanner.Bytes(), &urlsJSON)
		if err != nil {
			return fmt.Errorf("error unmarshalling temp file %w", err)
		}
		urls[urlsJSON.ShortURL] = urlsJSON.OriginalURL
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf(errorClose, err)
	}

	for k, v := range urls {
		if err := f.memoryStorage.SaveURL(v, k); err != nil {
			return fmt.Errorf("error saving in local memory %w", err)
		}
	}

	return nil
}

func (f *FileStorage) SaveURL(shortURL string, fullURL string) error {
	if f.filename == "" {
		return nil
	}

	file, err := os.OpenFile(f.filename, os.O_WRONLY|os.O_APPEND, perm)
	if err != nil {
		return fmt.Errorf(errorOpen, err)
	}

	writer := bufio.NewWriter(file)

	u := URLsJSON{
		UUID:        strconv.Itoa(len(f.memoryStorage.Urls)),
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

	if err := file.Close(); err != nil {
		return fmt.Errorf(errorClose, err)
	}

	return nil
}
