package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func NewConfig() *Config {
	flagServerAddress := flag.String("a", "", "server start URL")
	flagBaseURL := flag.String("b", "", "base of short URL")
	flag.Parse()

	config := Config{
		ServerAddress: getServerAddress(flagServerAddress),
		BaseURL:       getBaseURL(flagBaseURL),
	}

	fmt.Printf("server start URL is %s\n", config.ServerAddress)
	fmt.Printf("base of short URL is %s\n", config.BaseURL)
	return &config
}

func getServerAddress(flagServerAddress *string) string {
	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" && checkServerAddress(envServerAddress) {
		return envServerAddress
	}

	if flagServerAddress != nil && checkServerAddress(*flagServerAddress) {
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

func checkServerAddress(str string) bool {

	substr := strings.Split(str, ":")
	if len(substr) != 2 {
		fmt.Printf("%s is not in a form host:port", str)
		return false
	}

	if _, err := strconv.Atoi(substr[1]); err != nil {
		fmt.Printf("Port %s must have 4 digits", substr[1])
		return false
	}

	return true
}

func checkBaseURL(str string) bool {

	substr := strings.Split(str, ":")
	if len(substr) != 3 {
		fmt.Printf("%s is not in a form tls://host:port", str)
		return false
	}

	if _, err := strconv.Atoi(substr[2]); err != nil {
		fmt.Printf("Port %s must have 4 digits", substr[1])
		return false
	}

	if idx := strings.Index(substr[1], "//"); idx < 0 {
		fmt.Printf("%s is not in a form tls://host:port", str)
		return false
	}

	return true
}
