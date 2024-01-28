package config

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type Config struct {
	logger        *zap.Logger
	ServerAddress string
	BaseURL       string
	FilePath      string
	DSN           string
}

func NewConfig(logger *zap.Logger) *Config {
	serverAddress := flag.String("a", "localhost:8080", "server start URL")
	baseURL := flag.String("b", "http://localhost:8080", "base of short URL")
	filePath := flag.String("f", "tmp/short-url-db.json", "file storage path")
	dsn := flag.String("d", "", "db adress")
	flag.Parse()

	config := Config{
		logger:        logger,
		ServerAddress: getServerAddress(serverAddress),
		BaseURL:       getBaseURL(baseURL),
		FilePath:      getFilePath(filePath),
		DSN:           getDatabaseDsn(dsn),
	}

	logger.Sugar().Infof("server start URL: %s", config.ServerAddress)
	logger.Sugar().Infof("base of short URL: %s", config.BaseURL)
	if config.FilePath == "" {
		logger.Sugar().Info("file storage path is empty, disk recording is disabled")
	}
	if config.FilePath != "" {
		logger.Sugar().Infof("file storage path: %s", config.FilePath)
	}
	logger.Sugar().Infof("database connection address: %s", config.DSN)

	return &config
}

func getServerAddress(flagServerAddress *string) string {
	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		return envServerAddress
	}

	return *flagServerAddress
}

func getBaseURL(flagBaseURL *string) string {
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" && checkBaseURL(envBaseURL) {
		return envBaseURL
	}

	return *flagBaseURL
}

func getFilePath(filePath *string) string {
	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		return envFilePath
	}

	return *filePath
}

func getDatabaseDsn(databaseDsn *string) string {
	if envDatabaseDsn := os.Getenv("DATABASE_DSN"); envDatabaseDsn != "" {
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
