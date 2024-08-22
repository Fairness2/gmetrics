package sendpool

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"gmetrics/cmd/agent/config"
	"gmetrics/internal/logger"
	"gmetrics/internal/payload"
	"sync"
)

// ErrorPoolIsClosed ошибка, что пул закрыт
var ErrorPoolIsClosed = errors.New("pool is closed")

// response структура ответа из горрутины
type response struct {
	Res *resty.Response
	Err error
}

// poolPayload структура тела для запроса на сервер
type poolPayload struct {
	Body []payload.Metrics
	Out  chan response // Канал с данными для обратной связи
}

// Pool пул отправщиков на сервер
type Pool struct {
	wg               sync.WaitGroup    // Группа ожидания для корректиного закрытия пула
	in               chan *poolPayload // Канал для отправки в горрутины
	isClosed         bool              // Флаг, что пул закрыт
	client           *resty.Client     // Клиент для подключения к серверам
	encodeWriterPool sync.Pool         // Шифровальщики тела
}

// New Создание нового пула отправщиков.
// Закрывается по завершению контекста
func New(ctx context.Context, size int) *Pool {
	in := make(chan *poolPayload, size)
	pool := &Pool{
		wg:     sync.WaitGroup{},
		in:     in,
		client: resty.New(),
		encodeWriterPool: sync.Pool{
			New: func() interface{} {
				writer, err := gzip.NewWriterLevel(nil, gzip.BestSpeed)
				if err != nil {
					logger.Log.Error(err)
					return nil
				}
				return writer
			},
		}}
	for i := 0; i < size; i++ {
		pool.wg.Add(1)
		go pool.worker(ctx)
	}
	// Запускаем следящую горрутину для закрытия пула
	go func() {
		<-ctx.Done()
		// ставим флаг закрытия
		pool.isClosed = true
		pool.wg.Wait()
		close(pool.in)
	}()

	return pool
}

// Send отправка метрик на сервер
func (p *Pool) Send(body []payload.Metrics) (*resty.Response, error) {
	if p.isClosed {
		return nil, ErrorPoolIsClosed
	}
	out := make(chan response)
	p.in <- &poolPayload{Body: body, Out: out}
	res := <-out

	return res.Res, res.Err
}

// worker горрутина для отправки на сервер метрик
func (p *Pool) worker(ctx context.Context) {
	defer p.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case poolPayload, ok := <-p.in:
			if !ok {
				return
			}
			p.processRequest(poolPayload)
		}
	}
}

// processRequest обработка запроса отправки
func (p *Pool) processRequest(body *poolPayload) {
	defer close(body.Out)
	res, err := p.sendToServer(body.Body)
	body.Out <- response{
		Res: res,
		Err: err,
	}
}

// sendToServer Отправка метрики
func (p *Pool) sendToServer(body []payload.Metrics) (*resty.Response, error) {
	logger.Log.Info("Sending metrics")
	// urlTemplate Шаблон урл: http://<АДРЕС_СЕРВЕРА>/update
	var urlUpdateTemplate = "%s/updates"
	sURL := fmt.Sprintf(urlUpdateTemplate, config.Params.ServerURL)
	client := p.client.R().SetHeader("Content-Type", "application/json")

	// Преобразуем тело в джейсон
	marshaledBody, err := p.marshalBody(body)
	if err != nil {
		return nil, err
	}

	// Сжимаем тело
	compressedBody, comressed, err := p.getBody(marshaledBody)
	if err != nil {
		return nil, err
	}
	client.SetBody(compressedBody)
	if comressed {
		client.SetHeader("Content-Encoding", "gzip")
	}

	// Устанавливаем подпись тела
	bodyHash, hashErr := p.hashBody(marshaledBody)
	if hashErr != nil {
		logger.Log.Error("Cant hash body", err)
	} else {
		client.SetHeader("HashSHA256", bodyHash)
	}

	// Отправляем запрос
	res, err := client.Post(sURL)
	logger.Log.Info("Finish sending metrics")

	return res, err
}

// marshalBody преобразует тело в строку JSON
func (p *Pool) marshalBody(body []payload.Metrics) ([]byte, error) {
	// Преобразовываем тело в строку джейсон
	return json.Marshal(body)
}

// getBody возвращает тело запроса, упаковывая входные данные в JSON, сжимая их с помощью gzip,
// если это возможно, и возвращая сжатое тело вместе с логическим значением, указывающим, было ли сжатие успешным.
// Если сжатие не удалось, функция регистрирует ошибку и возвращает несжатое тело.
// Функция также возвращает любую ошибку, возникшую во время упаковывания или сжатия.
// Возвращаемый [] byte содержит тело, логическое значение указывает,
// было ли сжатие успешным, а ошибка содержит любые обнаруженные ошибки.
func (p *Pool) getBody(body []byte) ([]byte, bool, error) {
	// Пробуем сжать тело
	compressedBody, err := p.compressBody(body)
	if err != nil {
		// Если не получилось, то ставим обычное боди
		logger.Log.Error("Cant compress body", err)
		return body, false, nil
	} else {
		// Если получилось, то ставим заголовок о методе кодировки и ставим закодированное тело
		return compressedBody, true, err
	}
}

// hashBody создаём подпись запроса
func (p *Pool) hashBody(body []byte) (string, error) {
	if config.Params.HashKey == "" {
		return "", errors.New("hash key is empty")
	}
	harsher := hmac.New(sha256.New, []byte(config.Params.HashKey))
	harsher.Write(body)
	return hex.EncodeToString(harsher.Sum(nil)), nil
}

// compressBody сжимаем данные в формат gzip
func (p *Pool) compressBody(body []byte) ([]byte, error) {
	// Создаём буфер, в который запишем сжатое тело
	var buf bytes.Buffer
	// Берём свободный врайтер для записи
	writer := p.encodeWriterPool.Get().(*gzip.Writer)
	// Если у нас он nil, то была ошибка
	if writer == nil {
		return nil, errors.New("writer is nil")
	}
	// Устанавливаем новый буфер во врайтер
	writer.Reset(&buf)
	// Кодируем данные
	_, err := writer.Write(body)
	if err != nil {
		return nil, err
	}
	// Делаем закрытие врайтера, чтобы он прописал все нужные байты в конце. После Reset он откроется снова (Проверял,ставил последовательно закрытие, ресет с новым буфером и снова запись)
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
