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

var t *template.Template // Массив для хранения шиблонов с их именами

//go:embed templates/*
var baseTemplate embed.FS

//go:embed templates/metrics.gohtml
var metricsTemplate string

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

type ShowedMetrics struct {
	Name  string
	Value string
}
