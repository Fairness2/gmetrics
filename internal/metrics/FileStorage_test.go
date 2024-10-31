package metrics

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"gmetrics/internal/contextkeys"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkFileStorage_SetMetrics(b *testing.B) {
	benchmarks := []struct {
		name     string
		gauges   map[string]Gauge
		counters map[string]Counter
	}{
		{
			name:     "one_value",
			gauges:   generateGaugesMap(1),
			counters: generateCounterMap(1),
		},
		{
			name:     "1000_value",
			gauges:   generateGaugesMap(1000),
			counters: generateCounterMap(1000),
		},
		{
			name:     "10000_value",
			gauges:   generateGaugesMap(1000),
			counters: generateCounterMap(1000),
		},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			store, _ := NewFileStorage("bench.json", false, true)
			defer func() {
				_ = store.Close()
				if rErr := os.Remove("bench.json"); rErr != nil {
					b.Errorf("Cant remove file bench.json")
				}
			}()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := store.SetGauges(bm.gauges); err != nil {
					b.Errorf("Cant set gauges: %v", err)
				}
				if err := store.AddCounters(bm.counters); err != nil {
					b.Errorf("Cant add counters: %v", err)
				}

			}
		})
	}
}

func TestNewFileStorage(t *testing.T) {
	defer func() {
		if err := os.Remove("test.json"); err != nil {
			t.Errorf("Cant remove file test.json")
		}
	}()
	memStore, err := NewFileStorage("test.json", true, true)
	assert.NoError(t, err, "Cant create file storage")
	assert.NotNil(t, memStore)
}

func TestFileStorage_SetGauge(t *testing.T) {
	defer func() {
		if err := os.Remove("test.json"); err != nil {
			t.Errorf("Cant remove file test.json")
		}
	}()
	memStore, err := NewFileStorage("test.json", true, true)
	assert.NoError(t, err, "Cant create file storage")
	if err != nil {
		return
	}
	testCases := []struct {
		name      string
		mName     string
		wantValue Gauge
	}{
		{
			name:      "set_new_gauge",
			mName:     "metricname",
			wantValue: Gauge(42.5),
		},
		{
			name:      "overwrite_gauge",
			mName:     "metricname",
			wantValue: Gauge(43.5),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := memStore.SetGauge(tc.mName, tc.wantValue)
			require.NoError(t, err)
			v, _ := memStore.GetGauge(tc.mName)
			assert.Equal(t, tc.wantValue, v)
		})
	}
}

func TestFileStorage_AddCounter(t *testing.T) {
	defer func() {
		if err := os.Remove("test.json"); err != nil {
			t.Errorf("Cant remove file test.json")
		}
	}()
	memStore, err := NewFileStorage("test.json", true, true)
	assert.NoError(t, err, "Cant create file storage")
	if err != nil {
		return
	}
	testCases := []struct {
		name      string
		mName     string
		addValue  Counter
		wantValue Counter
	}{
		{
			name:      "add_new_counter",
			mName:     "metricname",
			addValue:  Counter(5),
			wantValue: Counter(5),
		},
		{
			name:      "increment_counter",
			mName:     "metricname",
			addValue:  Counter(5),
			wantValue: Counter(10),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := memStore.AddCounter(tc.mName, tc.addValue)
			require.NoError(t, err)
			v, _ := memStore.GetCounter(tc.mName)
			assert.Equal(t, tc.wantValue, v)
		})
	}
}

func TestFileStorage_GetGauge(t *testing.T) {
	defer func() {
		if err := os.Remove("test.json"); err != nil {
			t.Errorf("Cant remove file test.json")
		}
	}()
	memStore, err := NewFileStorage("test.json", true, true)
	assert.NoError(t, err, "Cant create file storage")
	if err != nil {
		return
	}
	_ = memStore.SetGauge("temp", 42.5)

	testCases := []struct {
		name      string
		mName     string
		wantValue any
		ok        bool
	}{
		{
			name:      "get_gauge",
			mName:     "temp",
			ok:        true,
			wantValue: Gauge(42.5),
		},
		{
			name:      "get_non-existent_value",
			mName:     "load",
			ok:        false,
			wantValue: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v, ok := memStore.GetGauge(tc.mName)
			if tc.ok {
				assert.True(t, ok)
				assert.Equal(t, tc.wantValue, v)
			} else {
				assert.False(t, ok)
			}
		})
	}
}

