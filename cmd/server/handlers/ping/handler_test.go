package ping

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := []struct {
		name string
		db   func() IDB
		code int
	}{
		{
			name: "ping_context_returns_nil",
			db: func() IDB {
				db := NewMockIDB(ctrl)
				db.EXPECT().PingContext(gomock.Any()).Return(nil)
				return db
			},
			code: http.StatusOK,
		},
		{
			name: "ping_context_returns_error",
			db: func() IDB {
				db := NewMockIDB(ctrl)
				db.EXPECT().PingContext(gomock.Any()).Return(errors.New("an error occurred"))
				return db
			},
			code: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Post("/", NewController(tt.db()).Handler)
			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(router)

			request := resty.New().R()
			request.Method = http.MethodPost
			request.URL = srv.URL

			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			assert.Equal(t, tt.code, res.StatusCode(), "unexpected status code")
		})
	}
}
