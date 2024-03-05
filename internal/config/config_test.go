package config

import (
	"flag"
	"testing"
)

// TestGetBaseURL tests the getBaseURL function.
func TestGetBaseURL(t *testing.T) {
	t.Setenv("BASE_URL", "https://example.com:8080")
	baseURL := getBaseURL(flag.String("b", "http://localhost:8080", ""))

	if baseURL != "https://example.com:8080" {
		t.Errorf("Expected base URL to be https://example.com:8080, got %s", baseURL)
	}
}

// TestCheckBaseURL tests the checkBaseURL function.
func TestCheckBaseURL(t *testing.T) {
	validURL := "tls://example.com:8080"
	if !checkBaseURL(validURL) {
		t.Errorf("Expected %s to be a valid base URL", validURL)
	}
}

// TestGetFilePath tests the getFilePath function.
func TestGetFilePath(t *testing.T) {
	t.Setenv("FILE_STORAGE_PATH", "tmp/short-url-db.json")
	filePath := getFilePath(flag.String("f", "tmp/short-url-db.json", ""))

	if filePath != "tmp/short-url-db.json" {
		t.Errorf("Expected file path to be tmp/short-url-db.json, got %s", filePath)
	}
}

// TestGetDatabaseDsn tests the getDatabaseDsn function.
func TestGetDatabaseDsn(t *testing.T) {
	t.Setenv("DATABASE_DSN", "example-database-dsn")
	dsn := getDatabaseDsn(flag.String("d", "", ""))

	if dsn != "example-database-dsn" {
		t.Errorf("Expected database DSN to be example-database-dsn, got %s", dsn)
	}
}
