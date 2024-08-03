package ping

import (
	"context"
	"gmetrics/internal/database"
	"gmetrics/internal/helpers"
	"gmetrics/internal/logger"
	"net/http"
	"time"
)

// Handler возвращает состояние коннекта к базе данных
func Handler(response http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Проверяем соединение с бд
	err := database.DB.PingContext(ctx)
	if err != nil {
		logger.Log.Error(err)
		helpers.SetHTTPError(response, http.StatusInternalServerError, []byte{})
		return
	}
	helpers.SetHTTPError(response, http.StatusOK, []byte{})
}
