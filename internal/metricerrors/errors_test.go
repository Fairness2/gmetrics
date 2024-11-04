package metricerrors

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRetriable(t *testing.T) {

	testCases := []struct {
		name  string
		input error
		want  *Retriable
	}{
		{
			name:  "nil_error",
			input: nil,
			want:  &Retriable{original: nil},
		},
		{
			name:  "non_nil_error",
			input: errors.New("sample error"),
			want:  &Retriable{original: errors.New("sample error")},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewRetriable(tc.input)

			if got == nil {
				assert.Error(t, got, "got nil, want non-nil")
				return
			}
			assert.Equal(t, tc.want, got, "got %v, want %v", got, tc.want)
		})
	}
}

func TestUnwrap(t *testing.T) {

	testCases := []struct {
		name  string
		input error
		want  error
	}{
		{
			name:  "nil_error",
			input: nil,
			want:  nil,
		},
		{
			name:  "non_nil_error",
			input: errors.New("sample error"),
			want:  errors.New("sample error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := &Retriable{original: tc.input}
			got := r.Unwrap()

			assert.Equal(t, tc.want, got, "got %v, want %v", got, tc.want)
		})
	}
}

func TestError(t *testing.T) {

	testCases := []struct {
		name  string
		input error
		want  string
	}{
		{
			name:  "nil_error",
			input: nil,
			want:  "retriable error",
		},
		{
			name:  "non_nil_error",
			input: errors.New("sample error"),
			want:  "sample error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := &Retriable{original: tc.input}
			got := r.Error()

			assert.Equal(t, tc.want, got, "got %v, want %v", got, tc.want)
		})
	}
}
