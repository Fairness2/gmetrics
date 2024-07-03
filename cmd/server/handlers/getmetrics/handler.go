package getmetrics

import (
	"bytes"
	"gmetrics/internal/helpers"
	"gmetrics/internal/metrics"
	"html/template"
	"net/http"
)

var t map[string]*template.Template // Массив для хранения шиблонов с их именами

func init() {
	t = make(map[string]*template.Template)

	// Синтаксический разбор шаблона всегда в готовую переменную
	// Загрузка шаблонов вместе с основным шаблоном
	t["metrics.gohtml"] = template.Must(template.New("metrics.gohtml").ParseFiles(
		"cmd/server/web/templates/base.gohtml",
		"cmd/server/web/templates/metrics.gohtml"))
}

// Handler Возвращает страницу с метриками
//
// Parameters:
// - response: http.ResponseWriter объект, содержащий информацию о ответе HTTP.
// - request: http.Request объект, содержащий информацию о запросе HTTP.
func Handler(response http.ResponseWriter, request *http.Request) {
	gauges := metrics.MeStore.GetGauges()
	counters := metrics.MeStore.GetCounters()

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
	var buff bytes.Buffer                                           // Создание буфера для сохранения результата побработки шаблона
	err := t["metrics.gohtml"].ExecuteTemplate(&buff, "base", data) // Подключение шиблона к странице
	if err != nil {
		helpers.SetHTTPError(response, http.StatusInternalServerError, err.Error())
		return
	}
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Вывод буферизированного ответа
	_, err = response.Write(buff.Bytes())
	if err != nil {
		helpers.SetHTTPError(response, http.StatusInternalServerError, err.Error())
		return
	}
}

type ShowedMetrics struct {
	Name  string
	Value string
}
