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
	storage *storage.Storage
}

func NewShortener(storage *storage.Storage) *Shortener {
	return &Shortener{
		storage: storage,
	}
}

func (sh *Shortener) GetShortURL(fullURL string, path string) string {
	if shortURL := sh.storage.FindByFullURL(fullURL); shortURL != "" {
		return sh.storage.Config.BaseURL + path + shortURL
	}

	shortURL := sh.GenerateShortURL()
	for errors.Is(sh.storage.SaveURL(fullURL, shortURL), myErrors.ErrKeyAlreadyExists) {
		shortURL = sh.GenerateShortURL()
	}

	return sh.storage.Config.BaseURL + path + shortURL
}

func (sh *Shortener) GetFullURL(shortURL string) (string, error) {
	if fullURL := sh.storage.GetFullURL(shortURL); fullURL != "" {
		return fullURL, nil
	}

	return "", fmt.Errorf("URL `%s` not found", shortURL)
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
