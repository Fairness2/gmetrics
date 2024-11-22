package sendpool

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"gmetrics/internal/encrypt"
	"gmetrics/internal/logger"
	"gmetrics/internal/payload"
	"io"
	"sync"
)

var (
	// ErrorPoolIsClosed ошибка, что пул закрыт
	ErrorPoolIsClosed = errors.New("pool is closed")
	// ErrorEmptyHashKey указывает, что хэш-ключ пуст при попытке хэширования тела.
	ErrorEmptyHashKey = errors.New("hash key is empty")
	// ErrorWrongWorkerSize указывает, что количество воркеров в пуле подобрано не верно, возможно меньше нуля
	ErrorWrongWorkerSize = errors.New("wrong worker size")
	// ErrorEmptyClient Пул передан пустой http клиент
	ErrorEmptyClient = errors.New("empty client")

	// ErrorServerURLIsEmpty Переданные адрес сервера пустой
	ErrorServerURLIsEmpty = errors.New("server url is empty")

	// ErrorCantCompressBody Ошибка, что мы не смогли сжать тело
	ErrorCantCompressBody = errors.New("cant compress body")

	// ErrorCantEcryptBody Ошибка, что мы не смогли зашифровать тело
	ErrorCantEcryptBody = errors.New("cant encrypt body")
)

var (
	URLUpdates = "/updates" // адрес обновления метрик
)

// IClient Клиент для отправки метрик на сервер
type IClient interface {
	Post(url string, body []byte, headers ...Header) (MetricResponse, error)
	EnableManualCompression() bool
}

// response структура ответа из горрутины
type response struct {
	Res MetricResponse
	Err error
}

// MetricResponse интерфейс для ответов от сервера
type MetricResponse interface {
	StatusCode() int
}

// poolPayload структура тела для запроса на сервер
type poolPayload struct {
	Out  chan response // Канал с данными для обратной связи
	Body []payload.Metrics
}

// bodyPipe функция для преобразования тела запроса
type bodyPipe func(body []byte) ([]byte, error)

// Pool пул отправщиков на сервер
type Pool struct {
	client           IClient           // Клиент для подключения к серверам
	encodeWriterPool sync.Pool         // Шифровальщики тела
	in               chan *poolPayload // Канал для отправки в горрутины
	wg               sync.WaitGroup    // Группа ожидания для корректиного закрытия пула
	HashKey          string
	isClosed         bool           // Флаг, что пул закрыт
	publicKey        *rsa.PublicKey // Ключ для шифрования тела запроса к серверу
}

// newEncoder создает и возвращает новый модуль записи gzip с лучшим уровнем скорости сжатия.
func newEncoder() interface{} {
	writer, err := gzip.NewWriterLevel(nil, gzip.BestSpeed)
	if err != nil {
		logger.Log.Error(err)
		return nil
	}
	return writer
}

// New Создание нового пула отправщиков.
// Закрывается по завершению контекста
func New(ctx context.Context, size int, HashKey, ServerURL string, publicKey *rsa.PublicKey) (*Pool, error) {
	if ServerURL == "" {
		return nil, ErrorServerURLIsEmpty
	}
	client, err := NewRestClient(ServerURL)
	if err != nil {
		return nil, err
	}
	return NewWithClient(ctx, size, HashKey, client, publicKey)
}

// NewWithRPC Создание нового пула отправщиков c rpc клиентом.
// Закрывается по завершению контекста
func NewWithRPC(ctx context.Context, size int, HashKey, ServerURL string, publicKey *rsa.PublicKey) (*Pool, error) {
	if ServerURL == "" {
		return nil, ErrorServerURLIsEmpty
	}
	client, err := NewRPCClient(ctx, ServerURL)
	if err != nil {
		return nil, err
	}
	return NewWithClient(ctx, size, HashKey, client, publicKey)
}

// NewWithClient инициализирует новый пул с заданным размером, хеш-ключом и rest-клиентом и запускает рабочие горутины.
func NewWithClient(ctx context.Context, size int, HashKey string, client IClient, publicKey *rsa.PublicKey) (*Pool, error) {
	if size <= 0 {
		return nil, ErrorWrongWorkerSize
	}
	if HashKey == "" {
		return nil, ErrorEmptyHashKey
	}
	if client == nil {
		return nil, ErrorEmptyClient
	}
	in := make(chan *poolPayload, size)
	pool := &Pool{
		wg:     sync.WaitGroup{},
		in:     in,
		client: client,
		encodeWriterPool: sync.Pool{
			New: newEncoder,
		},
		HashKey:   HashKey,
		publicKey: publicKey,
	}
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
		if cc, ok := pool.client.(io.Closer); ok {
			if cErr := cc.Close(); cErr != nil {
				logger.Log.Error(cErr)
			}
		}
	}()

	return pool, nil
}

