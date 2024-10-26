package fileworker

import (
	"encoding/json"
	"io"
	"os"
	"sync"
)

// JSONWriter управляет записью данных JSON в файл с потокобезопасностью.
type JSONWriter struct {
	file  *os.File
	mutex sync.Mutex
}

// NewWriter создает новый JSONWriter для указанного имени файла. Возвращает ошибку в случае неудачи.
func NewWriter(filename string) (*JSONWriter, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &JSONWriter{
		file:  file,
		mutex: sync.Mutex{},
	}, nil
}

// Write сериализует заданное значение в JSON и записывает его в базовый файл потокобезопасным способом.
func (w *JSONWriter) Write(v any) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	// Отчищаем содержимое файла
	err = w.file.Truncate(0) // empty the content
	if err != nil {
		return err
	}
	// Ставим указатель записи на начало файла
	_, err = w.file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	// Записываем полученный джейсон в файл
	_, err = w.file.Write(body)
	if err != nil {
		return err
	}
	//w.file.Sync() // Заставляем записать всё на диск из буферов

	return nil
}

// Close завершает работу с файлом, закрывая его. Возвращает ошибку, если операция не удалась.
func (w *JSONWriter) Close() error {
	return w.file.Close()
}
