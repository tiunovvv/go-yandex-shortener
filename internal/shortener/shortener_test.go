package shortener

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	myErrors "github.com/tiunovvv/go-yandex-shortener/internal/errors"
	"github.com/tiunovvv/go-yandex-shortener/internal/models"
	"github.com/tiunovvv/go-yandex-shortener/internal/storage/mocks"
	"go.uber.org/zap"
)

// TestGetShortURL is a test function for the GetShortURL method.
func TestGetShortURL(t *testing.T) {
	type fields struct {
		store *mocks.MockStore
		log   *zap.SugaredLogger
	}
	type args struct {
		fullURL string
		userID  string
	}
	tests := []struct {
		name     string
		prepare  func(f *fields)
		args     args
		shortURL string
		err      error
	}{
		{
			name: "successful save",
			prepare: func(f *fields) {
				f.store.EXPECT().SaveURL(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any()).Return(nil)
			},
			args: args{
				fullURL: "http://axIsW.NAPvB",
				userID:  "KCcJLOWm",
			},
			shortURL: "",
			err:      nil,
		},
		{
			name: "URL already saved",
			prepare: func(f *fields) {
				gomock.InOrder(
					f.store.EXPECT().SaveURL(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any()).Return(myErrors.ErrURLAlreadySaved),

					f.store.EXPECT().GetShortURL(
						gomock.Any(),
						gomock.Any()).Return("QWERASDF"),
				)
			},
			args: args{
				fullURL: "http://axIsW.NAPvB",
				userID:  "KCcJLOWm",
			},
			shortURL: "QWERASDF",
			err:      myErrors.ErrURLAlreadySaved,
		},
		{
			name: "URL already exists",
			prepare: func(f *fields) {
				gomock.InOrder(
					f.store.EXPECT().SaveURL(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any()).Return(myErrors.ErrKeyAlreadyExists),

					f.store.EXPECT().SaveURL(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any()).Return(nil),
				)
			},
			args: args{
				fullURL: "http://axIsW.NAPvB",
				userID:  "KCcJLOWm",
			},
			shortURL: "",
			err:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			f := fields{
				store: mocks.NewMockStore(ctrl),
				log:   zap.NewNop().Sugar(),
			}

			if tt.prepare != nil {
				tt.prepare(&f)
			}

			sh := &Shortener{
				store: f.store,
				log:   f.log,
			}

			shortURL, err := sh.GetShortURL(context.Background(), tt.args.fullURL, tt.args.userID)
			assert.Equal(t, err, tt.err)
			if len(tt.shortURL) != 0 {
				assert.Equal(t, shortURL, tt.shortURL)
			}
		})
	}
}

