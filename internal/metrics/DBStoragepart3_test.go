package metrics

import (
	"context"
	"database/sql"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDBStorage_GetGauge(t *testing.T) {
	errorPGConnection := &pgconn.PgError{Code: pgerrcode.ConnectionException}
	errorScan := errors.New("errorScan")
	testCases := []struct {
		name          string
		success       bool
		getExecutor   func(t *testing.T) SQLExecutor
		close         bool
		expectedGauge Gauge
		gaugeName     string
		getStorage    func(t *testing.T) IStorage
	}{
		{
			name:    "storage_close",
			success: false,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return executor
			},
			expectedGauge: Gauge(0),
			close:         true,
			gaugeName:     "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauge(gomock.Any()).Return(Gauge(9), false).AnyTimes()
				return store
			},
		},
		{
			name:    "success_form_store",
			success: true,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				row := NewMockIRows(ctrl)
				row.EXPECT().Scan(gomock.Any()).Return(nil).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Gauge); ok {
							*ptr = Gauge(12)
						}
					}
				}).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(row).AnyTimes()
				return executor
			},
			expectedGauge: Gauge(12),
			gaugeName:     "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauge(gomock.Any()).Return(Gauge(12), true).AnyTimes()
				return store
			},
		},
		{
			name:    "success_form_bd",
			success: true,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				row := NewMockIRows(ctrl)
				row.EXPECT().Scan(gomock.Any()).Return(nil).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Gauge); ok {
							*ptr = Gauge(12)
						}
					}
				}).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(row).AnyTimes()
				return executor
			},
			expectedGauge: Gauge(12),
			gaugeName:     "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauge(gomock.Any()).Return(Gauge(9), false).AnyTimes()
				return store
			},
		},
		{
			name:    "error_pg",
			success: false,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				row := NewMockIRows(ctrl)
				row.EXPECT().Scan(gomock.Any()).Return(errorPGConnection).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Gauge); ok {
							*ptr = Gauge(12)
						}
					}
				}).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(row).AnyTimes()
				return executor
			},
			expectedGauge: Gauge(12),
			gaugeName:     "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauge(gomock.Any()).Return(Gauge(9), false).AnyTimes()
				return store
			},
		},
		{
			name:    "error_no_rows",
			success: false,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				row := NewMockIRows(ctrl)
				row.EXPECT().Scan(gomock.Any()).Return(sql.ErrNoRows).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Gauge); ok {
							*ptr = Gauge(12)
						}
					}
				}).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(row).AnyTimes()
				return executor
			},
			expectedGauge: Gauge(12),
			gaugeName:     "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauge(gomock.Any()).Return(Gauge(9), false).AnyTimes()
				return store
			},
		},
		{
			name:    "error_pg_after_success",
			success: true,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				row := NewMockIRows(ctrl)
				first := row.EXPECT().Scan(gomock.Any()).Return(errorPGConnection).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Gauge); ok {
							*ptr = Gauge(12)
						}
					}
				}).Times(1)
				row.EXPECT().Scan(gomock.Any()).Return(nil).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Gauge); ok {
							*ptr = Gauge(12)
						}
					}
				}).After(first)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(row).AnyTimes()
				return executor
			},
			expectedGauge: Gauge(12),
			gaugeName:     "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauge(gomock.Any()).Return(Gauge(9), false).AnyTimes()
				return store
			},
		},
		{
			name:    "error_pg_after_scan_error",
			success: false,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				row := NewMockIRows(ctrl)
				first := row.EXPECT().Scan(gomock.Any()).Return(errorPGConnection).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Gauge); ok {
							*ptr = Gauge(12)
						}
					}
				}).Times(1)
				row.EXPECT().Scan(gomock.Any()).Return(errorScan).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Gauge); ok {
							*ptr = Gauge(12)
						}
					}
				}).After(first)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(row).AnyTimes()
				return executor
			},
			expectedGauge: Gauge(12),
			gaugeName:     "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauge(gomock.Any()).Return(Gauge(9), false).AnyTimes()
				return store
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dbStorage := DBStorage{
				IStorage: tc.getStorage(t),
				storeCtx: context.Background(),
				db:       tc.getExecutor(t),
				close:    tc.close,
			}
			res, ok := dbStorage.GetGauge(tc.gaugeName)
			if !tc.success {
				assert.False(t, ok)
			} else {
				assert.True(t, ok)
				assert.Equal(t, tc.expectedGauge, res)
			}
		})
	}
}

