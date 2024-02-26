package config

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Config struct {
	ServerAddress string
	BaseURL       string
	FilePath      string
	DSN           string
}

var (
	config *Config
	once   sync.Once
)

func GetConfig() *Config {
	once.Do(
		func() {
			serverAddress := flag.String("a", "localhost:8080", "server start URL")
			baseURL := flag.String("b", "http://localhost:8080", "base of short URL")
			filePath := flag.String("f", "tmp/short-url-db.json", "file storage path")
			dsn := flag.String("d", "", "database adress")
			flag.Parse()

			config = &Config{
				ServerAddress: getServerAddress(serverAddress),
				BaseURL:       getBaseURL(baseURL),
				FilePath:      getFilePath(filePath),
				DSN:           getDatabaseDsn(dsn),
			}
		})

	return config
}

func getServerAddress(flagServerAddress *string) string {
	if envServerAddress, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		return envServerAddress
	}

	return *flagServerAddress
}

func getBaseURL(flagBaseURL *string) string {
	if envBaseURL, ok := os.LookupEnv("BASE_URL"); ok && checkBaseURL(envBaseURL) {
		return envBaseURL
	}

	return *flagBaseURL
}

func getFilePath(filePath *string) string {
	if envFilePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		return envFilePath
	}

	return *filePath
}

func getDatabaseDsn(databaseDsn *string) string {
	if envDatabaseDsn, ok := os.LookupEnv("DATABASE_DSN"); ok {
		return envDatabaseDsn
	}

	return *databaseDsn
}

func checkBaseURL(str string) bool {
	substr := strings.Split(str, ":")
	const (
		length     = 3
		isNotAForm = "%s is not in a form tls://host:port"
	)

	if len(substr) != length {
		log.Printf(isNotAForm, str)
		return false
	}

	if _, err := strconv.Atoi(substr[2]); err != nil {
		log.Printf("Port %s must have 4 digits", substr[1])
		return false
	}

	if idx := strings.Index(substr[1], "//"); idx < 0 {
		log.Printf(isNotAForm, str)
		return false
	}

	return true
}
