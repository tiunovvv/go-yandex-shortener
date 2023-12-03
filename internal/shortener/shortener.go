package shortener

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/tiunovvv/go-yandex-shortener/internal/storage"

	myErrors "github.com/tiunovvv/go-yandex-shortener/internal/errors"
)

type Store interface {
	GetShortURL(fullURL string) string
	GetFullURL(shortURL string) (string, error)
	SaveURL(shortURL string, fullURL string) error
}

type Shortener struct {
	store Store
}

func NewShortener(fileStore *storage.FileStore) *Shortener {
	return &Shortener{
		store: fileStore,
	}
}

func (sh *Shortener) GetShortURL(fullURL string) string {
	if shortURL := sh.store.GetShortURL(fullURL); shortURL != "" {
		return shortURL
	}
	shortURL := sh.GenerateShortURL()
	for errors.Is(sh.store.SaveURL(fullURL, shortURL), myErrors.ErrKeyAlreadyExists) {
		shortURL = sh.GenerateShortURL()
	}

	return shortURL
}

func (sh *Shortener) GetFullURL(shortURL string) (string, error) {
	fullURL, err := sh.store.GetFullURL(shortURL)
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
