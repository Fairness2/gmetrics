package metrics

import (
	"context"
	"database/sql"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/golang/mock/gomock"
)

func BenchmarkDBStorage_SetMetrics(b *testing.B) {
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
	ctrl := gomock.NewController(b)
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()
			executor := NewMockSQLExecutor(ctrl)
			executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&MockSQLResult{}, nil)
			executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&sql.Row{})
			executor.EXPECT().QueryContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&sql.Rows{}, nil)
			executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).AnyTimes().Return(&sql.Tx{}, nil)
			storage := NewMemStorage()
			store := &DBStorage{
				IStorage: storage,
				storeCtx: ctx,
				db:       executor,
				syncMode: false,
				close:    false,
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := store.SetGauges(bm.gauges); err != nil {
					b.Fatal(err)
				}
				if err := store.AddCounters(bm.counters); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

type MockSQLResult struct{}

func (m MockSQLResult) LastInsertId() (int64, error) {
	return 1, nil
}

func (m MockSQLResult) RowsAffected() (int64, error) {
	return 1, nil
}

func TestDBStorage_IsSyncMode(t *testing.T) {
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
			memStore := DBStorage{
				IStorage: NewMemStorage(),
				db:       NewMockSQLExecutor(gomock.NewController(t)),
				syncMode: tc.syncMode,
				close:    false,
			}
			assert.Equal(t, tc.expectResult, memStore.IsSyncMode())
		})
	}
}

func TestDBStorage_Close(t *testing.T) {
	testCases := []struct {
		name             string
		withErrorOnClose bool
		wantError        error
		close            bool
	}{
		{
			name:             "success on close",
			withErrorOnClose: false,
			wantError:        nil,
			close:            false,
		},
		{
			name:             "error on close",
			withErrorOnClose: true,
			wantError:        ErrorStorageDatabaseClosed,
			close:            true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			durationFileStorage := &DBStorage{
				IStorage: NewMemStorage(),
				db:       NewMockSQLExecutor(gomock.NewController(t)),
				syncMode: false,
				close:    tc.close,
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

func TestDBStorage_clean(t *testing.T) {
	errorBeginTX := errors.New("error begin tx")
	errorGauge := errors.New("error begin truncate gauge")
	errorCounter := errors.New("error begin truncate counter")
	errorCommit := errors.New("error commit")
	errorRollback := errors.New("error rollback")
	errorRollbackOK := errors.New("sql: transaction has already been committed or rolled back")
	testCases := []struct {
		name        string
		close       bool
		expectErr   error
		getExecutor func(t *testing.T) SQLExecutor
	}{
		{
			name:      "success_clean",
			close:     false,
			expectErr: nil,
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
			name:      "error_with_tx_begin",
			close:     false,
			expectErr: errorBeginTX,
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
		{
			name:      "storage_closed",
			close:     true,
			expectErr: ErrorStorageDatabaseClosed,
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
			name:      "error_with_gauge_truncate",
			close:     false,
			expectErr: errorGauge,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().ExecContext(gomock.Any(), "TRUNCATE TABLE t_gauge").Return(NewMockIResult(ctrl), errorGauge).AnyTimes()
				tx.EXPECT().ExecContext(gomock.Any(), "TRUNCATE TABLE t_counter").Return(NewMockIResult(ctrl), nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
		},
		{
			name:      "error_with_gauge_counter",
			close:     false,
			expectErr: errorCounter,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().ExecContext(gomock.Any(), "TRUNCATE TABLE t_gauge").Return(NewMockIResult(ctrl), nil).AnyTimes()
				tx.EXPECT().ExecContext(gomock.Any(), "TRUNCATE TABLE t_counter").Return(NewMockIResult(ctrl), errorCounter).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
		},
		{
			name:      "error_with_commit",
			close:     false,
			expectErr: errorCommit,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(errorCommit).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().ExecContext(gomock.Any(), "TRUNCATE TABLE t_gauge").Return(NewMockIResult(ctrl), nil).AnyTimes()
				tx.EXPECT().ExecContext(gomock.Any(), "TRUNCATE TABLE t_counter").Return(NewMockIResult(ctrl), nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
		},
		{
			name:      "error_with_rollback",
			close:     false,
			expectErr: nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(errorRollback).AnyTimes()
				tx.EXPECT().ExecContext(gomock.Any(), "TRUNCATE TABLE t_gauge").Return(NewMockIResult(ctrl), nil).AnyTimes()
				tx.EXPECT().ExecContext(gomock.Any(), "TRUNCATE TABLE t_counter").Return(NewMockIResult(ctrl), nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
		},
		{
			name:      "error_with_rollback_ok",
			close:     false,
			expectErr: nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(errorRollbackOK).AnyTimes()
				tx.EXPECT().ExecContext(gomock.Any(), "TRUNCATE TABLE t_gauge").Return(NewMockIResult(ctrl), nil).AnyTimes()
				tx.EXPECT().ExecContext(gomock.Any(), "TRUNCATE TABLE t_counter").Return(NewMockIResult(ctrl), nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dbStorage := DBStorage{
				IStorage: NewMemStorage(),
				storeCtx: context.Background(),
				db:       tc.getExecutor(t),
				close:    tc.close,
			}

			err := dbStorage.clean()
			if tc.expectErr != nil {
				assert.ErrorIs(t, err, tc.expectErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
