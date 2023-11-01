package storage

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/tiunovvv/go-yandex-shortener/cmd/config"
)

const (
	charset        = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	shortURLLength = 8
	schemePrefix   = "://"
	urlNotFound    = "URL not found"
)

type URLShortener struct {
	Urls   map[string]string
	Config *config.Config
}

func GenerateShortURL() ([]byte, error) {
	ret := make([]byte, shortURLLength)
	for i := 0; i < shortURLLength; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return ret, err
		}
		ret[i] = charset[num.Int64()]
	}
	return ret, nil
}

func CreateStorage(config *config.Config) *URLShortener {
	return &URLShortener{
		Urls:   make(map[string]string),
		Config: config,
	}
}

func AppendToMap(u *URLShortener, url []byte) ([]byte, error) {
	if shortURL, urlExist := checkForValue(url, u.Urls); urlExist {
		return shortURL, nil
	} else {
		shortURL, err := GenerateShortURL()
		if err != nil {
			return shortURL, err
		}
		stringShortURL := string(shortURL)
		u.Urls[stringShortURL] = string(url)
		return shortURL, nil
	}
}

func checkForValue(url []byte, urls map[string]string) ([]byte, bool) {
	var shortURL []byte
	for key, value := range urls {
		if value == string(url) {
			shortURL := []byte(key)
			return shortURL, true
		}
	}
	return shortURL, false
}

func GetFullURL(u *URLShortener, shortURL string) ([]byte, error) {
	var fullURLByte []byte
	if fullURL, found := u.Urls[shortURL]; found {
		fullURLByte := []byte(fullURL)
		return fullURLByte, nil
	} else {
		return fullURLByte, fmt.Errorf(urlNotFound)
	}
}
