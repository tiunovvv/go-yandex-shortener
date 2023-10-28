package shortener

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"
)

const (
	charset        = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	shortUrlLength = 8
	schemePrefix   = "://"
	urlNotFound    = "URL not found"
)

type URLShortener struct {
	Urls map[string]string
}

func GenerateShortUrl() ([]byte, error) {
	ret := make([]byte, shortUrlLength)
	for i := 0; i < shortUrlLength; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return ret, err
		}
		ret[i] = charset[num.Int64()]
	}
	return ret, nil
}

func CreateUrlMap() *URLShortener {
	return &URLShortener{
		Urls: make(map[string]string),
	}
}

func AppendToMap(u *URLShortener, bodyUrl *url.URL) ([]byte, error) {
	url := []byte(bodyUrl.Scheme + schemePrefix + bodyUrl.Host + bodyUrl.RequestURI())
	if shortUrl, urlExist := checkForValue(url, u.Urls); urlExist {
		return shortUrl, nil
	} else {
		shortUrl, err := GenerateShortUrl()
		if err != nil {
			return shortUrl, err
		}
		stringShortUrl := string(shortUrl)
		u.Urls[stringShortUrl] = string(url)
		return shortUrl, nil
	}
}

func checkForValue(url []byte, urls map[string]string) ([]byte, bool) {
	var shortUrl []byte
	for key, value := range urls {
		if value == string(url) {
			shortUrl := []byte(key)
			return shortUrl, true
		}
	}
	return shortUrl, false
}

func GetFullUrl(u *URLShortener, shortUrl string) ([]byte, error) {
	var fullUrlByte []byte
	if fullUrl, found := u.Urls[shortUrl]; found {
		fullUrlByte := []byte(fullUrl)
		return fullUrlByte, nil
	} else {
		return fullUrlByte, fmt.Errorf(urlNotFound)
	}
}
