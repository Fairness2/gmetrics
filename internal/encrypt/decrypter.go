package encrypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"gmetrics/internal/helpers"
	"io"
	"math"
	"net/http"
)

// ErrorEmptyKey Ошибка, что был передан пустой ключ для шифрования или дешифрования
var ErrorEmptyKey = errors.New("empty key")

// Decrypter Класс для дешифрования данных
type Decrypter struct {
	privateKey *rsa.PrivateKey
}

// NewDecrypter Создаёт новый Decrypter с переданным ключом
func NewDecrypter(privateKey *rsa.PrivateKey) Decrypter {
	return Decrypter{privateKey: privateKey}
}

// decrypt Дешифрование переданого тела
func (d Decrypter) decrypt(message []byte) ([]byte, error) {
	label := []byte("")
	hash := sha256.New()
	encryptedBlocks, err := splitMessage(message, d.privateKey.Size())
	if err != nil {
		return nil, err
	}
	blocks := make([][]byte, len(encryptedBlocks))
	for i, block := range encryptedBlocks {
		newBlock, err := rsa.DecryptOAEP(hash, rand.Reader, d.privateKey, block, label)
		if err != nil {
			return nil, err
		}
		blocks[i] = newBlock
	}
	return bytes.Join(blocks, nil), nil
}

// Middleware Функция для создания мидлваре для дешифровки сообщения
func (d Decrypter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h := r.Header.Get("X-Body-Encrypted"); h == "" || d.privateKey == nil {
			next.ServeHTTP(w, r)
		}
		newWriter := w
		// Читаем тело запроса
		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			helpers.SetHTTPResponse(w, http.StatusBadRequest, []byte(err.Error()))
			return
		}
		decryptBody, err := d.decrypt(rawBody)
		if err != nil {
			helpers.SetHTTPResponse(w, http.StatusBadRequest, []byte(err.Error()))
		}
		// Ставим тело снова, чтобы его можно было прочитать снова.
		r.Body = io.NopCloser(bytes.NewBuffer(decryptBody))

		next.ServeHTTP(newWriter, r)
	})
}

// Разделение текста на блоки нужного размера
func splitMessage(body []byte, blockSize int) ([][]byte, error) {
	var ln = math.Ceil(float64(len(body)) / float64(blockSize))
	blocks := make([][]byte, 0, int(ln))
	for i := 0; i < len(body); i += blockSize {
		end := i + blockSize
		if end > len(body) {
			end = len(body)
		}
		blocks = append(blocks, body[i:end])
	}
	return blocks, nil
}

// Encrypt шифрует данное тело, используя шифрование RSA с открытым ключом из структуры пула.
// Возвращает зашифрованное тело или ошибку, если шифрование не удалось.
func Encrypt(body []byte, key *rsa.PublicKey) ([]byte, error) {
	if key == nil {
		return body, ErrorEmptyKey
	}
	label := []byte("")
	hash := sha256.New()
	blockSize := key.Size() - 2*hash.Size() - 2
	blocks, err := splitMessage(body, blockSize)
	if err != nil {
		return body, err
	}
	encryptedBlocks := make([][]byte, len(blocks))
	for i, block := range blocks {
		newBlock, err := rsa.EncryptOAEP(hash, rand.Reader, key, block, label)
		if err != nil {
			return body, err
		}
		encryptedBlocks[i] = newBlock
	}
	newBody := bytes.Join(encryptedBlocks, nil)
	return newBody, nil
}
