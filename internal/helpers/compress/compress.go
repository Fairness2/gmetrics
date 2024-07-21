package compress

import (
	"compress/gzip"
	"io"
	"net/http"
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

// HttpWriter представляет собой средство записи HTTP-ответов, которое имплементирует ResponseWriter
// и определяет io.WriteCloser для сжатия данных ответа.
type HttpWriter struct {
	http.ResponseWriter
	cWriter io.WriteCloser
}

// Write сжимает переданные данные
func (hw *HttpWriter) Write(p []byte) (int, error) {
	return hw.cWriter.Write(p)
}

// Close закрывает врайтер сжатия данных и отчищает его буфер
func (hw *HttpWriter) Close() error {
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

// NewGZIPHTTPWriter возвращает новый экземпляр HttpWriter, который использует gzip.Writer для сжатия данных из оригинального http.ResponseWriter
func NewGZIPHTTPWriter(original http.ResponseWriter) (*HttpWriter, error) {
	writer, err := gzip.NewWriterLevel(original, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	return &HttpWriter{
		ResponseWriter: original,
		cWriter:        writer,
	}, nil
}