func TestFileStorage_GetCounter(t *testing.T) {
	defer func() {
		if err := os.Remove("test.json"); err != nil {
			t.Errorf("Cant remove file test.json")
		}
	}()
	memStore, err := NewFileStorage("test.json", true, true)
	assert.NoError(t, err, "Cant create file storage")
	if err != nil {
		return
	}
	_ = memStore.AddCounter("hits", 1)

	testCases := []struct {
		name      string
		mName     string
		wantValue any
		ok        bool
	}{
		{
			name:      "get_counter",
			mName:     "hits",
			ok:        true,
			wantValue: Counter(1),
		},
		{
			name:      "get_non-existent_value",
			mName:     "load",
			ok:        false,
			wantValue: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v, ok := memStore.GetCounter(tc.mName)
			if tc.ok {
				assert.True(t, ok)
				assert.Equal(t, tc.wantValue, v)
			} else {
				assert.False(t, ok)
			}
		})
	}
}

func TestFileStorage_RestoreFromFile(t *testing.T) {
	testCases := []struct {
		name        string
		fileName    string
		isErr       bool
		hasFile     bool
		fileContent string
	}{
		{
			name:        "valid_file",
			fileName:    "test.json",
			isErr:       false,
			hasFile:     true,
			fileContent: `{"gauge":{"Alloc":3009592},"counter":{"GetSet75":2210657517}}`,
		},
		{
			name:        "invalid_file",
			fileName:    "non_existing.json",
			isErr:       true,
			hasFile:     true,
			fileContent: `{"gauge":{"Alloc":3009592},"counter":{"GetSet75":2210657517},}`,
		},
		{
			name:     "no_file",
			fileName: "non_existing.json",
			isErr:    true,
			hasFile:  false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if tc.fileName != "non_existing.json" {
					if err := os.Remove(tc.fileName); err != nil {
						t.Errorf("Cant remove file %s", tc.fileName)
					}
				}
			}()
			if tc.hasFile {
				_ = os.WriteFile(tc.fileName, []byte(tc.fileContent), 0644)
			}
			memStore := DurationFileStorage{IStorage: NewMemStorage()}
			err := restoreFromFile(tc.fileName, memStore.IStorage)
			if tc.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFileStorage_AddCounters(t *testing.T) {
	defer func() {
		if err := os.Remove("test.json"); err != nil {
			t.Errorf("Cant remove file test.json")
		}
	}()
	store, err := NewFileStorage("test.json", true, true)
	assert.NoError(t, err, "Cant create file storage")
	if err != nil {
		return
	}
	testCases := []struct {
		name        string
		counters    map[string]Counter
		wantErr     bool
		wantCounter map[string]Counter
	}{
		{
			name: "add_new_counters",
			counters: map[string]Counter{
				"counter1": 1,
				"counter2": 2,
			},
			wantErr: false,
			wantCounter: map[string]Counter{
				"counter1": 1,
				"counter2": 2,
			},
		},
		{
			name: "add_existing_counters",
			counters: map[string]Counter{
				"counter1": 3,
				"counter2": 4,
			},
			wantErr: false,
			wantCounter: map[string]Counter{
				"counter1": 4,
				"counter2": 6,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := store.AddCounters(tc.counters)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for k, v := range tc.wantCounter {
					v1, vErr := store.GetCounter(k)
					assert.True(t, vErr, "Cant get counter %s", k)
					assert.Equal(t, v, v1)
				}
			}
		})
	}
}

func TestFileStorage_SetGauges(t *testing.T) {
	defer func() {
		if err := os.Remove("test.json"); err != nil {
			t.Errorf("Cant remove file test.json")
		}
	}()
	store, err := NewFileStorage("test.json", true, true)
	assert.NoError(t, err, "Cant create file storage")
	if err != nil {
		return
	}
	testCases := []struct {
		name       string
		gauges     map[string]Gauge
		wantErr    bool
		wantGauges map[string]Gauge
	}{
		{
			name: "set_new_gauges",
			gauges: map[string]Gauge{
				"gauge1": 1,
				"gauge2": 2,
			},
			wantErr: false,
			wantGauges: map[string]Gauge{
				"gauge1": 1,
				"gauge2": 2,
			},
		},
		{
			name: "set_existing_counters",
			gauges: map[string]Gauge{
				"gauge1": 3,
				"gauge2": 4,
			},
			wantErr: false,
			wantGauges: map[string]Gauge{
				"gauge1": 3,
				"gauge2": 4,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := store.SetGauges(tc.gauges)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for k, v := range tc.gauges {
					v1, vErr := store.GetGauge(k)
					assert.True(t, vErr, "Cant get gauge %s", k)
					assert.Equal(t, v, v1)
				}
			}
		})
	}
}