// Send отправка метрик на сервер
func (p *Pool) Send(body []payload.Metrics) (MetricResponse, error) {
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
func (p *Pool) sendToServer(body []payload.Metrics) (MetricResponse, error) {
	logger.Log.Info("Sending metrics")
	headers := make([]Header, 0, 4)
	headers = append(headers, Header{
		Name:  "Content-Type",
		Value: "application/json",
	})

	// Преобразуем тело в джейсон
	marshaledBody, err := p.marshalBody(body)
	if err != nil {
		return nil, err
	}
	pipes := make([]bodyPipe, 0, 2)
	pipes = append(pipes, p.encryptBody)
	if p.client.EnableManualCompression() {
		pipes = append(pipes, p.compressBody)
	}

	// Шифруем тело и Сжимаем тело
	compressedBody, err := p.bodyPipeline(marshaledBody, pipes...)
	if err != nil && !errors.Is(err, ErrorCantCompressBody) && !errors.Is(err, ErrorCantEcryptBody) {
		return nil, err
	}
	if !errors.Is(err, ErrorCantCompressBody) {
		headers = append(headers, Header{
			Name:  "Content-Encoding",
			Value: "gzip",
		})
	}
	if !errors.Is(err, ErrorCantEcryptBody) {
		headers = append(headers, Header{
			Name:  "X-Body-Encrypted",
			Value: "1",
		})
	}

	// Устанавливаем подпись тела
	bodyHash, hashErr := p.hashBody(compressedBody)
	if hashErr != nil {
		logger.Log.Error("Cant hash body", err)
	} else {
		headers = append(headers, Header{
			Name:  "HashSHA256",
			Value: bodyHash,
		})
	}

	// Отправляем запрос
	res, err := p.client.Post(URLUpdates, compressedBody, headers...)
	logger.Log.Info("Finish sending metrics")

	return res, err
}

// marshalBody преобразует тело в строку JSON
func (p *Pool) marshalBody(body []payload.Metrics) ([]byte, error) {
	// Преобразовываем тело в строку джейсон
	return json.Marshal(body)
}

// hashBody создаём подпись запроса
func (p *Pool) hashBody(body []byte) (string, error) {
	if p.HashKey == "" {
		return "", ErrorEmptyHashKey
	}
	harsher := hmac.New(sha256.New, []byte(p.HashKey))
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
		return nil, errors.Join(errors.New("writer is nil"), ErrorCantCompressBody)
	}
	// Устанавливаем новый буфер во врайтер
	writer.Reset(&buf)
	// Кодируем данные
	_, err := writer.Write(body)
	if err != nil {
		return nil, errors.Join(err, ErrorCantCompressBody)
	}
	// Делаем закрытие врайтера, чтобы он прописал все нужные байты в конце. После Reset он откроется снова (Проверял,ставил последовательно закрытие, ресет с новым буфером и снова запись)
	err = writer.Close()
	if err != nil {
		return nil, errors.Join(err, ErrorCantCompressBody)
	}

	return buf.Bytes(), nil
}

// encryptBody шифрует данное тело, используя шифрование RSA с открытым ключом из структуры пула.
// Возвращает зашифрованное тело или ошибку, если шифрование не удалось.
func (p *Pool) encryptBody(body []byte) ([]byte, error) {
	newBody, err := encrypt.Encrypt(body, p.publicKey)
	if err != nil {
		return body, errors.Join(err, ErrorCantEcryptBody)
	}
	return newBody, nil
}

// bodyPipeline обрабатывает данное тело с помощью ряда функций, представленных в вариативном аргументе процесса.
func (p *Pool) bodyPipeline(body []byte, processes ...bodyPipe) ([]byte, error) {
	var err error
	for _, process := range processes {
		var pErr error
		body, pErr = process(body)
		if pErr != nil {
			err = errors.Join(err, pErr)
		}
	}
	return body, err
}
