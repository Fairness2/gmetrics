package sendpool

import (
	"context"
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gmetrics/internal/metrics"
	"gmetrics/internal/payload"
	"net/http"
	"sync"
	"testing"
)

func Benchmark(b *testing.B) {
	ctrl := gomock.NewController(b)
	restClient := NewMockIClient(ctrl)
	restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&resty.Response{RawResponse: &http.Response{StatusCode: 200}}, nil).
		AnyTimes()
	restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
	p, err := NewWithClient(context.TODO(), 2, "aboba", restClient, nil)
	if err != nil {
		assert.NoError(b, err)
		return
	}
	var intValue int64 = 64
	floatValue := 64.64
	body := []payload.Metrics{
		{
			ID:    "PollCount",
			MType: metrics.TypeCounter,
			Delta: &intValue,
		},
		{
			ID:    "GaugeVal",
			MType: metrics.TypeGauge,
			Value: &floatValue,
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, sErr := p.Send(body); sErr != nil {
			b.Errorf("Send() error = %v", sErr)
		}
	}
}

// TestPoolCompressBody tests the method compressBody of the Pool struct.
func TestPoolCompressBody(t *testing.T) {
	tests := []struct {
		name    string
		body    []byte
		wantErr bool
	}{
		{
			name:    "normal_case",
			body:    []byte("test"),
			wantErr: false,
		},
		{
			name:    "empty_body",
			body:    []byte(""),
			wantErr: false,
		},
		{
			name:    "nil_body",
			body:    nil,
			wantErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &Pool{
				encodeWriterPool: sync.Pool{
					New: newEncoder,
				},
			}
			_, err := p.compressBody(tc.body)
			if (err != nil) != tc.wantErr {
				t.Errorf("Pool.compressBody() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

// TestPoolHashBody tests the method hashBody of the Pool struct.
func TestPoolHashBody(t *testing.T) {
	tests := []struct {
		name    string
		body    []byte
		hashKey string
		wantErr bool
	}{
		{
			name:    "valid_hash_key_and_body",
			body:    []byte("test"),
			hashKey: "secret",
			wantErr: false,
		},
		{
			name:    "empty_hash_key",
			body:    []byte("test"),
			hashKey: "",
			wantErr: true,
		},
		{
			name:    "nil_body",
			body:    nil,
			hashKey: "secret",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pool{
				HashKey: tt.hashKey,
			}
			_, err := p.hashBody(tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("Pool.hashBody() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPoolMarshalBody tests the method marshalBody of the Pool struct.
func TestPoolMarshalBody(t *testing.T) {
	tests := []struct {
		name    string
		body    func() []payload.Metrics
		wantErr bool
	}{
		{
			name: "normal_case",
			body: func() []payload.Metrics {
				var intValue int64 = 64
				floatValue := 64.64
				return []payload.Metrics{
					{
						ID:    "PollCount",
						MType: metrics.TypeCounter,
						Delta: &intValue,
					},
					{
						ID:    "GaugeVal",
						MType: metrics.TypeGauge,
						Value: &floatValue,
					},
				}
			},
			wantErr: false,
		},
		{
			name: "empty_metrics",
			body: func() []payload.Metrics {
				return []payload.Metrics{}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pool{}
			_, err := p.marshalBody(tt.body())
			if (err != nil) != tt.wantErr {
				t.Errorf("Pool.marshalBody() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestPool_sendToServer tests the method sendToServer of the Pool struct.
func TestPool_sendToServer(t *testing.T) {
	httpError := errors.New("http error")
	ctrl := gomock.NewController(t)
	tests := []struct {
		name         string
		body         func() []payload.Metrics
		hashKey      string
		err          error
		restClient   func() *MockIClient
		resultStatus int
	}{
		{
			name: "normal_case",
			body: func() []payload.Metrics {
				var intValue int64 = 64
				floatValue := 64.64
				return []payload.Metrics{
					{
						ID:    "PollCount",
						MType: metrics.TypeCounter,
						Delta: &intValue,
					},
					{
						ID:    "GaugeVal",
						MType: metrics.TypeGauge,
						Value: &floatValue,
					},
				}
			},
			hashKey: "secret",
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 200}}, nil).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 200,
		},
		/*{
			name: "invalid_hash_key",
			body: func() []payload.Metrics {
				return []payload.Metrics{}
			},
			hashKey: "",
			err:     ErrorEmptyHashKey,
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 200}}, nil).
					AnyTimes()
				return restClient
			},
			resultStatus: 200,
		},*/
		{
			name: "empty_metrics",
			body: func() []payload.Metrics {
				return []payload.Metrics{}
			},
			hashKey: "secret",
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 200}}, nil).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 200,
		},
		{
			name: "http_error",
			body: func() []payload.Metrics {
				return []payload.Metrics{}
			},
			hashKey: "secret",
			err:     httpError,
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 200}}, httpError).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 200,
		},
		{
			name: "http_error",
			body: func() []payload.Metrics {
				return []payload.Metrics{}
			},
			hashKey: "secret",
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 400}}, nil).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pool{
				encodeWriterPool: sync.Pool{
					New: newEncoder,
				},
				client:  tt.restClient(),
				HashKey: tt.hashKey,
			}
			resp, err := p.sendToServer(tt.body())
			if tt.err != nil {
				assert.NotNil(t, err)
				assert.ErrorIs(t, tt.err, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.resultStatus, resp.StatusCode())
			}
		})
	}
}

// TestSendToServer tests the method sendToServer of the Pool struct.
func TestPool_processRequest(t *testing.T) {
	httpError := errors.New("http error")
	ctrl := gomock.NewController(t)
	tests := []struct {
		name         string
		body         func() *poolPayload
		hashKey      string
		err          error
		restClient   func() *MockIClient
		resultStatus int
	}{
		{
			name: "normal_case",
			body: func() *poolPayload {
				var intValue int64 = 64
				floatValue := 64.64
				out := make(chan response)
				return &poolPayload{Body: []payload.Metrics{
					{
						ID:    "PollCount",
						MType: metrics.TypeCounter,
						Delta: &intValue,
					},
					{
						ID:    "GaugeVal",
						MType: metrics.TypeGauge,
						Value: &floatValue,
					},
				}, Out: out}
			},
			hashKey: "secret",
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 200}}, nil).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 200,
		},
		{
			name: "empty_metrics",
			body: func() *poolPayload {
				out := make(chan response)
				return &poolPayload{Body: []payload.Metrics{}, Out: out}
			},
			hashKey: "secret",
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 200}}, nil).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 200,
		},
		{
			name: "http_error",
			body: func() *poolPayload {
				out := make(chan response)
				return &poolPayload{Body: []payload.Metrics{}, Out: out}
			},
			hashKey: "secret",
			err:     httpError,
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 200}}, httpError).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 200,
		},
		{
			name: "http_error",
			body: func() *poolPayload {
				out := make(chan response)
				return &poolPayload{Body: []payload.Metrics{}, Out: out}
			},
			hashKey: "secret",
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 400}}, nil).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pool{
				encodeWriterPool: sync.Pool{
					New: newEncoder,
				},
				client:  tt.restClient(),
				HashKey: tt.hashKey,
			}
			body := tt.body()
			go p.processRequest(body)
			res := <-body.Out
			if tt.err != nil {
				assert.NotNil(t, res.Err)
				assert.ErrorIs(t, tt.err, res.Err)
			} else {
				assert.NoError(t, res.Err)
				assert.Equal(t, tt.resultStatus, res.Res.StatusCode())
			}
		})
	}
}