func TestDurationFileStorage_FlushAndClose(t *testing.T) {
	flushError := errors.New("error on flush")
	closeError := errors.New("error on close")
	testCases := []struct {
		name             string
		withErrorOnFlush bool
		withErrorOnClose bool
		wantError        error
	}{
		{
			name:             "success_on_flush_and_close",
			withErrorOnFlush: false,
			withErrorOnClose: false,
			wantError:        nil,
		},
		{
			name:             "error_on_flush",
			withErrorOnFlush: true,
			withErrorOnClose: false,
			wantError:        flushError,
		},
		{
			name:             "error_on_close",
			withErrorOnFlush: false,
			withErrorOnClose: true,
			wantError:        closeError,
		},
		{
			name:             "error_on_flush_and_close",
			withErrorOnFlush: true,
			withErrorOnClose: true,
			wantError:        flushError,
		},
	}
	ctrl := gomock.NewController(t)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockWriter := NewMockWriter(ctrl)
			if tc.withErrorOnClose {
				mockWriter.EXPECT().Close().Return(closeError).AnyTimes()
			} else {
				mockWriter.EXPECT().Close().Return(nil).AnyTimes()
			}
			if tc.withErrorOnFlush {
				mockWriter.EXPECT().Write(gomock.Any()).Return(flushError).AnyTimes()
			} else {
				mockWriter.EXPECT().Write(gomock.Any()).Return(nil).AnyTimes()
			}
			durationFileStorage := &DurationFileStorage{
				IStorage: NewMemStorage(),
				writer:   mockWriter,
				syncMode: false,
			}
			err := durationFileStorage.FlushAndClose()
			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError, "Want error %s, got %s", tc.wantError, err)
			} else {
				assert.NoError(t, err, "Want no error, got %s", err)
			}
		})
	}
}

func TestDurationFileStorage_Close(t *testing.T) {
	closeError := errors.New("error on close")
	testCases := []struct {
		name             string
		withErrorOnClose bool
		wantError        error
	}{
		{
			name:             "success_on_close",
			withErrorOnClose: false,
			wantError:        nil,
		},
		{
			name:             "error_on_close",
			withErrorOnClose: true,
			wantError:        closeError,
		},
	}
	ctrl := gomock.NewController(t)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockWriter := NewMockWriter(ctrl)
			if tc.withErrorOnClose {
				mockWriter.EXPECT().Close().Return(closeError).AnyTimes()
			} else {
				mockWriter.EXPECT().Close().Return(nil).AnyTimes()
			}
			durationFileStorage := &DurationFileStorage{
				IStorage: NewMemStorage(),
				writer:   mockWriter,
				syncMode: false,
			}
			err := durationFileStorage.Close()
			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError, "Want error %s, got %s", tc.wantError, err)
			} else {
				assert.NoError(t, err, "Want no error, got %s", err)
			}
		})
	}
}

