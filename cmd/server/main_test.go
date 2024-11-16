package main

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gmetrics/cmd/server/config"
	"net/http"
	"testing"
	"time"
)

func TestGetRouter(t *testing.T) {
	config.Params = &config.CliConfig{}
	router := getRouter()
	assert.NotNil(t, router, "Router should not be nil")
}

func TestInitServer(t *testing.T) {
	tests := []struct {
		name    string
		preFunc func()
	}{
		{
			name: "valid_server_initialization",
			preFunc: func() {
				config.Params = &config.CliConfig{Address: "localhost:8080"}
			},
		},
		{
			name: "invalid_server_initialization",
			preFunc: func() {
				config.Params = &config.CliConfig{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.preFunc()
			s := initServer()
			assert.NotNil(t, s, "Server should not be nil")
		})
	}
}

func TestStopServer(t *testing.T) {
	config.Params = &config.CliConfig{Address: "localhost:8080"}
	tests := []struct {
		name        string
		preFunc     func(server *http.Server, ctx context.Context)
		expectedErr bool
		closed      bool
	}{
		{
			name: "valid_server_shutdown",
			preFunc: func(server *http.Server, ctx context.Context) {
				_ = server.ListenAndServe()
			},
			expectedErr: false,
		},
		{
			name: "shutdown_not_running_server",
			preFunc: func(server *http.Server, ctx context.Context) {

			},
			expectedErr: false,
		},
		{
			name: "shutdown_with_closed_context",
			preFunc: func(server *http.Server, ctx context.Context) {
				_ = server.ListenAndServe()
			},
			closed:      true,
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &http.Server{Addr: "localhost:8080"}
			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()
			go tt.preFunc(server, ctx)
			time.Sleep(1 * time.Second)
			if tt.closed {
				_ = server.Shutdown(ctx)
			}
			err := stopServer(server, ctx)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