func TestDBStorage_GetCounter(t *testing.T) {
	errorPGConnection := &pgconn.PgError{Code: pgerrcode.ConnectionException}
	errorScan := errors.New("errorScan")
	testCases := []struct {
		name            string
		success         bool
		getExecutor     func(t *testing.T) SQLExecutor
		close           bool
		expectedCounter Counter
		counterName     string
		getStorage      func(t *testing.T) IStorage
	}{
		{
			name:    "storage_close",
			success: false,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return executor
			},
			expectedCounter: Counter(0),
			close:           true,
			counterName:     "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetCounter(gomock.Any()).Return(Counter(9), false).AnyTimes()
				return store
			},
		},
		{
			name:    "success_form_store",
			success: true,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				row := NewMockIRows(ctrl)
				row.EXPECT().Scan(gomock.Any()).Return(nil).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Counter); ok {
							*ptr = Counter(12)
						}
					}
				}).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(row).AnyTimes()
				return executor
			},
			expectedCounter: Counter(12),
			counterName:     "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetCounter(gomock.Any()).Return(Counter(12), true).AnyTimes()
				return store
			},
		},
		{
			name:    "success_form_bd",
			success: true,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				row := NewMockIRows(ctrl)
				row.EXPECT().Scan(gomock.Any()).Return(nil).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Counter); ok {
							*ptr = Counter(12)
						}
					}
				}).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(row).AnyTimes()
				return executor
			},
			expectedCounter: Counter(12),
			counterName:     "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetCounter(gomock.Any()).Return(Counter(9), false).AnyTimes()
				return store
			},
		},
		{
			name:    "error_pg",
			success: false,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				row := NewMockIRows(ctrl)
				row.EXPECT().Scan(gomock.Any()).Return(errorPGConnection).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Counter); ok {
							*ptr = Counter(12)
						}
					}
				}).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(row).AnyTimes()
				return executor
			},
			expectedCounter: Counter(12),
			counterName:     "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetCounter(gomock.Any()).Return(Counter(9), false).AnyTimes()
				return store
			},
		},
		{
			name:    "error_no_rows",
			success: false,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				row := NewMockIRows(ctrl)
				row.EXPECT().Scan(gomock.Any()).Return(sql.ErrNoRows).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Counter); ok {
							*ptr = Counter(12)
						}
					}
				}).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(row).AnyTimes()
				return executor
			},
			expectedCounter: Counter(12),
			counterName:     "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetCounter(gomock.Any()).Return(Counter(9), false).AnyTimes()
				return store
			},
		},
		{
			name:    "error_pg_after_success",
			success: true,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				row := NewMockIRows(ctrl)
				first := row.EXPECT().Scan(gomock.Any()).Return(errorPGConnection).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Counter); ok {
							*ptr = Counter(12)
						}
					}
				}).Times(1)
				row.EXPECT().Scan(gomock.Any()).Return(nil).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Counter); ok {
							*ptr = Counter(12)
						}
					}
				}).After(first)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(row).AnyTimes()
				return executor
			},
			expectedCounter: Counter(12),
			counterName:     "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetCounter(gomock.Any()).Return(Counter(9), false).AnyTimes()
				return store
			},
		},
		{
			name:    "error_pg_after_scan_error",
			success: false,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				row := NewMockIRows(ctrl)
				first := row.EXPECT().Scan(gomock.Any()).Return(errorPGConnection).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Counter); ok {
							*ptr = Counter(12)
						}
					}
				}).Times(1)
				row.EXPECT().Scan(gomock.Any()).Return(errorScan).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Counter); ok {
							*ptr = Counter(12)
						}
					}
				}).After(first)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(row).AnyTimes()
				return executor
			},
			expectedCounter: Counter(12),
			counterName:     "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetCounter(gomock.Any()).Return(Counter(9), false).AnyTimes()
				return store
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dbStorage := DBStorage{
				IStorage: tc.getStorage(t),
				storeCtx: context.Background(),
				db:       tc.getExecutor(t),
				close:    tc.close,
			}
			res, ok := dbStorage.GetCounter(tc.counterName)
			if !tc.success {
				assert.False(t, ok)
			} else {
				assert.True(t, ok)
				assert.Equal(t, tc.expectedCounter, res)
			}
		})
	}
}
