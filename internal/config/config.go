package config

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Config struct.
type Config struct {
	ServerAddress string
	BaseURL       string
	FilePath      string
	DSN           string
}

// Varables for singletons.
var (
	config *Config
	once   sync.Once
)

// GetConfig singleton returns the config.
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

// getServerAddress returns the server address from environment variable.
func getServerAddress(flagServerAddress *string) string {
	if envServerAddress, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		return envServerAddress
	}

	return *flagServerAddress
}

// getBaseURL returns the base URL from environment variable.
func getBaseURL(flagBaseURL *string) string {
	if envBaseURL, ok := os.LookupEnv("BASE_URL"); ok && checkBaseURL(envBaseURL) {
		return envBaseURL
	}

	return *flagBaseURL
}

// getFilePath returns the file path from environment variable.
func getFilePath(filePath *string) string {
	if envFilePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		return envFilePath
	}

	return *filePath
}

// getDatabaseDsn returns the database DSN from environment variable.
func getDatabaseDsn(databaseDsn *string) string {
	if envDatabaseDsn, ok := os.LookupEnv("DATABASE_DSN"); ok {
		return envDatabaseDsn
	}

	return *databaseDsn
}

// checkBaseURL checks if the base URL is in a form tls://host:port.
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
