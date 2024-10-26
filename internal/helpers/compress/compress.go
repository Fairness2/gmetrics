package compress

import (
	"compress/gzip"
	"errors"
	"gmetrics/internal/logger"
	"io"
	"net/http"
	"sync"
)

// Reader представляет io.ReadCloser, который разжимает данные, считанные из базового средства чтения.
type Reader struct {
	originalReader io.ReadCloser
	cReader        io.ReadCloser
}

// Read чтение разжатых данных
func (r *Reader) Read(p []byte) (n int, err error) {
	return r.cReader.Read(p)
}

// Close закрытие оригинального и разжимающего читателя
func (r *Reader) Close() error {
	if err := r.originalReader.Close(); err != nil {
		return err
	}
	return r.cReader.Close()
}

// HTTPWriter представляет собой средство записи HTTP-ответов, которое имплементирует ResponseWriter
// и определяет io.WriteCloser для сжатия данных ответа.
type HTTPWriter struct {
	http.ResponseWriter
	cWriter io.WriteCloser
}

// Write сжимает переданные данные
func (hw *HTTPWriter) Write(p []byte) (int, error) {
	return hw.cWriter.Write(p)
}

// Close закрывает врайтер сжатия данных и отчищает его буфер
func (hw *HTTPWriter) Close() error {
	return hw.cWriter.Close()
}

// NewGZIPReader возвращает новый экземпляр CompressReader, использующий gzip.Reader для разжатия данных.
// Original представляет интерфейс io.ReadCloser, из которого будут считываться данные.
// Возвращает экземпляр CompressReader и ошибку, если нет возможности создать gzip.Reader из originalReader.
func NewGZIPReader(original io.ReadCloser) (*Reader, error) {
	reader, err := gzip.NewReader(original)
	if err != nil {
		return nil, err
	}
	return &Reader{
		originalReader: original,
		cReader:        reader,
	}, nil
}

// writerPool представляет собой синглтон экземпляр WriterPool, управляющий многократно используемыми экземплярами gzip.Writer для оптимизации ресурсов.
var writerPool *WriterPool

// writerOnce гарантирует, что writerPool инициализируется только один раз с помощью sync.Once, обеспечивая потокобезопасную отложенную инициализацию.
var writerOnce sync.Once

// GetGZIPHTTPWriter возвращает новый экземпляр HttpWriter, который использует gzip.Writer для сжатия данных из оригинального http.ResponseWriter
func GetGZIPHTTPWriter(original http.ResponseWriter) (*HTTPWriter, error) {
	writerOnce.Do(func() {
		writerPool = NewWriterPool()
	})
	writer := writerPool.Get()
	writer.Reset(original)
	if writer == nil {
		return nil, errors.New("can't get gzip writer")
	}
	return &HTTPWriter{
		ResponseWriter: original,
		cWriter:        writer,
	}, nil
}

// WriterPool управляет пулом повторно используемых экземпляров gzip.Writer для оптимизации использования ресурсов и повышения производительности.
type WriterPool struct {
	pool sync.Pool
}

// NewWriterPool Создание нового пула врайтеров со сжатием
func NewWriterPool() *WriterPool {
	return &WriterPool{
		pool: sync.Pool{
			New: newEncoder,
		},
	}
}

// Get Получение врайтера из пула готовых
func (wp *WriterPool) Get() *gzip.Writer {
	return wp.pool.Get().(*gzip.Writer)
}

// newEncoder инициализирует и возвращает новый gzip.Writer. Возвращает ноль в случае возникновения ошибки.
func newEncoder() interface{} {
	writer, err := gzip.NewWriterLevel(nil, gzip.BestSpeed)
	if err != nil {
		logger.Log.Error(err)
		return nil
	}
	return writer
}
