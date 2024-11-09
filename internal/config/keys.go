package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

var ErrorEmptyPublicKeyPath = errors.New("no public key path specified")

// ParseKeyFromFile получаем ключ из указанного файла
func ParseKeyFromFile[T rsa.PublicKey | rsa.PrivateKey](publicKeyPath string, blockType string) (*T, error) {
	if publicKeyPath == "" {
		return nil, ErrorEmptyPublicKeyPath
	}

	rawKey, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(rawKey)
	if block == nil || block.Type != blockType {
		return nil, errors.New("failed to decode PEM block containing public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pub.(*T), nil
}