// TestGetShortURLBatch is a test function for the GetShortURLBatch method.
func TestGetShortURLBatch(t *testing.T) {
	type fields struct {
		store *mocks.MockStore
		log   *zap.SugaredLogger
	}
	type args struct {
		reqSlice []models.ReqAPIBatch
		userID   string
	}
	tests := []struct {
		name    string
		prepare func(f *fields)
		args    args
		err     error
	}{
		{
			name: "successful get",
			prepare: func(f *fields) {
				f.store.EXPECT().SaveURLBatch(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			err: nil,
		},
		{
			name: "unsuccessful get",
			prepare: func(f *fields) {
				f.store.EXPECT().SaveURLBatch(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("error"))
			},
			err: fmt.Errorf("error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			f := fields{
				store: mocks.NewMockStore(ctrl),
				log:   zap.NewNop().Sugar(),
			}

			if tt.prepare != nil {
				tt.prepare(&f)
			}

			sh := &Shortener{
				store: f.store,
				log:   f.log,
			}

			_, err := sh.GetShortURLBatch(context.Background(), tt.args.reqSlice, tt.args.userID)
			if tt.err == nil {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGetFullURL is a test function for the GetFullURL method.
func TestGetFullURL(t *testing.T) {
	type fields struct {
		store *mocks.MockStore
		log   *zap.SugaredLogger
	}
	type args struct {
		shortURL string
	}
	tests := []struct {
		name    string
		prepare func(f *fields)
		args    args
		fullURL string
		delete  bool
		err     error
	}{
		{
			name: "successful get",
			prepare: func(f *fields) {
				f.store.EXPECT().GetFullURL(gomock.Any(), gomock.Any()).Return("http://axIsW.NAPvB", false, nil)
			},
			args: args{
				shortURL: "QWERASDF",
			},
			fullURL: "http://axIsW.NAPvB",
			delete:  false,
		},
		{
			name: "unsuccessful get",
			prepare: func(f *fields) {
				f.store.EXPECT().GetFullURL(gomock.Any(), gomock.Any()).Return("", false, nil)
			},
			args: args{
				shortURL: "QWERASDF",
			},
			fullURL: "",
			delete:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			f := fields{
				store: mocks.NewMockStore(ctrl),
				log:   zap.NewNop().Sugar(),
			}

			if tt.prepare != nil {
				tt.prepare(&f)
			}

			sh := &Shortener{
				store: f.store,
				log:   f.log,
			}

			fullURL, deleteFlag, err := sh.GetFullURL(context.Background(), tt.args.shortURL)
			assert.Equal(t, fullURL, tt.fullURL)
			assert.Equal(t, deleteFlag, tt.delete)
			if len(tt.fullURL) != 0 {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGetURLByUserID description of the Go function.
func TestGetURLByUserID(t *testing.T) {
	type fields struct {
		store *mocks.MockStore
		log   *zap.SugaredLogger
	}
	type args struct {
		baseURL string
		userID  string
	}
	tests := []struct {
		name    string
		prepare func(f *fields)
		args    args
		urls    []models.UsersURLs
	}{
		{
			name: "successful get",
			prepare: func(f *fields) {
				f.store.EXPECT().GetURLByUserID(gomock.Any(), gomock.Any()).Return(map[string]string{
					"KCcJLOWm": "http://axIsW.NAPvB",
					"yetoic5G": "http://wJUDw.rHNqd"})
			},
			args: args{
				baseURL: "http://localhost:8080",
				userID:  "KCcJLOWm",
			},
			urls: []models.UsersURLs{
				{ShortURL: "http://localhost:8080/KCcJLOWm", OriginalURL: "http://axIsW.NAPvB"},
				{ShortURL: "http://localhost:8080/yetoic5G", OriginalURL: "http://wJUDw.rHNqd"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			f := fields{
				store: mocks.NewMockStore(ctrl),
				log:   zap.NewNop().Sugar(),
			}

			if tt.prepare != nil {
				tt.prepare(&f)
			}

			sh := &Shortener{
				store: f.store,
				log:   f.log,
			}

			urls := sh.GetURLByUserID(context.Background(), tt.args.baseURL, tt.args.userID)
			assert.Equal(t, urls, tt.urls)
		})
	}
}

func TestCheckConnect(t *testing.T) {
	type fields struct {
		store *mocks.MockStore
		log   *zap.SugaredLogger
	}
	tests := []struct {
		name    string
		prepare func(f *fields)
		err     error
	}{
		{
			name:    "successful get",
			prepare: func(f *fields) { f.store.EXPECT().GetPing(gomock.Any()).Return(nil) },
			err:     nil,
		},
		{
			name:    "unsuccessful get",
			prepare: func(f *fields) { f.store.EXPECT().GetPing(gomock.Any()).Return(fmt.Errorf("error")) },
			err:     fmt.Errorf("error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			f := fields{
				store: mocks.NewMockStore(ctrl),
				log:   zap.NewNop().Sugar(),
			}

			if tt.prepare != nil {
				tt.prepare(&f)
			}

			sh := &Shortener{
				store: f.store,
				log:   f.log,
			}

			err := sh.CheckConnect(context.Background())
			if tt.err != nil {
				assert.Error(t, err)
			}
		})
	}
}
