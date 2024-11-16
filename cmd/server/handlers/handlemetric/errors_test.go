package handlemetric

import (
	"errors"
	"net/http"
	"testing"
)

func TestUnwrap(t *testing.T) {
	tests := []struct {
		name string
		err  *UpdateMetricError
		want error
	}{
		{
			name: "BaseCase",
			err:  &UpdateMetricError{error: errors.New("test error"), HTTPStatus: http.StatusOK},
			want: nil,
		},
		{
			name: "NilError",
			err:  &UpdateMetricError{error: nil, HTTPStatus: http.StatusOK},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Unwrap(); got != tt.want {
				t.Errorf("UpdateMetricError.Unwrap() = %v, want %v", got, tt.want)
			}
		})
	}
}
