package handlemetric

import (
	"fmt"
	"gmetrics/internal/metrics"
	"gmetrics/internal/payload"
	"strconv"
)

// updateMetricByStringValue updates the specified metric with the given value.
// It supports gauge and counter metric types.
//
// Parameters:
// - metricType: the type of the metric (either "gauge" or "counter")
// - metricName: the name of the metric
// - metricValue: the value of the metric
//
// Returns:
// - success message and nil error if the metric is updated successfully
// - empty string and UpdateMetricError if there is an error updating the metric
//
// UpdateMetricError is a custom error type that contains an error message and an HTTP status code.
//
// For gauge metrics, metricValue must be a valid float64. If it's not a valid float64,
// UpdateMetricError with the message "metric value is not a valid float" and
// an HTTP status code of http.StatusBadRequest will be returned.
//
// For counter metrics, metricValue must be a valid int64. If it's not a valid int64,
// UpdateMetricError with the message "metric value is not a valid int" and
// an HTTP status code of http.StatusBadRequest will be returned.
//
// If the metricType is neither "gauge" nor "counter", an UpdateMetricError
// with the message "invalid metric type" and an HTTP status code of http.StatusBadRequest will be returned.
func updateMetricByStringValue(metricType, metricName, metricValue string) (string, *UpdateMetricError) {
	switch metricType {
	case metrics.TypeGauge:
		convertedValue, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			//log.Println(err)
			return "", NotValidGaugeError
		}
		metrics.MeStore.SetGauge(metricName, metrics.Gauge(convertedValue))
		return fmt.Sprintf("metric %s successfully set", metricName), nil
	case metrics.TypeCounter:
		convertedValue, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			//log.Println(err)
			return "", NotValidCounterError
		}
		metrics.MeStore.AddCounter(metricName, metrics.Counter(convertedValue))
		return fmt.Sprintf("metric %s successfully add", metricName), nil
	default:
		return "", InvalidMetricTypeError
	}
}

// updateMetricByRequestBody updates the specified metric with the given body.
// It supports gauge and counter metric types.
//
// Parameters:
// - body: the request body containing the metric information
//
// Returns:
// - success message and nil error if the metric is updated successfully
// - empty string and UpdateMetricError if there is an error updating the metric
//
// UpdateMetricError is a custom error type that contains an error message and an HTTP status code.
func updateMetricByRequestBody(body payload.Metrics) (string, *UpdateMetricError) {
	if body.ID == "" {
		return "", BadRequestError
	}

	var responseMessage string
	switch body.MType {
	case metrics.TypeGauge:
		if body.Value == nil {
			return "", BadRequestError
		}
		metrics.MeStore.SetGauge(body.ID, metrics.Gauge(*body.Value))
		responseMessage = fmt.Sprintf("metric %s successfully set", body.ID)
	case metrics.TypeCounter:
		if body.Delta == nil {
			return "", BadRequestError
		}
		metrics.MeStore.AddCounter(body.ID, metrics.Counter(*body.Delta))
		responseMessage = fmt.Sprintf("metric %s successfully add", body.ID)
	default:
		return "", InvalidMetricTypeError
	}

	return responseMessage, nil
}