func TestPool_worker(t *testing.T) {
	httpError := errors.New("http error")
	ctrl := gomock.NewController(t)
	tests := []struct {
		name         string
		body         func() *poolPayload
		hashKey      string
		err          error
		restClient   func() *MockIClient
		resultStatus int
		cancel       bool
	}{
		{
			name: "normal_case",
			body: func() *poolPayload {
				var intValue int64 = 64
				floatValue := 64.64
				out := make(chan response)
				return &poolPayload{Body: []payload.Metrics{
					{
						ID:    "PollCount",
						MType: metrics.TypeCounter,
						Delta: &intValue,
					},
					{
						ID:    "GaugeVal",
						MType: metrics.TypeGauge,
						Value: &floatValue,
					},
				}, Out: out}
			},
			hashKey: "secret",
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 200}}, nil).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 200,
		},
		{
			name: "empty_metrics",
			body: func() *poolPayload {
				out := make(chan response)
				return &poolPayload{Body: []payload.Metrics{}, Out: out}
			},
			hashKey: "secret",
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 200}}, nil).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 200,
		},
		{
			name: "http_error",
			body: func() *poolPayload {
				out := make(chan response)
				return &poolPayload{Body: []payload.Metrics{}, Out: out}
			},
			hashKey: "secret",
			err:     httpError,
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 200}}, httpError).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 200,
		},
		{
			name: "http_error",
			body: func() *poolPayload {
				out := make(chan response)
				return &poolPayload{Body: []payload.Metrics{}, Out: out}
			},
			hashKey: "secret",
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 400}}, nil).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 400,
		},
		{
			name: "cancel_context",
			body: func() *poolPayload {
				out := make(chan response)
				return &poolPayload{Body: []payload.Metrics{}, Out: out}
			},
			hashKey: "secret",
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 400}}, nil).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 400,
			cancel:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := &Pool{
				encodeWriterPool: sync.Pool{
					New: newEncoder,
				},
				client:  tt.restClient(),
				HashKey: tt.hashKey,
				in:      make(chan *poolPayload),
			}
			pool.wg.Add(1)
			body := tt.body()
			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()
			if tt.cancel {
				cancel()
				pool.worker(ctx)
				assert.True(t, true)
				return
			}
			go func() {
				pool.in <- body
			}()
			go pool.worker(ctx)
			res := <-body.Out
			if tt.err != nil {
				assert.NotNil(t, res.Err)
				assert.ErrorIs(t, tt.err, res.Err)
			} else {
				assert.NoError(t, res.Err)
				assert.Equal(t, tt.resultStatus, res.Res.StatusCode())
			}
		})
	}
}

