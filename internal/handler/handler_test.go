package handler

// func TestPostHandler(t *testing.T) {
// 	type post struct {
// 		request string
// 		body    string
// 	}
// 	type want struct {
// 		statusCode int
// 		body       string
// 	}
// 	tests := []struct {
// 		name string
// 		post post
// 		want want
// 	}{
// 		{
// 			name: "positive test:short body",
// 			post: post{
// 				request: "http://localhost:8080/",
// 				body:    "http://www.yandex.ru",
// 			},
// 			want: want{
// 				statusCode: 201,
// 				body:       "http://localhost:8080/",
// 			},
// 		},
// 		{
// 			name: "negativ test:initial body",
// 			post: post{
// 				request: "http://localhost:8080/",
// 				body:    "",
// 			},
// 			want: want{
// 				statusCode: 400,
// 				body:       "",
// 			},
// 		},
// 		{
// 			name: "negativ test:body is not url",
// 			post: post{
// 				request: "http://localhost:8080/",
// 				body:    "12345",
// 			},
// 			want: want{
// 				statusCode: 500,
// 				body:       "",
// 			},
// 		},
// 	}

// 	config := &config.Config{
// 		BaseURL:       "http://localhost:8080/",
// 		ServerAddress: "localhost:8080",
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			request := httptest.NewRequest(http.MethodPost, tt.post.request, bytes.NewReader([]byte(tt.post.body)))
// 			w := httptest.NewRecorder()

// 			storage := storage.NewStorage(config)
// 			shortener := shortener.NewShortener(storage)
// 			handler := NewHandler(shortener)

// 			router := handler.InitRoutes()
// 			router.ServeHTTP(w, request)
// 			result := w.Result()

// 			assert.Equal(t, tt.want.statusCode, result.StatusCode)

// 			body, err := io.ReadAll(result.Body)
// 			require.NoError(t, err)
// 			err = result.Body.Close()
// 			require.NoError(t, err)
// 			assert.Contains(t, string(body), tt.want.body)
// 		})
// 	}
// }

// func TestGetHandler(t *testing.T) {
// 	type want struct {
// 		statusCode int
// 		location   string
// 	}
// 	tests := []struct {
// 		name     string
// 		request  string
// 		mapKey   string
// 		mapValue string
// 		want     want
// 	}{
// 		{
// 			name:     "positive test #1",
// 			mapKey:   "OWjwkttu",
// 			mapValue: "http://www.yandex.ru",
// 			request:  "http://localhost:8080/OWjwkttu",
// 			want: want{
// 				statusCode: 307,
// 				location:   "http://www.yandex.ru",
// 			},
// 		},
// 		{
// 			name:     "negativ test: initial shortURL",
// 			mapKey:   "OWjwkttu",
// 			mapValue: "http://www.yandex.ru",
// 			request:  "http://localhost:8080/",
// 			want: want{
// 				statusCode: 404,
// 				location:   "",
// 			},
// 		},
// 		{
// 			name:     "negativ test: shortURL doesn't exist",
// 			mapKey:   "OWjwkttu",
// 			mapValue: "http://www.yandex.ru",
// 			request:  "http://localhost:8080/123",
// 			want: want{
// 				statusCode: 400,
// 				location:   "",
// 			},
// 		},
// 	}

// 	config := &config.Config{
// 		BaseURL:       "http://localhost:8080/",
// 		ServerAddress: "localhost:8080",
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			request := httptest.NewRequest(http.MethodGet, tt.request, nil)

// 			w := httptest.NewRecorder()
// 			storage := storage.NewStorage(config)

// 			storage.Urls[tt.mapKey] = tt.mapValue
// 			shortener := shortener.NewShortener(storage)
// 			handler := NewHandler(shortener)
// 			router := handler.InitRoutes()

// 			router.ServeHTTP(w, request)
// 			result := w.Result()

// 			_, err := io.ReadAll(result.Body)
// 			require.NoError(t, err)
// 			err = result.Body.Close()
// 			require.NoError(t, err)

// 			assert.Equal(t, tt.want.statusCode, result.StatusCode)
// 			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
// 		})
// 	}
// }