package config

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/tiunovvv/go-yandex-shortener/internal/logger"
)

type Config struct {
	logger          *logger.Logger
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
}

func NewConfig(logger *logger.Logger) *Config {
	flagServerAddress := flag.String("a", "", "server start URL")
	flagBaseURL := flag.String("b", "", "base of short URL")
	fileStoragePath := flag.String("f", "", "file storage path")
	flag.Parse()

	config := Config{
		logger:          logger,
		ServerAddress:   getServerAddress(flagServerAddress),
		BaseURL:         getBaseURL(flagBaseURL),
		FileStoragePath: getFileStoragePath(fileStoragePath),
	}

	logger.Sugar().Infof("server start URL is %s", config.ServerAddress)
	logger.Sugar().Infof("base of short URL is %s", config.BaseURL)
	if config.FileStoragePath == "" {
		logger.Sugar().Info("file storage path is empty, disk recording is disabled")
	}
	if config.FileStoragePath != "" {
		logger.Sugar().Infof("file storage path is %s", config.FileStoragePath)
	}

	return &config
}

func getServerAddress(flagServerAddress *string) string {
	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		return envServerAddress
	}

	if flagServerAddress != nil && *flagServerAddress != "" {
		return *flagServerAddress
	}

	return "localhost:8080"
}

func getBaseURL(flagBaseURL *string) string {
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" && checkBaseURL(envBaseURL) {
		return envBaseURL
	}

	if flagBaseURL != nil && checkBaseURL(*flagBaseURL) {
		return *flagBaseURL
	}

	return "http://localhost:8080"
}

func getFileStoragePath(fileStoragePath *string) string {
	if envfileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envfileStoragePath != "" {
		return envfileStoragePath
	}

	if fileStoragePath != nil && *fileStoragePath != "" {
		return *fileStoragePath
	}

	return ""
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
