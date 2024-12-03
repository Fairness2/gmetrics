package main

import (
	"context"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"gmetrics/cmd/server/config"
	"gmetrics/internal/database"
	"gmetrics/internal/metrics"
	"net"
	"net/http"
	"os"
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

func TestStartRPC(t *testing.T) {
	tests := []struct {
		name        string
		listener    net.Listener
		expectedErr bool
	}{
		{
			name: "valid_rpc_initialization",
			listener: func() net.Listener {
				l, _ := net.Listen("tcp", ":0")
				return l
			}(),
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Params = &config.CliConfig{}
			go func() {
				<-time.After(1 * time.Second)
				if tt.listener != nil {
					defer tt.listener.Close()
				}
			}()
			err := startRPC(tt.listener)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.NotContains(t, err.Error(), "use of closed network connection")
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "use of closed network connection")
			}
		})
	}
}

func TestCloseDB(t *testing.T) {
	tests := []struct {
		name        string
		preFunc     func()
		expectedErr bool
	}{
		{
			name: "valid_database_closing",
			preFunc: func() {
				database.DB, _ = sql.Open("sqlite3", ":memory:")
			},
			expectedErr: false,
		},
		{
			name:        "closing_nil_database",
			expectedErr: false,
			preFunc: func() {
				database.DB = nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preFunc != nil {
				tt.preFunc()
			}
			err := closeDB()
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			/*defer closeDB()
			if database.DB != nil {
				err := database.DB.Ping()
				assert.Equal(t, tt.expectedErr, err)
			}*/
		})
	}
}

func TestInitStore(t *testing.T) {
	tests := []struct {
		name      string
		preFunc   func()
		wantStore string
	}{
		/*{
			name: "init_db_store",
			preFunc: func() {
				config.Params = &config.CliConfig{DatabaseDSN: "sqlite3_memory"}
				database.DB, _ = sql.Open("sqlite3", ":memory:")
			},
			wantStore: "db",
		},*/
		{
			name: "init_file_store",
			preFunc: func() {
				config.Params = &config.CliConfig{FileStorage: "test_storage_file.json"}
			},
			wantStore: "file",
		},
		{
			name: "init_memory_store",
			preFunc: func() {
				config.Params = &config.CliConfig{}
			},
			wantStore: "mem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.preFunc()
			if config.Params.FileStorage != "" {
				defer os.Remove(config.Params.FileStorage)
			}
			InitStore(context.TODO())
			switch tt.wantStore {
			case "db":
				_, ok := metrics.MeStore.(*metrics.DBStorage)
				assert.True(t, ok)
			case "file":
				_, ok := metrics.MeStore.(*metrics.DurationFileStorage)
				assert.True(t, ok)
			case "mem":
				_, ok := metrics.MeStore.(*metrics.MemStorage)
				assert.True(t, ok)
			}
		})
	}
}
