package shortener

import (
	"errors"
	"math/rand"
	"time"

	"github.com/tiunovvv/go-yandex-shortener/internal/storage"

	myErrors "github.com/tiunovvv/go-yandex-shortener/internal/errors"
)

type Shortener struct {
	Storage *storage.FileStore
}

func NewShortener(storage *storage.FileStore) *Shortener {
	return &Shortener{
		Storage: storage,
	}
}

func (sh *Shortener) GetShortURL(fullURL string) string {
	if shortURL := sh.Storage.MemoryStore.GetShortURL(fullURL); shortURL != "" {
		return shortURL
	}
	shortURL := sh.GenerateShortURL()
	for errors.Is(sh.Storage.MemoryStore.SaveURL(fullURL, shortURL), myErrors.ErrKeyAlreadyExists) {
		shortURL = sh.GenerateShortURL()
	}

	sh.Storage.SaveURLInFile(fullURL, shortURL)

	return shortURL
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
