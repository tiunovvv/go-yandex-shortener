package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type Producer struct {
	file   *os.File
	writer *bufio.Writer
}

func NewProducer(filename string) (*Producer, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if _, err := os.Create(filename); err != nil {
			return nil, fmt.Errorf("error creating temp file %w", err)
		}
	}
	const (
		perm = 0600
	)
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, perm)
	if err != nil {
		return nil, fmt.Errorf("error opening temp file %w", err)
	}

	return &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *Producer) WriteURL(u URLsJSON) error {
	data, err := json.Marshal(u)
	if err != nil {
		return fmt.Errorf("error masrshaling data %w", err)
	}

	if _, err := p.writer.Write(data); err != nil {
		return fmt.Errorf("error writing data into temp file %w", err)
	}

	if err := p.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("error writing newline into temp file %w", err)
	}

	if err := p.writer.Flush(); err != nil {
		return fmt.Errorf("error flushing temp file %w", err)
	}
	return nil
}