func TestPool_Send(t *testing.T) {
	httpError := errors.New("http error")
	ctrl := gomock.NewController(t)
	tests := []struct {
		name         string
		body         func() []payload.Metrics
		hashKey      string
		err          error
		restClient   func() *MockIClient
		resultStatus int
		cancel       bool
		closed       bool
	}{
		{
			name: "normal_case",
			body: func() []payload.Metrics {
				var intValue int64 = 64
				floatValue := 64.64
				return []payload.Metrics{
					{
						ID:    "PollCount",
						MType: metrics.TypeCounter,
						Delta: &intValue,
					},
					{
						ID:    "GaugeVal",
						MType: metrics.TypeGauge,
						Value: &floatValue,
					},
				}
			},
			hashKey: "secret",
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 200}}, nil).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 200,
		},
		{
			name: "empty_metrics",
			body: func() []payload.Metrics {
				return []payload.Metrics{}
			},
			hashKey: "secret",
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 200}}, nil).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 200,
		},
		{
			name: "http_error",
			body: func() []payload.Metrics {
				return []payload.Metrics{}
			},
			hashKey: "secret",
			err:     httpError,
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 200}}, httpError).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 200,
		},
		{
			name: "http_error",
			body: func() []payload.Metrics {
				return []payload.Metrics{}
			},
			hashKey: "secret",
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 400}}, nil).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 400,
		},
		{
			name: "cancel_context",
			body: func() []payload.Metrics {
				return []payload.Metrics{}
			},
			hashKey: "secret",
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 400}}, nil).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 400,
			cancel:       true,
		},
		{
			name: "pool_closed",
			body: func() []payload.Metrics {
				return []payload.Metrics{}
			},
			hashKey: "secret",
			restClient: func() *MockIClient {
				restClient := NewMockIClient(ctrl)
				restClient.EXPECT().Post(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&resty.Response{RawResponse: &http.Response{StatusCode: 400}}, nil).
					AnyTimes()
				restClient.EXPECT().EnableManualCompression().Return(true).AnyTimes()
				return restClient
			},
			resultStatus: 400,
			cancel:       true,
			closed:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := &Pool{
				encodeWriterPool: sync.Pool{
					New: newEncoder,
				},
				client:   tt.restClient(),
				HashKey:  tt.hashKey,
				in:       make(chan *poolPayload),
				isClosed: tt.closed,
			}
			if tt.closed {
				res, err := pool.Send(tt.body())
				assert.Nil(t, res)
				assert.ErrorIs(t, err, ErrorPoolIsClosed)
				return
			}
			pool.wg.Add(1)
			body := tt.body()
			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()
			if tt.cancel {
				cancel()
				pool.worker(ctx)
				assert.True(t, true)
				return
			}
			go pool.worker(ctx)

			res, err := pool.Send(body)
			if tt.err != nil {
				assert.NotNil(t, err)
				assert.ErrorIs(t, tt.err, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.resultStatus, res.StatusCode())
			}
		})
	}
}
