package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tiunovvv/go-yandex-shortener/cmd/config"
	"github.com/tiunovvv/go-yandex-shortener/pkg/storage"
)

func TestPostHandler(t *testing.T) {
	type post struct {
		request string
		body    string
	}
	type want struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name string
		post post
		want want
	}{
		{
			name: "positive test:short body",
			post: post{
				request: "http://localhost:8080/",
				body:    "http://www.yandex.ru",
			},
			want: want{
				statusCode: 201,
				body:       "http://localhost:8080/",
			},
		},
		{
			name: "positive test:very long body",
			post: post{
				request: "http://localhost:8080/",
				body:    "https://www.google.com/search?q=golang+assert+string+contains+string&oq=golang+assert+string+contains+string&gs_lcrp=EgZjaHJvbWUyBggAEEUYOTIKCAEQIRgWGB0YHjIKCAIQIRgWGB0YHtIBCDkyNzVqMGo0qAIAsAIA&sourceid=chrome&ie=UTF-8",
			},
			want: want{
				statusCode: 201,
				body:       "http://localhost:8080/",
			},
		},
		{
			name: "negativ test:initial body",
			post: post{
				request: "http://localhost:8080/",
				body:    "",
			},
			want: want{
				statusCode: 400,
				body:       "",
			},
		},
		{
			name: "negativ test:body is not url",
			post: post{
				request: "http://localhost:8080/",
				body:    "12345",
			},
			want: want{
				statusCode: 400,
				body:       "",
			},
		},
	}

	serverStartURL := config.ServerStartURL{"localhost", 8080}
	baseShortURL := config.BaseShortURL{"http", "localhost", 8080}
	config := config.Config{&serverStartURL, &baseShortURL}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.post.request, bytes.NewReader([]byte(tt.post.body)))
			w := httptest.NewRecorder()

			storage := storage.CreateStorage(&config)
			handler := NewHandler(storage)
			router := handler.InitRoutes()
			router.ServeHTTP(w, request)
			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)
			assert.Contains(t, string(body), tt.want.body)
		})
	}
}

func TestGetHandler(t *testing.T) {
	type want struct {
		statusCode int
		location   string
	}
	tests := []struct {
		name     string
		request  string
		mapKey   string
		mapValue string
		want     want
	}{
		{
			name:     "positive test #1",
			mapKey:   "OWjwkttu",
			mapValue: "http://www.yandex.ru",
			request:  "http://localhost:8080/OWjwkttu",
			want: want{
				statusCode: 307,
				location:   "http://www.yandex.ru",
			},
		},
		{
			name:     "negativ test: initial shortURL",
			mapKey:   "OWjwkttu",
			mapValue: "http://www.yandex.ru",
			request:  "http://localhost:8080/",
			want: want{
				statusCode: 404,
				location:   "",
			},
		},
		{
			name:     "negativ test: shortURL doesn't exist",
			mapKey:   "OWjwkttu",
			mapValue: "http://www.yandex.ru",
			request:  "http://localhost:8080/123",
			want: want{
				statusCode: 400,
				location:   "",
			},
		},
	}

	serverStartURL := config.ServerStartURL{"localhost", 8080}
	baseShortURL := config.BaseShortURL{"http", "localhost", 8080}
	config := config.Config{&serverStartURL, &baseShortURL}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			request := httptest.NewRequest(http.MethodGet, tt.request, nil)

			w := httptest.NewRecorder()

			storage := storage.CreateStorage(&config)
			storage.Urls[tt.mapKey] = tt.mapValue

			handler := NewHandler(storage)
			router := handler.InitRoutes()

			router.ServeHTTP(w, request)
			result := w.Result()

			_, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}
