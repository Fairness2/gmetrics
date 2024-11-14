package encrypt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSplitMessage(t *testing.T) {
	cases := []struct {
		name      string
		body      []byte
		blockSize int
		want      [][]byte
	}{
		{
			name:      "single_block",
			body:      []byte("Hello"),
			blockSize: 5,
			want:      [][]byte{[]byte("Hello")},
		},
		{
			name:      "multiple_blocks",
			body:      []byte("Hello, World!"),
			blockSize: 5,
			want:      [][]byte{[]byte("Hello"), []byte(", Wor"), []byte("ld!")},
		},
		{
			name:      "empty_input",
			body:      []byte(""),
			blockSize: 5,
			want:      [][]byte{},
		},
		{
			name:      "block_size_larger_than_input",
			body:      []byte("Hello"),
			blockSize: 10,
			want:      [][]byte{[]byte("Hello")},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := splitMessage(tt.body, tt.blockSize)
			if len(got) != len(tt.want) {
				t.Errorf("splitMessage() got length = %v, want length = %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if string(got[i]) != string(tt.want[i]) {
					t.Errorf("splitMessage() got block = %v, want block = %v", string(got[i]), string(tt.want[i]))
				}
			}
		})
	}
}

func TestEncrypt(t *testing.T) {
	testKey, err := parsePublicKey([]byte(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAp/UP2HLR5SvocWZMwng64x1mI7L8FqdF4Ae1o08Ygy3I4e8RGV3MLiVBXJqrMIE+ZKafvelxxc3rmUYyJE7xutn7PVLPSQmTah1BNXnOZFiiPXTCLpRUROoObyBW5LTgxUWs/d2AzmHWBPViyBo3ZMUl6/3lhfSFG8tjVIWEp5LCth0A7bpkVWfn7DfaKjOALDAkRwwULWxTiUXVWQcfQn6qPc3Q+ek2PPNuUhQatUQP7OijoexxtRn8+W0dIgwox5zJvwNc0oHapdK+qEkim22YtPEfYNIinYajMHIkx+/8B2RYiDak9ikBsx5UX5UPSBT+kIuNO5KZSciN03Cb7wIDAQAB
-----END PUBLIC KEY-----`))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name      string
		body      []byte
		key       *rsa.PublicKey
		wantError error
	}{
		{
			name:      "successful_encryption",
			body:      []byte("Hello, World!"),
			key:       testKey, // mock key
			wantError: nil,
		},
		{
			name:      "empty_key",
			body:      []byte("Hello, World!"),
			key:       nil,
			wantError: ErrorEmptyKey,
		},
		{
			name:      "long_message",
			body:      []byte("Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!Hello, World!"),
			key:       testKey, // mock key
			wantError: nil,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			message, err := Encrypt(tt.body, tt.key)
			fmt.Printf("%v", message)
			if tt.wantError != nil {
				assert.EqualError(t, err, tt.wantError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// parsePublicKey тестовый ключ публичный
func parsePublicKey(rawKey []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(rawKey)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pub.(*rsa.PublicKey), nil
}

// parsePublicKey тестовый ключ приватный
func parsePrivateKey(rawKey []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(rawKey)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pub.(*rsa.PrivateKey), nil
}

func TestDecrypt(t *testing.T) {
	testKey, err := parsePrivateKey([]byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCn9Q/YctHlK+hxZkzCeDrjHWYjsvwWp0XgB7WjTxiDLcjh7xEZXcwuJUFcmqswgT5kpp+96XHFzeuZRjIkTvG62fs9Us9JCZNqHUE1ec5kWKI9dMIulFRE6g5vIFbktODFRaz93YDOYdYE9WLIGjdkxSXr/eWF9IUby2NUhYSnksK2HQDtumRVZ+fsN9oqM4AsMCRHDBQtbFOJRdVZBx9Cfqo9zdD56TY8825SFBq1RA/s6KOh7HG1Gfz5bR0iDCjHnMm/A1zSgdql0r6oSSKbbZi08R9g0iKdhqMwciTH7/wHZFiINqT2KQGzHlRflQ9IFP6Qi407kplJyI3TcJvvAgMBAAECggEAHACVJDq8eO9xoRpzuMaP1tbLdS89rU81LK1MYM5qoVBMWjLgEHEdfiIS/CwDV6Jssx4+qsyVfeufmJ3l9Ty+O69lHmvEiIJStBHtkctdmEhYwFNLnrV3OUgmoOts4VOw1+MOfQLlm0MfihMZZZBNZP0jne1mS4ehe6lUxb4/CCr/+LTh6anpqUxCEPpvYHrJyENKD50jFZe8DGVV2MNJ3r6b+jzcP/I/gm+4a0n4n4ERGM0KS1g1i2rT+W+fxCIzXbTto9yWatk2w0zi2E9njHBlNEeoRvynSvccIk5xJYIMvSkmRHKZfjIhYprHJueVSNj993xn4yuD1bbv3VBopQKBgQDfvSUMC8iL4eEaxmNTSNFiV64qIRiadhFg1LjnjVjheDoFN31cQhJGJkmX+kZBI2tIkWWQDviy8HDaZin8oeWuNeWyjIQ45UKw9QEA+RBKwttnU/2i36e+vcmDWAhUOIDEfMu+DvnIAEpMSFSnIAUNVO4Mi+KekNhMGCztb4FOlQKBgQDALNsrwLzgQo+vKhsiZ+wJ3f6yAAOyEenmA+h/HQke09noSDCCqb83fCRXNCu1qjhzJ97kwNmZyDmUP3cQnwfnUYAYDq3+Cwiy6mesV5o38eBBgXNlq1oD6qW6Wgi2kRHkccyAPgL7hFoq1tP9yDtRbG1jzZGlhNhKeHmg+GBTcwKBgEu8EtZJBtGS3Efb77M5aucHFwVbvqBKZweH+i8nQXbQ45LwfZbFJrpoK3EuXqmd+6rMzLw+1SB9EzZabsv9YWnfBKmztu4rbK/Jv1U8+a7U1r/bRnfjjTybsaKsIeWgWrYoKC9lkleJAZ1gvobz58HjhdDpaQSTsyPO6yZUIEkhAoGBAJrgS65CQbX2zrebloyu9iKpn4cyzcen+joesjQnYV9P2xEBhN75EJsV2G/TItrgmWftHQx8g6IVJJpeX4WstQDuxO4efoj7uYH/uZfCbg5iR5pjSm4In54CcJfz0YvY9HOIZwh/cYXkj4pw4h5oTa38VViWpqefnXS/DT72jSMTAoGBALbDrCG/Xhm4lSeew9D3AK6i/SCuWXq2W52YAO2DFvuLYc0TeI5zynNmFhIvwpDf4kDERIRvXzQ8pedlWY6MK5uoajwZqmz8Jajezx4O6cKpe+z0b0gq05qe9A0o5YlvZPIZmgt6LjRP9cmGWIIGTFj/fkW207mAqv9GGCp09wzm
-----END RSA PRIVATE KEY-----`))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name         string
		message      []byte
		wantError    bool
		getDecrypter func() *Decrypter
	}{
		{
			name:      "successful_decryption",
			message:   []byte{114, 63, 75, 164, 52, 126, 200, 175, 152, 199, 114, 154, 189, 175, 171, 57, 245, 185, 58, 12, 252, 95, 26, 163, 83, 96, 152, 141, 34, 157, 43, 211, 125, 211, 96, 233, 202, 91, 15, 168, 66, 84, 241, 17, 248, 166, 61, 196, 2, 225, 192, 47, 117, 146, 193, 142, 88, 119, 63, 9, 51, 192, 164, 239, 110, 19, 148, 217, 83, 18, 223, 201, 30, 182, 230, 146, 86, 109, 133, 82, 86, 88, 242, 157, 69, 133, 172, 81, 47, 255, 61, 91, 200, 158, 211, 129, 217, 63, 88, 30, 52, 19, 84, 56, 242, 75, 69, 102, 84, 123, 2, 94, 235, 25, 57, 221, 163, 202, 135, 13, 157, 106, 242, 7, 95, 37, 255, 25, 54, 202, 142, 238, 111, 167, 148, 24, 69, 29, 60, 75, 195, 212, 80, 146, 94, 147, 132, 254, 161, 103, 195, 39, 197, 244, 54, 96, 217, 119, 70, 101, 80, 110, 115, 45, 197, 176, 97, 124, 95, 54, 143, 85, 115, 95, 61, 241, 24, 43, 50, 8, 159, 217, 8, 229, 57, 187, 254, 54, 58, 234, 142, 194, 152, 255, 45, 213, 220, 5, 254, 228, 38, 9, 64, 205, 182, 148, 58, 211, 208, 201, 245, 159, 74, 42, 21, 218, 43, 171, 178, 238, 175, 180, 111, 83, 60, 124, 207, 122, 201, 73, 96, 101, 225, 32, 81, 68, 160, 176, 196, 53, 3, 40, 150, 186, 72, 182, 28, 232, 101, 6, 102, 68, 197, 152, 10, 119},
			wantError: false,
			getDecrypter: func() *Decrypter {
				return &Decrypter{privateKey: testKey}
			},
		},
		{
			name:      "long_message",
			message:   []byte{110, 130, 9, 210, 33, 76, 22, 72, 103, 22, 108, 228, 115, 153, 254, 135, 44, 181, 142, 246, 119, 42, 149, 204, 18, 155, 202, 164, 195, 195, 117, 168, 94, 251, 111, 51, 219, 55, 242, 112, 90, 54, 194, 219, 84, 210, 22, 38, 116, 124, 81, 152, 229, 248, 233, 19, 54, 252, 14, 184, 184, 106, 14, 46, 216, 74, 113, 238, 144, 188, 15, 182, 31, 202, 107, 140, 3, 108, 208, 129, 20, 111, 0, 21, 85, 28, 7, 87, 44, 201, 5, 246, 152, 126, 190, 201, 2, 55, 97, 194, 28, 247, 2, 245, 68, 168, 20, 187, 144, 100, 230, 143, 194, 189, 0, 185, 210, 226, 68, 222, 230, 194, 240, 63, 149, 207, 102, 130, 38, 81, 247, 129, 49, 144, 3, 166, 14, 45, 90, 250, 129, 222, 113, 181, 233, 81, 244, 109, 20, 1, 111, 126, 49, 55, 195, 232, 16, 229, 102, 184, 170, 7, 229, 208, 224, 79, 112, 96, 158, 193, 7, 6, 192, 103, 50, 215, 45, 173, 93, 41, 232, 11, 89, 118, 0, 40, 201, 196, 55, 36, 20, 250, 14, 250, 147, 31, 33, 134, 200, 182, 145, 252, 234, 44, 212, 245, 127, 147, 109, 255, 224, 182, 239, 104, 116, 172, 27, 210, 12, 205, 74, 235, 41, 143, 189, 208, 61, 75, 199, 43, 125, 229, 201, 169, 3, 35, 70, 19, 28, 226, 118, 50, 77, 115, 187, 12, 112, 179, 207, 114, 83, 188, 225, 28, 65, 52, 135, 60, 134, 147, 106, 228, 156, 215, 112, 76, 146, 9, 149, 161, 131, 142, 52, 107, 244, 105, 97, 121, 182, 218, 83, 253, 150, 246, 210, 51, 0, 249, 61, 213, 17, 215, 64, 190, 222, 61, 116, 10, 30, 121, 92, 186, 56, 104, 79, 12, 69, 115, 215, 30, 51, 47, 218, 242, 146, 147, 141, 106, 141, 24, 19, 25, 118, 41, 38, 177, 43, 247, 68, 82, 82, 199, 26, 28, 91, 104, 247, 32, 27, 61, 227, 250, 84, 203, 126, 125, 122, 155, 239, 202, 64, 82, 253, 218, 90, 232, 92, 177, 147, 245, 161, 62, 162, 241, 13, 150, 52, 138, 175, 104, 250, 249, 156, 122, 81, 137, 255, 194, 152, 5, 183, 214, 50, 45, 107, 126, 152, 122, 47, 0, 163, 44, 8, 193, 217, 223, 66, 119, 231, 113, 233, 18, 221, 110, 82, 184, 175, 203, 171, 152, 115, 114, 56, 23, 39, 159, 239, 7, 139, 56, 62, 145, 237, 229, 180, 36, 103, 138, 164, 213, 113, 195, 96, 133, 44, 82, 49, 38, 44, 91, 97, 172, 39, 138, 144, 49, 212, 47, 7, 108, 127, 101, 51, 246, 66, 144, 248, 12, 204, 127, 136, 85, 96, 2, 127, 189, 226, 168, 69, 220, 169, 118, 91, 197, 26, 239, 214, 44, 6, 79, 95, 112, 143, 21, 179, 108, 238, 180, 34, 80, 250, 183, 48, 150, 39, 19, 193, 251, 114, 80, 17, 30, 116, 40, 139, 235, 173, 110, 9, 159, 244, 193, 41, 247, 11, 230, 45, 58, 88, 223, 193, 43, 40, 201, 5, 224, 4, 206, 180, 71, 241, 201, 181, 29, 137, 239, 195, 18, 219, 241, 137, 70, 170, 234, 37, 129, 148, 124, 4, 84, 82, 5, 90, 207, 140, 137, 86, 70, 120, 139, 171, 100, 24, 60, 160, 33, 157, 8, 216, 121, 2, 51, 195, 48, 74, 47, 188, 240, 250, 24, 1, 178, 133, 110, 141, 105, 114, 208, 18, 152, 85, 50, 194, 174, 114, 144, 51, 201, 100, 55, 100, 140, 190, 189, 66, 58, 58, 173, 58, 44, 97, 107, 215, 103, 37, 240, 178, 200, 240, 234, 15, 211, 32, 126, 214, 169, 250, 252, 135, 240, 222, 121, 21, 83, 178, 89, 207, 180, 185, 195, 123, 169, 8, 60, 70, 91, 221, 195, 191, 148, 226, 197, 156, 234, 52, 109, 175, 57, 193, 115, 138, 46, 251, 34, 159, 106, 179, 29, 128, 106, 200, 150, 26, 160, 53, 73, 180, 102, 247, 74, 250, 124, 147, 32, 221, 166, 141, 16, 59, 5, 184, 126, 237, 171, 210, 177, 63, 252, 228, 172, 201, 173, 152, 162, 228, 183, 86, 171, 251, 33, 207, 107, 49, 55, 157, 138, 197, 225, 55, 187, 147, 86, 106, 67, 195, 239, 167, 39, 202, 152, 82, 131, 240, 32, 30, 217, 64, 184, 236, 4, 113, 183, 179, 245, 81, 235, 233, 231, 158, 197, 225, 99, 68, 132, 189, 235, 0, 32, 132, 99, 143, 54, 204, 133, 10, 79, 27, 235},
			wantError: false,
			getDecrypter: func() *Decrypter {
				return &Decrypter{privateKey: testKey}
			},
		},
		{
			name:      "broke_decryption",
			message:   []byte{141, 34, 157, 43, 211, 125, 211, 96, 233, 202, 91, 15, 168, 66, 84, 241, 17, 248, 166, 61, 196, 2, 225, 192, 47, 117, 146, 193, 142, 88, 119, 63, 9, 51, 192, 164, 239, 110, 19, 148, 217, 83, 18, 223, 201, 30, 182, 230, 146, 86, 109, 133, 82, 86, 88, 242, 157, 69, 133, 172, 81, 47, 255, 61, 91, 200, 158, 211, 129, 217, 63, 88, 30, 52, 19, 84, 56, 242, 75, 69, 102, 84, 123, 2, 94, 235, 25, 57, 221, 163, 202, 135, 13, 157, 106, 242, 7, 95, 37, 255, 25, 54, 202, 142, 238, 111, 167, 148, 24, 69, 29, 60, 75, 195, 212, 80, 146, 94, 147, 132, 254, 161, 103, 195, 39, 197, 244, 54, 96, 217, 119, 70, 101, 80, 110, 115, 45, 197, 176, 97, 124, 95, 54, 143, 85, 115, 95, 61, 241, 24, 43, 50, 8, 159, 217, 8, 229, 57, 187, 254, 54, 58, 234, 142, 194, 152, 255, 45, 213, 220, 5, 254, 228, 38, 9, 64, 205, 182, 148, 58, 211, 208, 201, 245, 159, 74, 42, 21, 218, 43, 171, 178, 238, 175, 180, 111, 83, 60, 124, 207, 122, 201, 73, 96, 101, 225, 32, 81, 68, 160, 176, 196, 53, 3, 40, 150, 186, 72, 182, 28, 232, 101, 6, 102, 68, 197, 152, 10, 119},
			wantError: true,
			getDecrypter: func() *Decrypter {
				return &Decrypter{privateKey: testKey}
			},
		},
		{
			name:      "nil_key",
			message:   []byte{114, 63, 75, 164, 52, 126, 200, 175, 152, 199, 114, 154, 189, 175, 171, 57, 245, 185, 58, 12, 252, 95, 26, 163, 83, 96, 152, 141, 34, 157, 43, 211, 125, 211, 96, 233, 202, 91, 15, 168, 66, 84, 241, 17, 248, 166, 61, 196, 2, 225, 192, 47, 117, 146, 193, 142, 88, 119, 63, 9, 51, 192, 164, 239, 110, 19, 148, 217, 83, 18, 223, 201, 30, 182, 230, 146, 86, 109, 133, 82, 86, 88, 242, 157, 69, 133, 172, 81, 47, 255, 61, 91, 200, 158, 211, 129, 217, 63, 88, 30, 52, 19, 84, 56, 242, 75, 69, 102, 84, 123, 2, 94, 235, 25, 57, 221, 163, 202, 135, 13, 157, 106, 242, 7, 95, 37, 255, 25, 54, 202, 142, 238, 111, 167, 148, 24, 69, 29, 60, 75, 195, 212, 80, 146, 94, 147, 132, 254, 161, 103, 195, 39, 197, 244, 54, 96, 217, 119, 70, 101, 80, 110, 115, 45, 197, 176, 97, 124, 95, 54, 143, 85, 115, 95, 61, 241, 24, 43, 50, 8, 159, 217, 8, 229, 57, 187, 254, 54, 58, 234, 142, 194, 152, 255, 45, 213, 220, 5, 254, 228, 38, 9, 64, 205, 182, 148, 58, 211, 208, 201, 245, 159, 74, 42, 21, 218, 43, 171, 178, 238, 175, 180, 111, 83, 60, 124, 207, 122, 201, 73, 96, 101, 225, 32, 81, 68, 160, 176, 196, 53, 3, 40, 150, 186, 72, 182, 28, 232, 101, 6, 102, 68, 197, 152, 10, 119},
			wantError: true,
			getDecrypter: func() *Decrypter {
				return &Decrypter{privateKey: nil}
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.getDecrypter().decrypt(tt.message)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	testKey, err := parsePrivateKey([]byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCn9Q/YctHlK+hxZkzCeDrjHWYjsvwWp0XgB7WjTxiDLcjh7xEZXcwuJUFcmqswgT5kpp+96XHFzeuZRjIkTvG62fs9Us9JCZNqHUE1ec5kWKI9dMIulFRE6g5vIFbktODFRaz93YDOYdYE9WLIGjdkxSXr/eWF9IUby2NUhYSnksK2HQDtumRVZ+fsN9oqM4AsMCRHDBQtbFOJRdVZBx9Cfqo9zdD56TY8825SFBq1RA/s6KOh7HG1Gfz5bR0iDCjHnMm/A1zSgdql0r6oSSKbbZi08R9g0iKdhqMwciTH7/wHZFiINqT2KQGzHlRflQ9IFP6Qi407kplJyI3TcJvvAgMBAAECggEAHACVJDq8eO9xoRpzuMaP1tbLdS89rU81LK1MYM5qoVBMWjLgEHEdfiIS/CwDV6Jssx4+qsyVfeufmJ3l9Ty+O69lHmvEiIJStBHtkctdmEhYwFNLnrV3OUgmoOts4VOw1+MOfQLlm0MfihMZZZBNZP0jne1mS4ehe6lUxb4/CCr/+LTh6anpqUxCEPpvYHrJyENKD50jFZe8DGVV2MNJ3r6b+jzcP/I/gm+4a0n4n4ERGM0KS1g1i2rT+W+fxCIzXbTto9yWatk2w0zi2E9njHBlNEeoRvynSvccIk5xJYIMvSkmRHKZfjIhYprHJueVSNj993xn4yuD1bbv3VBopQKBgQDfvSUMC8iL4eEaxmNTSNFiV64qIRiadhFg1LjnjVjheDoFN31cQhJGJkmX+kZBI2tIkWWQDviy8HDaZin8oeWuNeWyjIQ45UKw9QEA+RBKwttnU/2i36e+vcmDWAhUOIDEfMu+DvnIAEpMSFSnIAUNVO4Mi+KekNhMGCztb4FOlQKBgQDALNsrwLzgQo+vKhsiZ+wJ3f6yAAOyEenmA+h/HQke09noSDCCqb83fCRXNCu1qjhzJ97kwNmZyDmUP3cQnwfnUYAYDq3+Cwiy6mesV5o38eBBgXNlq1oD6qW6Wgi2kRHkccyAPgL7hFoq1tP9yDtRbG1jzZGlhNhKeHmg+GBTcwKBgEu8EtZJBtGS3Efb77M5aucHFwVbvqBKZweH+i8nQXbQ45LwfZbFJrpoK3EuXqmd+6rMzLw+1SB9EzZabsv9YWnfBKmztu4rbK/Jv1U8+a7U1r/bRnfjjTybsaKsIeWgWrYoKC9lkleJAZ1gvobz58HjhdDpaQSTsyPO6yZUIEkhAoGBAJrgS65CQbX2zrebloyu9iKpn4cyzcen+joesjQnYV9P2xEBhN75EJsV2G/TItrgmWftHQx8g6IVJJpeX4WstQDuxO4efoj7uYH/uZfCbg5iR5pjSm4In54CcJfz0YvY9HOIZwh/cYXkj4pw4h5oTa38VViWpqefnXS/DT72jSMTAoGBALbDrCG/Xhm4lSeew9D3AK6i/SCuWXq2W52YAO2DFvuLYc0TeI5zynNmFhIvwpDf4kDERIRvXzQ8pedlWY6MK5uoajwZqmz8Jajezx4O6cKpe+z0b0gq05qe9A0o5YlvZPIZmgt6LjRP9cmGWIIGTFj/fkW207mAqv9GGCp09wzm
-----END RSA PRIVATE KEY-----`))
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		desc          string
		hasHeader     bool
		body          []byte
		expectedError bool
		getDecrypter  func() *Decrypter
	}{
		{
			desc:          "correct_hash",
			hasHeader:     true,
			body:          []byte{114, 63, 75, 164, 52, 126, 200, 175, 152, 199, 114, 154, 189, 175, 171, 57, 245, 185, 58, 12, 252, 95, 26, 163, 83, 96, 152, 141, 34, 157, 43, 211, 125, 211, 96, 233, 202, 91, 15, 168, 66, 84, 241, 17, 248, 166, 61, 196, 2, 225, 192, 47, 117, 146, 193, 142, 88, 119, 63, 9, 51, 192, 164, 239, 110, 19, 148, 217, 83, 18, 223, 201, 30, 182, 230, 146, 86, 109, 133, 82, 86, 88, 242, 157, 69, 133, 172, 81, 47, 255, 61, 91, 200, 158, 211, 129, 217, 63, 88, 30, 52, 19, 84, 56, 242, 75, 69, 102, 84, 123, 2, 94, 235, 25, 57, 221, 163, 202, 135, 13, 157, 106, 242, 7, 95, 37, 255, 25, 54, 202, 142, 238, 111, 167, 148, 24, 69, 29, 60, 75, 195, 212, 80, 146, 94, 147, 132, 254, 161, 103, 195, 39, 197, 244, 54, 96, 217, 119, 70, 101, 80, 110, 115, 45, 197, 176, 97, 124, 95, 54, 143, 85, 115, 95, 61, 241, 24, 43, 50, 8, 159, 217, 8, 229, 57, 187, 254, 54, 58, 234, 142, 194, 152, 255, 45, 213, 220, 5, 254, 228, 38, 9, 64, 205, 182, 148, 58, 211, 208, 201, 245, 159, 74, 42, 21, 218, 43, 171, 178, 238, 175, 180, 111, 83, 60, 124, 207, 122, 201, 73, 96, 101, 225, 32, 81, 68, 160, 176, 196, 53, 3, 40, 150, 186, 72, 182, 28, 232, 101, 6, 102, 68, 197, 152, 10, 119},
			expectedError: false,
			getDecrypter: func() *Decrypter {
				return &Decrypter{privateKey: testKey}
			},
		},
		{
			desc:          "broke_decryption",
			body:          []byte{141, 34, 157, 43, 211, 125, 211, 96, 233, 202, 91, 15, 168, 66, 84, 241, 17, 248, 166, 61, 196, 2, 225, 192, 47, 117, 146, 193, 142, 88, 119, 63, 9, 51, 192, 164, 239, 110, 19, 148, 217, 83, 18, 223, 201, 30, 182, 230, 146, 86, 109, 133, 82, 86, 88, 242, 157, 69, 133, 172, 81, 47, 255, 61, 91, 200, 158, 211, 129, 217, 63, 88, 30, 52, 19, 84, 56, 242, 75, 69, 102, 84, 123, 2, 94, 235, 25, 57, 221, 163, 202, 135, 13, 157, 106, 242, 7, 95, 37, 255, 25, 54, 202, 142, 238, 111, 167, 148, 24, 69, 29, 60, 75, 195, 212, 80, 146, 94, 147, 132, 254, 161, 103, 195, 39, 197, 244, 54, 96, 217, 119, 70, 101, 80, 110, 115, 45, 197, 176, 97, 124, 95, 54, 143, 85, 115, 95, 61, 241, 24, 43, 50, 8, 159, 217, 8, 229, 57, 187, 254, 54, 58, 234, 142, 194, 152, 255, 45, 213, 220, 5, 254, 228, 38, 9, 64, 205, 182, 148, 58, 211, 208, 201, 245, 159, 74, 42, 21, 218, 43, 171, 178, 238, 175, 180, 111, 83, 60, 124, 207, 122, 201, 73, 96, 101, 225, 32, 81, 68, 160, 176, 196, 53, 3, 40, 150, 186, 72, 182, 28, 232, 101, 6, 102, 68, 197, 152, 10, 119},
			hasHeader:     true,
			expectedError: true,
			getDecrypter: func() *Decrypter {
				return &Decrypter{privateKey: testKey}
			},
		},
		{
			desc:          "nil_key",
			body:          []byte{114, 63, 75, 164, 52, 126, 200, 175, 152, 199, 114, 154, 189, 175, 171, 57, 245, 185, 58, 12, 252, 95, 26, 163, 83, 96, 152, 141, 34, 157, 43, 211, 125, 211, 96, 233, 202, 91, 15, 168, 66, 84, 241, 17, 248, 166, 61, 196, 2, 225, 192, 47, 117, 146, 193, 142, 88, 119, 63, 9, 51, 192, 164, 239, 110, 19, 148, 217, 83, 18, 223, 201, 30, 182, 230, 146, 86, 109, 133, 82, 86, 88, 242, 157, 69, 133, 172, 81, 47, 255, 61, 91, 200, 158, 211, 129, 217, 63, 88, 30, 52, 19, 84, 56, 242, 75, 69, 102, 84, 123, 2, 94, 235, 25, 57, 221, 163, 202, 135, 13, 157, 106, 242, 7, 95, 37, 255, 25, 54, 202, 142, 238, 111, 167, 148, 24, 69, 29, 60, 75, 195, 212, 80, 146, 94, 147, 132, 254, 161, 103, 195, 39, 197, 244, 54, 96, 217, 119, 70, 101, 80, 110, 115, 45, 197, 176, 97, 124, 95, 54, 143, 85, 115, 95, 61, 241, 24, 43, 50, 8, 159, 217, 8, 229, 57, 187, 254, 54, 58, 234, 142, 194, 152, 255, 45, 213, 220, 5, 254, 228, 38, 9, 64, 205, 182, 148, 58, 211, 208, 201, 245, 159, 74, 42, 21, 218, 43, 171, 178, 238, 175, 180, 111, 83, 60, 124, 207, 122, 201, 73, 96, 101, 225, 32, 81, 68, 160, 176, 196, 53, 3, 40, 150, 186, 72, 182, 28, 232, 101, 6, 102, 68, 197, 152, 10, 119},
			hasHeader:     true,
			expectedError: false,
			getDecrypter: func() *Decrypter {
				return &Decrypter{privateKey: nil}
			},
		},
		{
			desc:          "long_message",
			body:          []byte{110, 130, 9, 210, 33, 76, 22, 72, 103, 22, 108, 228, 115, 153, 254, 135, 44, 181, 142, 246, 119, 42, 149, 204, 18, 155, 202, 164, 195, 195, 117, 168, 94, 251, 111, 51, 219, 55, 242, 112, 90, 54, 194, 219, 84, 210, 22, 38, 116, 124, 81, 152, 229, 248, 233, 19, 54, 252, 14, 184, 184, 106, 14, 46, 216, 74, 113, 238, 144, 188, 15, 182, 31, 202, 107, 140, 3, 108, 208, 129, 20, 111, 0, 21, 85, 28, 7, 87, 44, 201, 5, 246, 152, 126, 190, 201, 2, 55, 97, 194, 28, 247, 2, 245, 68, 168, 20, 187, 144, 100, 230, 143, 194, 189, 0, 185, 210, 226, 68, 222, 230, 194, 240, 63, 149, 207, 102, 130, 38, 81, 247, 129, 49, 144, 3, 166, 14, 45, 90, 250, 129, 222, 113, 181, 233, 81, 244, 109, 20, 1, 111, 126, 49, 55, 195, 232, 16, 229, 102, 184, 170, 7, 229, 208, 224, 79, 112, 96, 158, 193, 7, 6, 192, 103, 50, 215, 45, 173, 93, 41, 232, 11, 89, 118, 0, 40, 201, 196, 55, 36, 20, 250, 14, 250, 147, 31, 33, 134, 200, 182, 145, 252, 234, 44, 212, 245, 127, 147, 109, 255, 224, 182, 239, 104, 116, 172, 27, 210, 12, 205, 74, 235, 41, 143, 189, 208, 61, 75, 199, 43, 125, 229, 201, 169, 3, 35, 70, 19, 28, 226, 118, 50, 77, 115, 187, 12, 112, 179, 207, 114, 83, 188, 225, 28, 65, 52, 135, 60, 134, 147, 106, 228, 156, 215, 112, 76, 146, 9, 149, 161, 131, 142, 52, 107, 244, 105, 97, 121, 182, 218, 83, 253, 150, 246, 210, 51, 0, 249, 61, 213, 17, 215, 64, 190, 222, 61, 116, 10, 30, 121, 92, 186, 56, 104, 79, 12, 69, 115, 215, 30, 51, 47, 218, 242, 146, 147, 141, 106, 141, 24, 19, 25, 118, 41, 38, 177, 43, 247, 68, 82, 82, 199, 26, 28, 91, 104, 247, 32, 27, 61, 227, 250, 84, 203, 126, 125, 122, 155, 239, 202, 64, 82, 253, 218, 90, 232, 92, 177, 147, 245, 161, 62, 162, 241, 13, 150, 52, 138, 175, 104, 250, 249, 156, 122, 81, 137, 255, 194, 152, 5, 183, 214, 50, 45, 107, 126, 152, 122, 47, 0, 163, 44, 8, 193, 217, 223, 66, 119, 231, 113, 233, 18, 221, 110, 82, 184, 175, 203, 171, 152, 115, 114, 56, 23, 39, 159, 239, 7, 139, 56, 62, 145, 237, 229, 180, 36, 103, 138, 164, 213, 113, 195, 96, 133, 44, 82, 49, 38, 44, 91, 97, 172, 39, 138, 144, 49, 212, 47, 7, 108, 127, 101, 51, 246, 66, 144, 248, 12, 204, 127, 136, 85, 96, 2, 127, 189, 226, 168, 69, 220, 169, 118, 91, 197, 26, 239, 214, 44, 6, 79, 95, 112, 143, 21, 179, 108, 238, 180, 34, 80, 250, 183, 48, 150, 39, 19, 193, 251, 114, 80, 17, 30, 116, 40, 139, 235, 173, 110, 9, 159, 244, 193, 41, 247, 11, 230, 45, 58, 88, 223, 193, 43, 40, 201, 5, 224, 4, 206, 180, 71, 241, 201, 181, 29, 137, 239, 195, 18, 219, 241, 137, 70, 170, 234, 37, 129, 148, 124, 4, 84, 82, 5, 90, 207, 140, 137, 86, 70, 120, 139, 171, 100, 24, 60, 160, 33, 157, 8, 216, 121, 2, 51, 195, 48, 74, 47, 188, 240, 250, 24, 1, 178, 133, 110, 141, 105, 114, 208, 18, 152, 85, 50, 194, 174, 114, 144, 51, 201, 100, 55, 100, 140, 190, 189, 66, 58, 58, 173, 58, 44, 97, 107, 215, 103, 37, 240, 178, 200, 240, 234, 15, 211, 32, 126, 214, 169, 250, 252, 135, 240, 222, 121, 21, 83, 178, 89, 207, 180, 185, 195, 123, 169, 8, 60, 70, 91, 221, 195, 191, 148, 226, 197, 156, 234, 52, 109, 175, 57, 193, 115, 138, 46, 251, 34, 159, 106, 179, 29, 128, 106, 200, 150, 26, 160, 53, 73, 180, 102, 247, 74, 250, 124, 147, 32, 221, 166, 141, 16, 59, 5, 184, 126, 237, 171, 210, 177, 63, 252, 228, 172, 201, 173, 152, 162, 228, 183, 86, 171, 251, 33, 207, 107, 49, 55, 157, 138, 197, 225, 55, 187, 147, 86, 106, 67, 195, 239, 167, 39, 202, 152, 82, 131, 240, 32, 30, 217, 64, 184, 236, 4, 113, 183, 179, 245, 81, 235, 233, 231, 158, 197, 225, 99, 68, 132, 189, 235, 0, 32, 132, 99, 143, 54, 204, 133, 10, 79, 27, 235},
			expectedError: false,
			hasHeader:     true,
			getDecrypter: func() *Decrypter {
				return &Decrypter{privateKey: testKey}
			},
		},
		{
			desc:          "empty_header",
			hasHeader:     false,
			body:          []byte{114, 63, 75, 164, 52, 126, 200, 175, 152, 199, 114, 154, 189, 175, 171, 57, 245, 185, 58, 12, 252, 95, 26, 163, 83, 96, 152, 141, 34, 157, 43, 211, 125, 211, 96, 233, 202, 91, 15, 168, 66, 84, 241, 17, 248, 166, 61, 196, 2, 225, 192, 47, 117, 146, 193, 142, 88, 119, 63, 9, 51, 192, 164, 239, 110, 19, 148, 217, 83, 18, 223, 201, 30, 182, 230, 146, 86, 109, 133, 82, 86, 88, 242, 157, 69, 133, 172, 81, 47, 255, 61, 91, 200, 158, 211, 129, 217, 63, 88, 30, 52, 19, 84, 56, 242, 75, 69, 102, 84, 123, 2, 94, 235, 25, 57, 221, 163, 202, 135, 13, 157, 106, 242, 7, 95, 37, 255, 25, 54, 202, 142, 238, 111, 167, 148, 24, 69, 29, 60, 75, 195, 212, 80, 146, 94, 147, 132, 254, 161, 103, 195, 39, 197, 244, 54, 96, 217, 119, 70, 101, 80, 110, 115, 45, 197, 176, 97, 124, 95, 54, 143, 85, 115, 95, 61, 241, 24, 43, 50, 8, 159, 217, 8, 229, 57, 187, 254, 54, 58, 234, 142, 194, 152, 255, 45, 213, 220, 5, 254, 228, 38, 9, 64, 205, 182, 148, 58, 211, 208, 201, 245, 159, 74, 42, 21, 218, 43, 171, 178, 238, 175, 180, 111, 83, 60, 124, 207, 122, 201, 73, 96, 101, 225, 32, 81, 68, 160, 176, 196, 53, 3, 40, 150, 186, 72, 182, 28, 232, 101, 6, 102, 68, 197, 152, 10, 119},
			expectedError: false,
			getDecrypter: func() *Decrypter {
				return &Decrypter{privateKey: testKey}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			router := chi.NewRouter()
			router.Use(tc.getDecrypter().Middleware)
			router.Post("/", func(writer http.ResponseWriter, request *http.Request) {})
			// запускаем тестовый сервер, будет выбран первый свободный порт
			srv := httptest.NewServer(router)
			// останавливаем сервер после завершения теста
			defer srv.Close()

			request := resty.New().R()
			if tc.hasHeader {
				request.Header.Set("X-Body-Encrypted", "1")
			}

			request.SetBody(tc.body)
			request.Method = http.MethodPost
			request.URL = srv.URL
			res, err := request.Send()
			assert.NoError(t, err, "error making HTTP request")
			if !tc.expectedError {
				assert.Equal(t, http.StatusOK, res.StatusCode())
			} else {
				assert.Equal(t, http.StatusBadRequest, res.StatusCode())
			}
		})
	}
}

func TestNewDecrypter(t *testing.T) {
	testKey, err := parsePrivateKey([]byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCn9Q/YctHlK+hxZkzCeDrjHWYjsvwWp0XgB7WjTxiDLcjh7xEZXcwuJUFcmqswgT5kpp+96XHFzeuZRjIkTvG62fs9Us9JCZNqHUE1ec5kWKI9dMIulFRE6g5vIFbktODFRaz93YDOYdYE9WLIGjdkxSXr/eWF9IUby2NUhYSnksK2HQDtumRVZ+fsN9oqM4AsMCRHDBQtbFOJRdVZBx9Cfqo9zdD56TY8825SFBq1RA/s6KOh7HG1Gfz5bR0iDCjHnMm/A1zSgdql0r6oSSKbbZi08R9g0iKdhqMwciTH7/wHZFiINqT2KQGzHlRflQ9IFP6Qi407kplJyI3TcJvvAgMBAAECggEAHACVJDq8eO9xoRpzuMaP1tbLdS89rU81LK1MYM5qoVBMWjLgEHEdfiIS/CwDV6Jssx4+qsyVfeufmJ3l9Ty+O69lHmvEiIJStBHtkctdmEhYwFNLnrV3OUgmoOts4VOw1+MOfQLlm0MfihMZZZBNZP0jne1mS4ehe6lUxb4/CCr/+LTh6anpqUxCEPpvYHrJyENKD50jFZe8DGVV2MNJ3r6b+jzcP/I/gm+4a0n4n4ERGM0KS1g1i2rT+W+fxCIzXbTto9yWatk2w0zi2E9njHBlNEeoRvynSvccIk5xJYIMvSkmRHKZfjIhYprHJueVSNj993xn4yuD1bbv3VBopQKBgQDfvSUMC8iL4eEaxmNTSNFiV64qIRiadhFg1LjnjVjheDoFN31cQhJGJkmX+kZBI2tIkWWQDviy8HDaZin8oeWuNeWyjIQ45UKw9QEA+RBKwttnU/2i36e+vcmDWAhUOIDEfMu+DvnIAEpMSFSnIAUNVO4Mi+KekNhMGCztb4FOlQKBgQDALNsrwLzgQo+vKhsiZ+wJ3f6yAAOyEenmA+h/HQke09noSDCCqb83fCRXNCu1qjhzJ97kwNmZyDmUP3cQnwfnUYAYDq3+Cwiy6mesV5o38eBBgXNlq1oD6qW6Wgi2kRHkccyAPgL7hFoq1tP9yDtRbG1jzZGlhNhKeHmg+GBTcwKBgEu8EtZJBtGS3Efb77M5aucHFwVbvqBKZweH+i8nQXbQ45LwfZbFJrpoK3EuXqmd+6rMzLw+1SB9EzZabsv9YWnfBKmztu4rbK/Jv1U8+a7U1r/bRnfjjTybsaKsIeWgWrYoKC9lkleJAZ1gvobz58HjhdDpaQSTsyPO6yZUIEkhAoGBAJrgS65CQbX2zrebloyu9iKpn4cyzcen+joesjQnYV9P2xEBhN75EJsV2G/TItrgmWftHQx8g6IVJJpeX4WstQDuxO4efoj7uYH/uZfCbg5iR5pjSm4In54CcJfz0YvY9HOIZwh/cYXkj4pw4h5oTa38VViWpqefnXS/DT72jSMTAoGBALbDrCG/Xhm4lSeew9D3AK6i/SCuWXq2W52YAO2DFvuLYc0TeI5zynNmFhIvwpDf4kDERIRvXzQ8pedlWY6MK5uoajwZqmz8Jajezx4O6cKpe+z0b0gq05qe9A0o5YlvZPIZmgt6LjRP9cmGWIIGTFj/fkW207mAqv9GGCp09wzm
-----END RSA PRIVATE KEY-----`))
	if err != nil {
		t.Fatal(err)
	}
	dec := NewDecrypter(testKey)
	assert.NotNil(t, dec)
}
