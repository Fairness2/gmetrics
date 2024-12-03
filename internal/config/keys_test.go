package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// Mock files and block type for testing
var (
	testPublicKeyPath  = "test_public.key"
	testBlockType      = "PUBLIC KEY"
	testInvalidKeyPath = "invalid.key"
)

// Create a sample RSA public key for testing
func createTestPublicKey(content []byte) {
	os.WriteFile(testPublicKeyPath, content, os.ModePerm)
}

func TestGetKeyFromFile(t *testing.T) {
	tests := []struct {
		name          string
		publicKeyPath string
		blockType     string
		wantErr       bool
		key           string
	}{
		{
			name:          "valid_key",
			publicKeyPath: testPublicKeyPath,
			blockType:     testBlockType,
			wantErr:       false,
			key: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAp/UP2HLR5SvocWZMwng64x1mI7L8FqdF4Ae1o08Ygy3I4e8RGV3MLiVBXJqrMIE+ZKafvelxxc3rmUYyJE7xutn7PVLPSQmTah1BNXnOZFiiPXTCLpRUROoObyBW5LTgxUWs/d2AzmHWBPViyBo3ZMUl6/3lhfSFG8tjVIWEp5LCth0A7bpkVWfn7DfaKjOALDAkRwwULWxTiUXVWQcfQn6qPc3Q+ek2PPNuUhQatUQP7OijoexxtRn8+W0dIgwox5zJvwNc0oHapdK+qEkim22YtPEfYNIinYajMHIkx+/8B2RYiDak9ikBsx5UX5UPSBT+kIuNO5KZSciN03Cb7wIDAQAB
-----END PUBLIC KEY-----`,
		},
		{
			name:          "invalid_key_path",
			publicKeyPath: testInvalidKeyPath,
			blockType:     testBlockType,
			wantErr:       true,
		},
		{
			name:          "empty_key_path",
			publicKeyPath: "",
			blockType:     testBlockType,
			wantErr:       true,
		},
		{
			name:          "invalid_key",
			publicKeyPath: testPublicKeyPath,
			blockType:     testBlockType,
			wantErr:       true,
			key:           `0BAQEFAAOCAQ8AMIIBCgKCAQEAp/UP2HLR5`,
		},
		{
			name:          "invalid_block_type",
			publicKeyPath: testPublicKeyPath,
			blockType:     testBlockType,
			wantErr:       true,
			key: `-----BEGIN RSA PRIVATE KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAp/UP2HLR5SvocWZMwng64x1mI7L8FqdF4Ae1o08Ygy3I4e8RGV3MLiVBXJqrMIE+ZKafvelxxc3rmUYyJE7xutn7PVLPSQmTah1BNXnOZFiiPXTCLpRUROoObyBW5LTgxUWs/d2AzmHWBPViyBo3ZMUl6/3lhfSFG8tjVIWEp5LCth0A7bpkVWfn7DfaKjOALDAkRwwULWxTiUXVWQcfQn6qPc3Q+ek2PPNuUhQatUQP7OijoexxtRn8+W0dIgwox5zJvwNc0oHapdK+qEkim22YtPEfYNIinYajMHIkx+/8B2RYiDak9ikBsx5UX5UPSBT+kIuNO5KZSciN03Cb7wIDAQAB
-----END RSA PRIVATE KEY-----`,
		},
	}
	// Iterate the test tables
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.publicKeyPath == testPublicKeyPath {
				defer os.Remove(testPublicKeyPath)
				createTestPublicKey([]byte(tt.key))
			}
			got, err := getKeyFromFile(tt.publicKeyPath, tt.blockType)

			if tt.wantErr {
				assert.Error(t, err, "getKeyFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NoError(t, err, "getKeyFromFile() error = %v, wantErr %v", err, tt.wantErr)
			assert.NotNil(t, got)
		})
	}
}

func TestParsePrivateKeyFromFile(t *testing.T) {
	tests := []struct {
		name          string
		publicKeyPath string
		wantErr       bool
		key           string
	}{
		{
			name:          "valid_key",
			publicKeyPath: testPublicKeyPath,
			wantErr:       false,
			key: `-----BEGIN RSA PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCn9Q/YctHlK+hxZkzCeDrjHWYjsvwWp0XgB7WjTxiDLcjh7xEZXcwuJUFcmqswgT5kpp+96XHFzeuZRjIkTvG62fs9Us9JCZNqHUE1ec5kWKI9dMIulFRE6g5vIFbktODFRaz93YDOYdYE9WLIGjdkxSXr/eWF9IUby2NUhYSnksK2HQDtumRVZ+fsN9oqM4AsMCRHDBQtbFOJRdVZBx9Cfqo9zdD56TY8825SFBq1RA/s6KOh7HG1Gfz5bR0iDCjHnMm/A1zSgdql0r6oSSKbbZi08R9g0iKdhqMwciTH7/wHZFiINqT2KQGzHlRflQ9IFP6Qi407kplJyI3TcJvvAgMBAAECggEAHACVJDq8eO9xoRpzuMaP1tbLdS89rU81LK1MYM5qoVBMWjLgEHEdfiIS/CwDV6Jssx4+qsyVfeufmJ3l9Ty+O69lHmvEiIJStBHtkctdmEhYwFNLnrV3OUgmoOts4VOw1+MOfQLlm0MfihMZZZBNZP0jne1mS4ehe6lUxb4/CCr/+LTh6anpqUxCEPpvYHrJyENKD50jFZe8DGVV2MNJ3r6b+jzcP/I/gm+4a0n4n4ERGM0KS1g1i2rT+W+fxCIzXbTto9yWatk2w0zi2E9njHBlNEeoRvynSvccIk5xJYIMvSkmRHKZfjIhYprHJueVSNj993xn4yuD1bbv3VBopQKBgQDfvSUMC8iL4eEaxmNTSNFiV64qIRiadhFg1LjnjVjheDoFN31cQhJGJkmX+kZBI2tIkWWQDviy8HDaZin8oeWuNeWyjIQ45UKw9QEA+RBKwttnU/2i36e+vcmDWAhUOIDEfMu+DvnIAEpMSFSnIAUNVO4Mi+KekNhMGCztb4FOlQKBgQDALNsrwLzgQo+vKhsiZ+wJ3f6yAAOyEenmA+h/HQke09noSDCCqb83fCRXNCu1qjhzJ97kwNmZyDmUP3cQnwfnUYAYDq3+Cwiy6mesV5o38eBBgXNlq1oD6qW6Wgi2kRHkccyAPgL7hFoq1tP9yDtRbG1jzZGlhNhKeHmg+GBTcwKBgEu8EtZJBtGS3Efb77M5aucHFwVbvqBKZweH+i8nQXbQ45LwfZbFJrpoK3EuXqmd+6rMzLw+1SB9EzZabsv9YWnfBKmztu4rbK/Jv1U8+a7U1r/bRnfjjTybsaKsIeWgWrYoKC9lkleJAZ1gvobz58HjhdDpaQSTsyPO6yZUIEkhAoGBAJrgS65CQbX2zrebloyu9iKpn4cyzcen+joesjQnYV9P2xEBhN75EJsV2G/TItrgmWftHQx8g6IVJJpeX4WstQDuxO4efoj7uYH/uZfCbg5iR5pjSm4In54CcJfz0YvY9HOIZwh/cYXkj4pw4h5oTa38VViWpqefnXS/DT72jSMTAoGBALbDrCG/Xhm4lSeew9D3AK6i/SCuWXq2W52YAO2DFvuLYc0TeI5zynNmFhIvwpDf4kDERIRvXzQ8pedlWY6MK5uoajwZqmz8Jajezx4O6cKpe+z0b0gq05qe9A0o5YlvZPIZmgt6LjRP9cmGWIIGTFj/fkW207mAqv9GGCp09wzm
-----END RSA PRIVATE KEY-----`,
		},
		{
			name:          "invalid_key_path",
			publicKeyPath: testInvalidKeyPath,
			wantErr:       true,
		},
		{
			name:          "invalid_key",
			publicKeyPath: testPublicKeyPath,
			wantErr:       true,
			key: `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCjcGqTkOq0CR3rTx0ZSQSIdTrDrFAYl29611xN8aVgMQIWtDB/
lD0W5TpKPuU9iaiG/sSn/VYt6EzN7Sr332jj7cyl2WrrHI6ujRswNy4HojMuqtfa
b5FFDpRmCuvl35fge18OvoQTJELhhJ1EvJ5KUeZiuJ3u3YyMnxxXzLuKbQIDAQAB
AoGAPrNDz7TKtaLBvaIuMaMXgBopHyQd3jFKbT/tg2Fu5kYm3PrnmCoQfZYXFKCo
ZUFIS/G1FBVWWGpD/MQ9tbYZkKpwuH+t2rGndMnLXiTC296/s9uix7gsjnT4Naci
5N6EN9pVUBwQmGrYUTHFc58ThtelSiPARX7LSU2ibtJSv8ECQQDWBRrrAYmbCUN7
ra0DFT6SppaDtvvuKtb+mUeKbg0B8U4y4wCIK5GH8EyQSwUWcXnNBO05rlUPbifs
DLv/u82lAkEAw39sTJ0KmJJyaChqvqAJ8guulKlgucQJ0Et9ppZyet9iVwNKX/aW
9UlwGBMQdafQ36nd1QMEA8AbAw4D+hw/KQJBANJbHDUGQtk2hrSmZNoV5HXB9Uiq
7v4N71k5ER8XwgM5yVGs2tX8dMM3RhnBEtQXXs9LW1uJZSOQcv7JGXNnhN0CQBZe
nzrJAWxh3XtznHtBfsHWelyCYRIAj4rpCHCmaGUM6IjCVKFUawOYKp5mmAyObkUZ
f8ue87emJLEdynC1CLkCQHduNjP1hemAGWrd6v8BHhE3kKtcK6KHsPvJR5dOfzbd
HAqVePERhISfN6cwZt5p8B3/JUwSR8el66DF7Jm57BM=
-----END RSA PRIVATE KEY-----`,
		},
	}
	// Iterate the test tables
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.publicKeyPath == testPublicKeyPath {
				defer os.Remove(testPublicKeyPath)
				createTestPublicKey([]byte(tt.key))
			}
			got, err := ParsePrivateKeyFromFile(tt.publicKeyPath)

			if tt.wantErr {
				assert.Error(t, err, "getKeyFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NoError(t, err, "getKeyFromFile() error = %v, wantErr %v", err, tt.wantErr)
			assert.NotNil(t, got)
		})
	}
}

func TestParsePublicKeyFromFile(t *testing.T) {
	tests := []struct {
		name          string
		publicKeyPath string
		wantErr       bool
		key           string
	}{
		{
			name:          "valid_key",
			publicKeyPath: testPublicKeyPath,
			wantErr:       false,
			key: `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAp/UP2HLR5SvocWZMwng64x1mI7L8FqdF4Ae1o08Ygy3I4e8RGV3MLiVBXJqrMIE+ZKafvelxxc3rmUYyJE7xutn7PVLPSQmTah1BNXnOZFiiPXTCLpRUROoObyBW5LTgxUWs/d2AzmHWBPViyBo3ZMUl6/3lhfSFG8tjVIWEp5LCth0A7bpkVWfn7DfaKjOALDAkRwwULWxTiUXVWQcfQn6qPc3Q+ek2PPNuUhQatUQP7OijoexxtRn8+W0dIgwox5zJvwNc0oHapdK+qEkim22YtPEfYNIinYajMHIkx+/8B2RYiDak9ikBsx5UX5UPSBT+kIuNO5KZSciN03Cb7wIDAQAB
-----END PUBLIC KEY-----`,
		},
		{
			name:          "invalid_key_path",
			publicKeyPath: testInvalidKeyPath,
			wantErr:       true,
		},
		{
			name:          "invalid_key",
			publicKeyPath: testPublicKeyPath,
			wantErr:       true,
			key: `-----BEGIN PUBLIC KEY-----
MIICXAIBAAKBgQCjcGqTkOq0CR3rTx0ZSQSIdTrDrFAYl29611xN8aVgMQIWtDB/
lD0W5TpKPuU9iaiG/sSn/VYt6EzN7Sr332jj7cyl2WrrHI6ujRswNy4HojMuqtfa
b5FFDpRmCuvl35fge18OvoQTJELhhJ1EvJ5KUeZiuJ3u3YyMnxxXzLuKbQIDAQAB
AoGAPrNDz7TKtaLBvaIuMaMXgBopHyQd3jFKbT/tg2Fu5kYm3PrnmCoQfZYXFKCo
ZUFIS/G1FBVWWGpD/MQ9tbYZkKpwuH+t2rGndMnLXiTC296/s9uix7gsjnT4Naci
5N6EN9pVUBwQmGrYUTHFc58ThtelSiPARX7LSU2ibtJSv8ECQQDWBRrrAYmbCUN7
ra0DFT6SppaDtvvuKtb+mUeKbg0B8U4y4wCIK5GH8EyQSwUWcXnNBO05rlUPbifs
DLv/u82lAkEAw39sTJ0KmJJyaChqvqAJ8guulKlgucQJ0Et9ppZyet9iVwNKX/aW
9UlwGBMQdafQ36nd1QMEA8AbAw4D+hw/KQJBANJbHDUGQtk2hrSmZNoV5HXB9Uiq
7v4N71k5ER8XwgM5yVGs2tX8dMM3RhnBEtQXXs9LW1uJZSOQcv7JGXNnhN0CQBZe
nzrJAWxh3XtznHtBfsHWelyCYRIAj4rpCHCmaGUM6IjCVKFUawOYKp5mmAyObkUZ
f8ue87emJLEdynC1CLkCQHduNjP1hemAGWrd6v8BHhE3kKtcK6KHsPvJR5dOfzbd
HAqVePERhISfN6cwZt5p8B3/JUwSR8el66DF7Jm57BM=
-----END PUBLIC KEY-----`,
		},
	}
	// Iterate the test tables
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.publicKeyPath == testPublicKeyPath {
				defer os.Remove(testPublicKeyPath)
				createTestPublicKey([]byte(tt.key))
			}
			got, err := ParsePublicKeyFromFile(tt.publicKeyPath)

			if tt.wantErr {
				assert.Error(t, err, "getKeyFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NoError(t, err, "getKeyFromFile() error = %v, wantErr %v", err, tt.wantErr)
			assert.NotNil(t, got)
		})
	}
}
