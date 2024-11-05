package sendpool

import (
	"context"
	"github.com/golang/mock/gomock"
	"gmetrics/internal/payload"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWithClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := []struct {
		name    string
		size    int
		hashKey string
		client  IClient
		wantErr error
	}{
		{
			name:    "valid_input",
			size:    10,
			hashKey: "test",
			client:  NewMockIClient(ctrl),
		},
		{
			name:    "negative_size",
			size:    -5,
			hashKey: "test",
			client:  NewMockIClient(ctrl),
			wantErr: ErrorWrongWorkerSize,
		},
		{
			name:    "empty_hash_key",
			size:    10,
			hashKey: "",
			client:  NewMockIClient(ctrl),
			wantErr: ErrorEmptyHashKey,
		},
		{
			name:    "nil_client",
			size:    10,
			hashKey: "test",
			client:  nil,
			wantErr: ErrorEmptyClient,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()

			got, err := NewWithClient(ctx, tt.size, tt.hashKey, tt.client)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			cancel()
			assert.NotNil(t, got)
			assert.Equal(t, tt.hashKey, got.HashKey)
			assert.Equal(t, tt.client, got.client)
			//assert.True(t, got.isClosed)
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		size      int
		hashKey   string
		wantErr   error
		serverURL string
	}{
		{
			name:      "valid_input",
			size:      10,
			hashKey:   "test",
			serverURL: "http://localhost:8080",
		},
		{
			name:      "negative_size",
			size:      -5,
			hashKey:   "test",
			wantErr:   ErrorWrongWorkerSize,
			serverURL: "http://localhost:8080",
		},
		{
			name:      "empty_hash_key",
			size:      10,
			hashKey:   "",
			wantErr:   ErrorEmptyHashKey,
			serverURL: "http://localhost:8080",
		},
		{
			name:      "empty_server_url",
			size:      10,
			hashKey:   "",
			wantErr:   ErrorServerURLIsEmpty,
			serverURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()

			got, err := New(ctx, tt.size, tt.hashKey, tt.serverURL)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			cancel()
			assert.NotNil(t, got)
			assert.Equal(t, tt.hashKey, got.HashKey)
			//assert.True(t, got.isClosed)
		})
	}
}
func TestMarshalBody(t *testing.T) {
	tests := []struct {
		name    string
		body    []payload.Metrics
		wantErr error
	}{
		{
			name: "valid_input",
			body: []payload.Metrics{
				{
					Value: nil,
					Delta: nil,
					ID:    "sadasd",
					MType: "asdasd",
				},
			},
		},
		{
			name:    "empty_input",
			body:    []payload.Metrics{},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pool{}
			_, err := p.marshalBody(tt.body)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
