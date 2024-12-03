package sendpool

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "gmetrics/internal/payload/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net/http"
	"testing"
)

func TestNewRPCResponse(t *testing.T) {
	tests := []struct {
		name string
		code codes.Code
		want int
	}{
		{
			name: "test_ok_status_code",
			code: codes.OK,
			want: http.StatusOK,
		},
		{
			name: "test_invalid_argument_status_code",
			code: codes.InvalidArgument,
			want: http.StatusBadRequest,
		},
		{
			name: "test_internal_and_unavailable_status_code",
			code: codes.Internal,
			want: http.StatusInternalServerError,
		},
		{
			name: "test_unknown_status_code",
			code: codes.Code(1234),
			want: http.StatusTeapot,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpcResponse := NewRPCResponse(tt.code)
			assert.Equal(t, tt.want, rpcResponse.StatusCode(), "StatusCode() should return %d", tt.want)
		})
	}
}

func TestEnableManualCompression(t *testing.T) {
	tests := []struct {
		name      string
		client    RPCClient
		expectVal bool
	}{
		{
			name:      "test_enable_manual_compression_false",
			client:    RPCClient{},
			expectVal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.client.EnableManualCompression()
			assert.Equal(t, tt.expectVal, result, "EnableManualCompression() should return %v", tt.expectVal)
		})
	}
}