func TestDurationFileStorage_IsSyncMode(t *testing.T) {
	testCases := []struct {
		name         string
		syncMode     bool
		expectResult bool
	}{
		{
			name:         "sync_mode_enabled",
			syncMode:     true,
			expectResult: true,
		},
		{
			name:         "sync_mode_disabled",
			syncMode:     false,
			expectResult: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			memStore := DurationFileStorage{
				IStorage: NewMemStorage(),
				writer:   NewMockWriter(gomock.NewController(t)),
				syncMode: tc.syncMode,
			}
			assert.Equal(t, tc.expectResult, memStore.IsSyncMode())
		})
	}
}

func TestDurationFileStorage_Flush(t *testing.T) {
	flushError := errors.New("error on flush")
	testCases := []struct {
		name             string
		withErrorOnFlush bool
		wantError        error
	}{
		{
			name:             "success_on_flush_and_close",
			withErrorOnFlush: false,
			wantError:        nil,
		},
		{
			name:             "error_on_flush",
			withErrorOnFlush: true,
			wantError:        flushError,
		},
	}
	ctrl := gomock.NewController(t)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockWriter := NewMockWriter(ctrl)
			if tc.withErrorOnFlush {
				mockWriter.EXPECT().Write(gomock.Any()).Return(flushError).AnyTimes()
			} else {
				mockWriter.EXPECT().Write(gomock.Any()).Return(nil).AnyTimes()
			}
			durationFileStorage := &DurationFileStorage{
				IStorage: NewMemStorage(),
				writer:   mockWriter,
				syncMode: false,
			}
			err := durationFileStorage.Flush()
			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError, "Want error %s, got %s", tc.wantError, err)
			} else {
				assert.NoError(t, err, "Want no error, got %s", err)
			}
		})
	}
}
func TestDurationFileStorage_Sync(t *testing.T) {
	ctrl := gomock.NewController(t)
	flushError := errors.New("error on flush")
	testCases := []struct {
		name               string
		asyncInterval      int64
		mockFunc           func(mockWriter *MockWriter)
		checkCancelContext bool
		wantError          error
	}{
		{
			name:          "success_on_sync",
			asyncInterval: 1,
			mockFunc: func(mockWriter *MockWriter) {
				mockWriter.EXPECT().Write(gomock.Any()).Return(nil).AnyTimes()
			},
			checkCancelContext: false,
		},
		{
			name:          "sync_cancelled_early",
			asyncInterval: 2,
			mockFunc: func(mockWriter *MockWriter) {
				mockWriter.EXPECT().Write(gomock.Any()).Return(nil).AnyTimes()
			},
			checkCancelContext: true,
		},
		{
			name:          "error_on_sync",
			asyncInterval: 1,
			mockFunc: func(mockWriter *MockWriter) {
				mockWriter.EXPECT().Write(gomock.Any()).Return(flushError).AnyTimes()
			},
			checkCancelContext: false,
			wantError:          flushError,
		},
		{
			name:          "sync_cancelled_early_with_error",
			asyncInterval: 2,
			mockFunc: func(mockWriter *MockWriter) {
				mockWriter.EXPECT().Write(gomock.Any()).Return(flushError).AnyTimes()
			},
			checkCancelContext: true,
			wantError:          flushError,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockWriter := NewMockWriter(ctrl)
			tc.mockFunc(mockWriter)
			durationFileStorage := &DurationFileStorage{
				IStorage: NewMemStorage(),
				writer:   mockWriter,
				syncMode: false,
			}

			ctx := context.TODO()
			ctx = context.WithValue(ctx, contextkeys.SyncInterval, tc.asyncInterval)
			start, cancel := context.WithCancel(ctx)
			go func() {
				<-time.After(time.Second * 2)
				cancel()
			}()
			if tc.checkCancelContext {
				go func() {
					<-time.After(time.Second * 1)
					cancel()
				}()
			}
			err := durationFileStorage.Sync(start)
			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError, "Want error %s, got %s", tc.wantError, err)

			} else {
				assert.NoError(t, err, "Want no error, got %s", err)
			}
		})
	}
}
