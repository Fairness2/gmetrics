package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

var ErrorEmptyKeyPath = errors.New("no key path specified")

// ParsePublicKeyFromFile получаем ключ из указанного файла
func ParsePublicKeyFromFile(publicKeyPath string) (*rsa.PublicKey, error) {
	block, err := getKeyFromFile(publicKeyPath, "PUBLIC KEY")
	if err != nil {
		return nil, err
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pub.(*rsa.PublicKey), nil
}

// ParsePrivateKeyFromFile получаем ключ из указанного файла
func ParsePrivateKeyFromFile(publicKeyPath string) (*rsa.PrivateKey, error) {
	block, err := getKeyFromFile(publicKeyPath, "RSA PRIVATE KEY")
	if err != nil {
		return nil, err
	}
	pub, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pub.(*rsa.PrivateKey), nil
}

// getKeyFromFile читает и декодирует ключ, закодированный в формате PEM, по заданному пути к файлу.
// Функция ожидает в качестве аргументов указанный путь к файлу, содержащему блок PEM, и желаемый тип блока.
// Он возвращает декодированный *pem.Block и ошибку, если она возникает во время чтения или декодирования содержимого файла.
func getKeyFromFile(publicKeyPath string, blockType string) (*pem.Block, error) {
	if publicKeyPath == "" {
		return nil, ErrorEmptyKeyPath
	}

	rawKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(rawKey)
	if block == nil || block.Type != blockType {
		return nil, errors.New("failed to decode PEM block containing public key")
	}
	return block, nil
}