// TestClearURL - test cases for clearURL function
func TestClearURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "http_url",
			url:  "http://www.example.com",
			want: "www.example.com",
		},
		{
			name: "https_url",
			url:  "https://www.example.com",
			want: "www.example.com",
		},
		{
			name: "no_protocol_url",
			url:  "www.example.com",
			want: "www.example.com",
		},
		{
			name: "empty_string",
			url:  "",
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			got := clearURL(tt.url)
			assert.Equal(t, tt.want, got)
		})
	}
}
func TestSendUpdates(t *testing.T) {
	errorInternal := status.Error(codes.Internal, "internal error")
	tests := []struct {
		name      string
		body      []byte
		client    func(t *testing.T) *RPCClient
		expectErr error
	}{
		{
			name: "test_valid_request",
			body: []byte("valid request"),
			client: func(t *testing.T) *RPCClient {
				ctrl := gomock.NewController(t)
				conn := NewMockRPCConnection(ctrl)
				conn.EXPECT().Invoke(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return &RPCClient{
					ctx:     context.TODO(),
					conn:    conn,
					service: pb.NewMetricsServiceClient(conn),
					netAddr: "addr",
				}
			},
			expectErr: nil,
		},
		{
			name: "error_on_invoke",
			body: []byte("valid request"),
			client: func(t *testing.T) *RPCClient {
				ctrl := gomock.NewController(t)
				conn := NewMockRPCConnection(ctrl)
				conn.EXPECT().Invoke(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errorInternal)
				return &RPCClient{
					ctx:     context.TODO(),
					conn:    conn,
					service: pb.NewMetricsServiceClient(conn),
					netAddr: "addr",
				}
			},
			expectErr: errorInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.client(t)
			_, err := client.sendUpdates(context.TODO(), tt.body)

			if tt.expectErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestCreateMeta - test cases for createMeta function
func TestCreateMeta(t *testing.T) {
	tests := []struct {
		name    string
		headers []Header
		addr    string
	}{
		{
			name: "valid_headers_with_net_addr",
			headers: []Header{
				{Name: "X-Test-Header", Value: "test1"},
				{Name: "X-Auth-Token", Value: "auth2"},
			},
			addr: "10.0.0.1",
		},
		{
			name:    "no_headers_with_netAddr",
			headers: []Header{},
			addr:    "10.0.0.2",
		},
		{
			name:    "no_headers_and_no_net_addr",
			headers: []Header{},
			addr:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := RPCClient{
				ctx:     context.Background(),
				netAddr: tt.addr,
			}
			got := client.createMeta(tt.headers)
			md, ok := metadata.FromOutgoingContext(got)
			assert.True(t, ok)
			for _, header := range tt.headers {
				assert.Equal(t, []string{header.Value}, md.Get(header.Name))
			}
			assert.Equal(t, []string{tt.addr}, md.Get("X-Real-IP"))
		})
	}
}

// TestRPCClientClose - test cases for Close function in RPCClient
func TestRPCClientClose(t *testing.T) {
	ctrl := gomock.NewController(t)
	tests := []struct {
		name      string
		closeFunc func() error // mock close function of RPCConnection
		wantErr   bool
	}{
		{
			name: "successful_close",
			closeFunc: func() error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "error_on_close",
			closeFunc: func() error {
				return errors.New("close error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := NewMockRPCConnection(ctrl)
			conn.EXPECT().Close().Return(tt.closeFunc()).Times(1)

			client := &RPCClient{
				conn: conn,
			}

			err := client.Close()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRPCClientPost - test cases for Post function in RPCClient
func TestRPCClientPost(t *testing.T) {
	errorInternal := status.Error(codes.Internal, "internal error")
	errorNoRPC := errors.New("no rpc")
	tests := []struct {
		name     string
		url      string
		body     []byte
		headers  []Header
		client   func(t *testing.T) *RPCClient
		wantErr  error
		wantCode int
	}{
		{
			name:    "valid_updates_url",
			url:     URLUpdates,
			body:    []byte("valid request"),
			headers: []Header{{Name: "Aboba", Value: "123"}},
			client: func(t *testing.T) *RPCClient {
				ctrl := gomock.NewController(t)
				conn := NewMockRPCConnection(ctrl)
				conn.EXPECT().Invoke(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return &RPCClient{
					ctx:     context.TODO(),
					conn:    conn,
					service: pb.NewMetricsServiceClient(conn),
					netAddr: "addr",
				}
			},
			wantErr:  nil,
			wantCode: http.StatusOK,
		},
		{
			name:    "invalid_url",
			url:     "invalid",
			body:    []byte("valid request"),
			headers: []Header{{Name: "Aboba", Value: "123"}},
			client: func(t *testing.T) *RPCClient {
				ctrl := gomock.NewController(t)
				conn := NewMockRPCConnection(ctrl)
				conn.EXPECT().Invoke(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return &RPCClient{
					ctx:     context.TODO(),
					conn:    conn,
					service: pb.NewMetricsServiceClient(conn),
					netAddr: "addr",
				}
			},
			wantErr:  ErrorMethodNotExists,
			wantCode: 0,
		},
		{
			name:    "returns_rpc_error",
			url:     URLUpdates,
			body:    []byte("valid request"),
			headers: []Header{{Name: "Aboba", Value: "123"}},
			client: func(t *testing.T) *RPCClient {
				ctrl := gomock.NewController(t)
				conn := NewMockRPCConnection(ctrl)
				conn.EXPECT().Invoke(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errorInternal)
				return &RPCClient{
					ctx:     context.TODO(),
					conn:    conn,
					service: pb.NewMetricsServiceClient(conn),
					netAddr: "addr",
				}
			},
			wantErr:  nil,
			wantCode: http.StatusInternalServerError,
		},
		{
			name:    "returns_no_rpc_error",
			url:     URLUpdates,
			body:    []byte("valid request"),
			headers: []Header{{Name: "Aboba", Value: "123"}},
			client: func(t *testing.T) *RPCClient {
				ctrl := gomock.NewController(t)
				conn := NewMockRPCConnection(ctrl)
				conn.EXPECT().Invoke(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errorNoRPC)
				return &RPCClient{
					ctx:     context.TODO(),
					conn:    conn,
					service: pb.NewMetricsServiceClient(conn),
					netAddr: "addr",
				}
			},
			wantErr:  errorNoRPC,
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.client(t)

			resp, err := client.Post(tt.url, tt.body, tt.headers...)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCode, resp.StatusCode(), "StatusCode should return %d", tt.wantCode)
			}
		})
	}
}

func TestNewRPCClient(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name: "create_new_rpcclient",
			url:  "localhost:8080",

			wantErr: false,
		},
		{
			name:    "return_error_when_invalid_url",
			url:     "invalid url",
			wantErr: false,
		},
		{
			name:    "return_error_when_empty_url",
			url:     "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewRPCClient(context.TODO(), tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}
