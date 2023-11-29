package shortener

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/tiunovvv/go-yandex-shortener/internal/storage"

	myErrors "github.com/tiunovvv/go-yandex-shortener/internal/errors"
)

type Shortener struct {
	fileStore *storage.FileStore
}

func NewShortener(fileStore *storage.FileStore) *Shortener {
	return &Shortener{
		fileStore: fileStore,
	}
}

func (sh *Shortener) GetShortURL(fullURL string) string {
	if shortURL := sh.fileStore.GetShortURL(fullURL); shortURL != "" {
		return shortURL
	}
	shortURL := sh.GenerateShortURL()
	for errors.Is(sh.fileStore.SaveURL(fullURL, shortURL), myErrors.ErrKeyAlreadyExists) {
		shortURL = sh.GenerateShortURL()
	}

	return shortURL
}

func (sh *Shortener) GetFullURL(shortURL string) (string, error) {
	fullURL, err := sh.fileStore.GetFullURL(shortURL)
	if err != nil {
		return "", fmt.Errorf("error geting fullURL from filestore: %w", err)
	}
	return fullURL, nil
}

func (sh *Shortener) GenerateShortURL() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	const length = 8
	str := make([]byte, length)

	charset := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	for i := range str {
		str[i] = charset[rand.Intn(len(charset))]
	}

	return string(str)
}
