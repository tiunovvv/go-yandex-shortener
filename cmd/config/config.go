package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	*ServerStartURL
	*BaseShortURL
}

type ServerStartURL struct {
	Host string
	Port int
}

func (s *ServerStartURL) Set(value string) error {
	hp := strings.Split(value, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}

	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}

	s.Host = hp[0]
	s.Port = port
	return nil
}

func (s *ServerStartURL) String() string {
	return s.Host + ":" + strconv.Itoa(s.Port)
}

type BaseShortURL struct {
	TLS  string
	Host string
	Port int
}

func (b *BaseShortURL) Set(value string) error {
	hp := strings.Split(value, ":")
	if len(hp) != 3 {
		return errors.New("need address in a form tls://host:port")
	}

	port, err := strconv.Atoi(hp[2])
	if err != nil {
		return err
	}

	idx := strings.Index(hp[1], "//")
	if idx < 0 {
		return fmt.Errorf("need address in a form tls://host:port")
	}

	b.TLS = hp[0]
	b.Host = hp[1]
	b.Host = strings.ReplaceAll(b.Host, "//", "")
	b.Port = port
	return nil
}

func (b *BaseShortURL) String() string {
	return b.TLS + "://" + b.Host + ":" + strconv.Itoa(b.Port)
}

func InitConfig() *Config {
	serverStartURL := ServerStartURL{"localhost", 8080}
	baseShortURL := BaseShortURL{"http", "localhost", 8080}

	flag.Var(&serverStartURL, "a", "server start URL")
	flag.Var(&baseShortURL, "b", "base of short URL")

	flag.Parse()

	envServerAddress := os.Getenv("SERVER_ADDRESS")
	if envServerAddress != "" {
		serverStartURL.Set(envServerAddress)
	}

	envBaseShortURL := os.Getenv("BASE_URL")

	if envBaseShortURL != "" {
		baseShortURL.Set(envBaseShortURL)
	}

	return &Config{
		&serverStartURL,
		&baseShortURL,
	}
}
