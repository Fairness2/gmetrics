package ping

import (
	"context"
	"gmetrics/internal/helpers"
	"gmetrics/internal/logger"
	"net/http"
	"time"
)

// IDB определяет интерфейс для операций с базой данных, включая проверку подключения.
type IDB interface {
	PingContext(ctx context.Context) error
}

// Controller отвечает за обработку запросов и взаимодействие с базой данных через интерфейс IDB.
type Controller struct {
	db IDB
}

// NewController инициализирует новый экземпляр контроллера с реализацией IDB.
func NewController(db IDB) *Controller {
	return &Controller{
		db: db,
	}
}

// Handler возвращает состояние коннекта к базе данных
//
// @Summary	  Проверка состояния подключения к базе данных
// @Description  Проверяет, установлено ли соединение с базой данных и возвращает соответствующий статус код.
// @Tags		 Пинг
// @Success	  200  {string}  "OK"
// @Failure	  500  {string}  "Internal Server Error"
// @Router	   /ping [get]
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
