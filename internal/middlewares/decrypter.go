package middlewares

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"gmetrics/internal/helpers"
	"io"
	"net/http"
)

type Decrypter struct {
	privateKey *rsa.PrivateKey
}

func NewDecrypter(privateKey *rsa.PrivateKey) Decrypter {
	return Decrypter{privateKey: privateKey}
}

func (d Decrypter) decrypt(message []byte) ([]byte, error) {
	label := []byte("")
	hash := sha256.New()
	return rsa.DecryptOAEP(hash, rand.Reader, d.privateKey, message, label)
}

func (d Decrypter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
