package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type Consumer struct {
	file    *os.File
	scanner *bufio.Scanner
}

func NewConsumer(filename string) (*Consumer, error) {
	const (
		perm = 0666
	)
	file, err := os.OpenFile(filename, os.O_RDONLY, perm)
	if err != nil {
		return nil, fmt.Errorf("error opening file %w", err)
	}

	return &Consumer{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *Consumer) ReadURLs(urls map[string]string) error {
	for c.scanner.Scan() {
		urlsJSON := URLsJSON{}
		err := json.Unmarshal(c.scanner.Bytes(), &urlsJSON)
		if err != nil {
			return fmt.Errorf("error unmarshalling string from temp file %w", err)
		}
		urls[urlsJSON.ShortURL] = urlsJSON.OriginalURL
	}

	return nil
}

func (c *Consumer) Close() error {
	err := c.file.Close()
	if err != nil {
		return fmt.Errorf("error closing file %w", err)
	}

	return nil
}
