package fileworker

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWriter(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		shouldErr bool
	}{
		{name: "valid_file_name", filename: "testfile.json", shouldErr: false},
		{name: "empty_file_name", filename: "", shouldErr: true},
		{name: "invalid_file_name", filename: ".", shouldErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewWriter(tc.filename)

			if tc.shouldErr {
				assert.Error(t, err, "NewWriter() should have returned an error")
				return
			} else {
				assert.NoError(t, err, "NewWriter() should not have returned an error")
			}

			if err == nil {
				if _, err = os.Stat(tc.filename); os.IsNotExist(err) {
					assert.Fail(t, "NewWriter() should have created file but didn't")
				}

				err = os.Remove(tc.filename)
				assert.NoError(t, err, "Failed to remove test file")
			}
		})
	}
}

func TestJSONWriter_Write(t *testing.T) {
	tests := []struct {
		name      string
		content   any
		shouldErr bool
	}{
		{name: "valid_content", content: struct{ Name string }{Name: "test"}, shouldErr: false},
		{name: "nil_content", content: nil, shouldErr: false},
		{name: "invalid_content", content: make(chan int), shouldErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.CreateTemp("", "test")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if rErr := os.Remove(f.Name()); rErr != nil {
					t.Fatal(rErr)
				}
			}()

			writer, err := NewWriter(f.Name())
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if cErr := writer.Close(); err != nil {
					log.Print(cErr)
				}
			}()

			err = writer.Write(tc.content)

			if tc.shouldErr {
				assert.Error(t, err, "Write() should have returned an error")
			} else {
				assert.NoError(t, err, "Write() should not have returned an error")
			}
		})
	}
}
