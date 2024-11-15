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

func TestDBStorage_addCounter(t *testing.T) {
	execError := errors.New("execError")
	testCases := []struct {
		name         string
		wantError    error
		getExecutor  func(t *testing.T) SQLExecutor
		counterValue Counter
		counterName  string
		getStorage   func(t *testing.T) IStorage
		syncMode     bool
	}{
		{
			name:      "success",
			wantError: nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return executor
			},
			counterValue: Counter(12),
			counterName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				return store
			},
			syncMode: true,
		},
		{
			name:      "error",
			wantError: execError,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, execError).AnyTimes()
				return executor
			},
			counterValue: Counter(12),
			counterName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
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
			}
			err := dbStorage.addCounter(tc.counterName, tc.counterValue)
			if tc.wantError != nil {
				assert.ErrorIs(t, tc.wantError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDBStorage_syncCounter(t *testing.T) {
	errorPGConnection := &pgconn.PgError{Code: pgerrcode.ConnectionException}
	execError := errors.New("execError")
	testCases := []struct {
		name         string
		wantError    error
		getExecutor  func(t *testing.T) SQLExecutor
		close        bool
		counterValue Counter
		counterName  string
		getStorage   func(t *testing.T) IStorage
	}{
		{
			name:      "storage_close",
			wantError: ErrorStorageDatabaseClosed,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				return executor
			},
			counterValue: Counter(0),
			close:        true,
			counterName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				return store
			},
		},
		{
			name:      "success_form_store",
			wantError: nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return executor
			},
			counterValue: Counter(12),
			counterName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				return store
			},
		},
		{
			name:      "exec_error",
			wantError: execError,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, execError).AnyTimes()
				return executor
			},
			counterValue: Counter(12),
			counterName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				return store
			},
		},
		{
			name:      "pg_error",
			wantError: errorPGConnection,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).AnyTimes()
				return executor
			},
			counterValue: Counter(12),
			counterName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				return store
			},
		},
		{
			name:      "pg_and_then_exec_error_error",
			wantError: execError,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				first := executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).Times(1)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, execError).After(first)
				return executor
			},
			counterValue: Counter(12),
			counterName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
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
			err := dbStorage.syncCounter(tc.counterName, tc.counterValue)
			if tc.wantError != nil {
				assert.ErrorIs(t, tc.wantError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDBStorage_AddCounter(t *testing.T) {
	errorPGConnection := &pgconn.PgError{Code: pgerrcode.ConnectionException}
	execError := errors.New("execError")
	execStore := errors.New("execStore")
	testCases := []struct {
		name         string
		wantError    error
		getExecutor  func(t *testing.T) SQLExecutor
		close        bool
		counterValue Counter
		counterName  string
		getStorage   func(t *testing.T) IStorage
		syncMode     bool
	}{
		{
			name:      "storage_close",
			wantError: ErrorStorageDatabaseClosed,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				return executor
			},
			counterValue: Counter(0),
			close:        true,
			counterName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().AddCounter(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return store
			},
			syncMode: true,
		},
		{
			name:      "success_form_store",
			wantError: nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return executor
			},
			counterValue: Counter(12),
			counterName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().AddCounter(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return store
			},
			syncMode: true,
		},
		{
			name:      "exec_error",
			wantError: execError,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, execError).AnyTimes()
				return executor
			},
			counterValue: Counter(12),
			counterName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().AddCounter(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return store
			},
			syncMode: true,
		},
		{
			name:      "pg_error",
			wantError: errorPGConnection,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).AnyTimes()
				return executor
			},
			counterValue: Counter(12),
			counterName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().AddCounter(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return store
			},
			syncMode: true,
		},
		{
			name:      "pg_and_then_exec_error_error",
			wantError: execError,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				first := executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).Times(1)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, execError).After(first)
				return executor
			},
			counterValue: Counter(12),
			counterName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().AddCounter(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return store
			},
			syncMode: true,
		},
		{
			name:      "pg_and_then_exec_error_error",
			wantError: execStore,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				//first := executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).Times(1)
				//.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, execError).After(first)
				return executor
			},
			counterValue: Counter(12),
			counterName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().AddCounter(gomock.Any(), gomock.Any()).Return(execStore).AnyTimes()
				return store
			},
			syncMode: true,
		},
		{
			name:      "succes_not_sync_mode",
			wantError: nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				//first := executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).Times(1)
				//.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, execError).After(first)
				return executor
			},
			counterValue: Counter(12),
			counterName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().AddCounter(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return store
			},
			syncMode: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dbStorage := DBStorage{
				IStorage: tc.getStorage(t),
				storeCtx: context.Background(),
				db:       tc.getExecutor(t),
				close:    tc.close,
				syncMode: tc.syncMode,
			}
			err := dbStorage.AddCounter(tc.counterName, tc.counterValue)
			if tc.wantError != nil {
				assert.ErrorIs(t, tc.wantError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDBStorage_setGauge(t *testing.T) {
	execError := errors.New("execError")
	testCases := []struct {
		name        string
		wantError   error
		getExecutor func(t *testing.T) SQLExecutor
		gaugeValue  Gauge
		gaugeName   string
		getStorage  func(t *testing.T) IStorage
	}{
		{
			name:      "success",
			wantError: nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return executor
			},
			gaugeValue: Gauge(12),
			gaugeName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				return store
			},
		},
		{
			name:      "error",
			wantError: execError,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, execError).AnyTimes()
				return executor
			},
			gaugeValue: Gauge(12),
			gaugeName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
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
			}
			err := dbStorage.setGauge(tc.gaugeName, tc.gaugeValue)
			if tc.wantError != nil {
				assert.ErrorIs(t, tc.wantError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDBStorage_syncGauge(t *testing.T) {
	errorPGConnection := &pgconn.PgError{Code: pgerrcode.ConnectionException}
	execError := errors.New("execError")
	testCases := []struct {
		name        string
		wantError   error
		getExecutor func(t *testing.T) SQLExecutor
		close       bool
		gaugeValue  Gauge
		gaugeName   string
		getStorage  func(t *testing.T) IStorage
	}{
		{
			name:      "storage_close",
			wantError: ErrorStorageDatabaseClosed,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				return executor
			},
			gaugeValue: Gauge(0),
			close:      true,
			gaugeName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				return store
			},
		},
		{
			name:      "success_form_store",
			wantError: nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return executor
			},
			gaugeValue: Gauge(12),
			gaugeName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				return store
			},
		},
		{
			name:      "exec_error",
			wantError: execError,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, execError).AnyTimes()
				return executor
			},
			gaugeValue: Gauge(12),
			gaugeName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				return store
			},
		},
		{
			name:      "pg_error",
			wantError: errorPGConnection,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).AnyTimes()
				return executor
			},
			gaugeValue: Gauge(12),
			gaugeName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				return store
			},
		},
		{
			name:      "pg_and_then_exec_error_error",
			wantError: execError,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				first := executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).Times(1)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, execError).After(first)
				return executor
			},
			gaugeValue: Gauge(12),
			gaugeName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
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
			err := dbStorage.syncGauge(tc.gaugeName, tc.gaugeValue)
			if tc.wantError != nil {
				assert.ErrorIs(t, tc.wantError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDBStorage_SetGauge(t *testing.T) {
	errorPGConnection := &pgconn.PgError{Code: pgerrcode.ConnectionException}
	execError := errors.New("execError")
	execStore := errors.New("execStore")
	testCases := []struct {
		name        string
		wantError   error
		getExecutor func(t *testing.T) SQLExecutor
		close       bool
		gaugeValue  Gauge
		gaugeName   string
		getStorage  func(t *testing.T) IStorage
		syncMode    bool
	}{
		{
			name:      "storage_close",
			wantError: ErrorStorageDatabaseClosed,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				return executor
			},
			gaugeValue: Gauge(0),
			close:      true,
			gaugeName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().SetGauge(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return store
			},
			syncMode: true,
		},
		{
			name:      "success_form_store",
			wantError: nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return executor
			},
			gaugeValue: Gauge(12),
			gaugeName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().SetGauge(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return store
			},
			syncMode: true,
		},
		{
			name:      "exec_error",
			wantError: execError,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, execError).AnyTimes()
				return executor
			},
			gaugeValue: Gauge(12),
			gaugeName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().SetGauge(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return store
			},
			syncMode: true,
		},
		{
			name:      "pg_error",
			wantError: errorPGConnection,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).AnyTimes()
				return executor
			},
			gaugeValue: Gauge(12),
			gaugeName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().SetGauge(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return store
			},
			syncMode: true,
		},
		{
			name:      "pg_and_then_exec_error_error",
			wantError: execError,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				first := executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).Times(1)
				executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, execError).After(first)
				return executor
			},
			gaugeValue: Gauge(12),
			gaugeName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().SetGauge(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return store
			},
			syncMode: true,
		},
		{
			name:      "pg_and_then_exec_error_error",
			wantError: execStore,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				//first := executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).Times(1)
				//executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, execError).After(first)
				return executor
			},
			gaugeValue: Gauge(12),
			gaugeName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().SetGauge(gomock.Any(), gomock.Any()).Return(execStore).AnyTimes()
				return store
			},
			syncMode: true,
		},
		{
			name:      "success_not_sync_mode",
			wantError: nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				//first := executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).Times(1)
				//executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, execError).After(first)
				return executor
			},
			gaugeValue: Gauge(12),
			gaugeName:  "aboba",
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().SetGauge(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				return store
			},
			syncMode: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dbStorage := DBStorage{
				IStorage: tc.getStorage(t),
				storeCtx: context.Background(),
				db:       tc.getExecutor(t),
				close:    tc.close,
				syncMode: tc.syncMode,
			}
			err := dbStorage.SetGauge(tc.gaugeName, tc.gaugeValue)
			if tc.wantError != nil {
				assert.ErrorIs(t, tc.wantError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDBStorage_NewDBStorage(t *testing.T) {
	restoreError := errors.New("restoreError")
	errorBeginTX := errors.New("errorBeginTX")
	testCases := []struct {
		name        string
		wantError   error
		getExecutor func(t *testing.T) SQLExecutor
		restore     bool
		syncMode    bool
	}{
		{
			name:      "success_with_restore",
			wantError: nil,
			restore:   true,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				rowsG := NewMockIRows(ctrl)
				rowsG.EXPECT().Close().Return(nil).AnyTimes()
				rowsG.EXPECT().Err().Return(nil).AnyTimes()
				firstTimesG := rowsG.EXPECT().Next().Return(true).Times(2)
				rowsG.EXPECT().Next().Return(false).After(firstTimesG)
				firstG := rowsG.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(nil).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 2 {
						if ptr, ok := args[1].(*Gauge); ok {
							*ptr = Gauge(123)
						}
						if ptr, ok := args[0].(*string); ok {
							*ptr = "test"
						}
					}
				}).Times(1)
				rowsG.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(nil).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 2 {
						if ptr, ok := args[1].(*Gauge); ok {
							*ptr = Gauge(124)
						}
						if ptr, ok := args[0].(*string); ok {
							*ptr = "test1"
						}
					}
				}).After(firstG)
				fe := executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(rowsG, nil).Times(1)

				rows := NewMockIRows(ctrl)
				rows.EXPECT().Close().Return(nil).AnyTimes()
				rows.EXPECT().Err().Return(nil).AnyTimes()
				firstTimes := rows.EXPECT().Next().Return(true).Times(2)
				rows.EXPECT().Next().Return(false).After(firstTimes)
				first := rows.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(nil).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 2 {
						if ptr, ok := args[1].(*Counter); ok {
							*ptr = Counter(123)
						}
						if ptr, ok := args[0].(*string); ok {
							*ptr = "test"
						}
					}
				}).Times(1)
				rows.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(nil).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 2 {
						if ptr, ok := args[1].(*Counter); ok {
							*ptr = Counter(124)
						}
						if ptr, ok := args[0].(*string); ok {
							*ptr = "test1"
						}
					}
				}).After(first)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(rows, nil).After(fe)
				return executor
			},
		},
		{
			name:      "error_with_restore",
			wantError: restoreError,
			restore:   true,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, restoreError).AnyTimes()
				return executor
			},
		},
		{
			name:      "success_not_restore",
			wantError: nil,
			restore:   false,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().ExecContext(gomock.Any(), "TRUNCATE TABLE t_gauge").Return(NewMockIResult(ctrl), nil).AnyTimes()
				tx.EXPECT().ExecContext(gomock.Any(), "TRUNCATE TABLE t_counter").Return(NewMockIResult(ctrl), nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
		},
		{
			name:      "error_not_restore",
			wantError: errorBeginTX,
			restore:   false,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().ExecContext(gomock.Any(), "TRUNCATE TABLE t_gauge").Return(NewMockIResult(ctrl), nil).AnyTimes()
				tx.EXPECT().ExecContext(gomock.Any(), "TRUNCATE TABLE t_counter").Return(NewMockIResult(ctrl), nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, errorBeginTX).AnyTimes()
				return executor
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewDBStorage(context.TODO(), tc.getExecutor(t), tc.restore, tc.syncMode)
			if tc.wantError != nil {
				assert.ErrorIs(t, tc.wantError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
