package ping

import (
	"context"
	"gmetrics/internal/helpers"
	"gmetrics/internal/logger"
	"net/http"
	"time"
)

type IDB interface {
	PingContext(ctx context.Context) error
}

type Controller struct {
	db IDB
}

func NewController(db IDB) *Controller {
	return &Controller{
		db: db,
	}
}

// Handler возвращает состояние коннекта к базе данных
func (c *Controller) Handler(response http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Проверяем соединение с бд
	err := c.db.PingContext(ctx)
	if err != nil {
		logger.Log.Error(err)
		helpers.SetHTTPResponse(response, http.StatusInternalServerError, []byte{})
		return
	}
	helpers.SetHTTPResponse(response, http.StatusOK, []byte{})
}
