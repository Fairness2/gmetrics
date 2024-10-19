package getmetrics

import (
	"bytes"
	"embed"
	_ "embed"
	"gmetrics/internal/helpers"
	"gmetrics/internal/logger"
	"gmetrics/internal/metrics"
	"html/template"
	"net/http"
)

// t Массив для хранения шиблонов с их именами
var t *template.Template

// baseTemplate содержит встроенную файловую систему со всеми файлами шаблонов, расположенными в каталоге шаблонов.
//
//go:embed templates/*
var baseTemplate embed.FS

// init инициализирует систему шаблонов приложения, анализируя и загружая шаблоны в глобальную переменную шаблона.
func init() {

	// Синтаксический разбор шаблона всегда в готовую переменную
	// Загрузка шаблонов вместе с основным шаблоном
	t = template.Must(template.New("metrics.gohtml").ParseFS(baseTemplate, "templates/base.gohtml", "templates/metrics.gohtml"))
}

// Handler Возвращает страницу с метриками
//
// Parameters:
// - response: http.ResponseWriter объект, содержащий информацию о ответе HTTP.
// - request: http.Request объект, содержащий информацию о запросе HTTP.
//
// @Summary	  Возвращает страницу с метриками
// @Description  Возвращает страницу с метриками
// @Tags		 Метрики
// @Produce	  html
// @Success	  200  {object}  string  "Metrics page"
// @Failure	  500  {object}  string  "Internal Server Error"
// @Router / [get]
func Handler(response http.ResponseWriter, request *http.Request) {
	gauges, errGauge := metrics.MeStore.GetGauges()
	if errGauge != nil {
		logger.Log.Error(errGauge)
	}
	counters, errCounter := metrics.MeStore.GetCounters()
	if errGauge != nil {
		logger.Log.Error(errCounter)
	}

	gaugeList := make([]ShowedMetrics, 0, len(gauges))
	for name, value := range gauges {
		gaugeList = append(gaugeList, ShowedMetrics{
			Name:  name,
			Value: value.ToString(),
		})
	}
	counterList := make([]ShowedMetrics, 0, len(counters))
	for name, value := range counters {
		counterList = append(counterList, ShowedMetrics{
			Name:  name,
			Value: value.ToString(),
		})
	}
	data := struct {
		GaugeList   []ShowedMetrics
		CounterList []ShowedMetrics
	}{
		GaugeList:   gaugeList,
		CounterList: counterList,
	}
	var buff bytes.Buffer                         // Создание буфера для сохранения результата побработки шаблона
	err := t.ExecuteTemplate(&buff, "base", data) // Подключение шиблона к странице
	if err != nil {
		helpers.SetHTTPResponse(response, http.StatusInternalServerError, []byte(err.Error()))
		return
	}
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Вывод буферизированного ответа
	_, err = response.Write(buff.Bytes())
	if err != nil {
		helpers.SetHTTPResponse(response, http.StatusInternalServerError, []byte(err.Error()))
		return
	}
}

// ShowedMetrics представляет собой структурированную метрику с именем и значением, которые будут отображаться в пользовательском интерфейсе.
type ShowedMetrics struct {
	Name  string
	Value string
}
