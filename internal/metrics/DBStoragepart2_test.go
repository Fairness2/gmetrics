package metrics

import (
	"context"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"gmetrics/internal/contextkeys"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestDBStorage_FlushAndClose(t *testing.T) {
	errorBeginTX := errors.New("errorBeginTX")
	errorCommit := errors.New("error commit")
	errorPrepareContext := errors.New("PrepareContext")
	errorPrepareClose := errors.New("close")
	errorPrepareExec := errors.New("exec")
	errorRollback := errors.New("error rollback")
	errorRollbackOK := errors.New("sql: transaction has already been committed or rolled back")
	errorPGConnection := &pgconn.PgError{Code: pgerrcode.ConnectionException}
	errorGetGauges := errors.New("error get gauges")
	errorGetCounters := errors.New("error get counters")
	testCases := []struct {
		name        string
		err         error
		getExecutor func(t *testing.T) SQLExecutor
		close       bool
		getStorage  func(t *testing.T) IStorage
		wantClose   bool
	}{
		{
			name: "storage_close",
			err:  ErrorStorageDatabaseClosed,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
			close: true,
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"g1": 9}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"c1": 9}, nil).AnyTimes()
				return store
			},
			wantClose: true,
		},
		{
			name: "success_with_single",
			err:  nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"g1": 9}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"c1": 9}, nil).AnyTimes()
				return store
			},
			wantClose: true,
		},
		{
			name: "success_with_single_multiple",
			err:  nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"g1": 5, "g2": 7, "g3": 4}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"c1": 5, "c2": 7, "c3": 4}, nil).AnyTimes()
				return store
			},
			wantClose: true,
		},
		{
			name: "multiple_counters_and_clear_and_set",
			err:  nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"g1": 5, "g2": 7, "g3": 4}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"c1": 5, "c2": 7, "c3": 4}, nil).AnyTimes()
				return store
			},
			wantClose: true,
		},
		{
			name: "error_when_begin_tx",
			err:  errorBeginTX,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, errorBeginTX).AnyTimes()
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"g1": 5, "g2": 7, "g3": 4}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"c1": 5, "c2": 7, "c3": 4}, nil).AnyTimes()
				return store
			},
		},
		{
			name: "error_when_commit_tx",
			err:  errorCommit,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(errorCommit).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"g1": 5, "g2": 7, "g3": 4}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"c1": 5, "c2": 7, "c3": 4}, nil).AnyTimes()
				return store
			},
		},
		{
			name: "error_when_prepare_context",
			err:  errorPrepareContext,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, errorPrepareContext).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"g1": 5, "g2": 7, "g3": 4}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"c1": 5, "c2": 7, "c3": 4}, nil).AnyTimes()
				return store
			},
		},
		{
			name: "error_when_stmt_close",
			err:  nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(errorPrepareClose).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"g1": 5, "g2": 7, "g3": 4}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"c1": 5, "c2": 7, "c3": 4}, nil).AnyTimes()
				return store
			},
			wantClose: true,
		},
		{
			name: "error_when_stmt_exec",
			err:  errorPrepareExec,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errorPrepareExec).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"g1": 5, "g2": 7, "g3": 4}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"c1": 5, "c2": 7, "c3": 4}, nil).AnyTimes()
				return store
			},
		},
		{
			name: "error_when_tx_rollback",
			err:  nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(errorRollback).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"g1": 5, "g2": 7, "g3": 4}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"c1": 5, "c2": 7, "c3": 4}, nil).AnyTimes()
				return store
			},
			wantClose: true,
		},
		{
			name: "error_when_tx_rollback_ok",
			err:  nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(errorRollbackOK).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"g1": 5, "g2": 7, "g3": 4}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"c1": 5, "c2": 7, "c3": 4}, nil).AnyTimes()
				return store
			},
			wantClose: true,
		},
		{
			name: "error_pg_conn",
			err:  errorPGConnection,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, errorPGConnection).AnyTimes()
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"g1": 5, "g2": 7, "g3": 4}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"c1": 5, "c2": 7, "c3": 4}, nil).AnyTimes()
				return store
			},
		},
		{
			name: "empty_maps",
			err:  nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{}, nil).AnyTimes()
				return store
			},
			wantClose: true,
		},
		{
			name: "error_get_gauges",
			err:  errorGetGauges,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{}, errorGetGauges).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{}, nil).AnyTimes()
				return store
			},
		},
		{
			name: "error_get_counters",
			err:  errorGetCounters,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{}, errorGetCounters).AnyTimes()
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
			err := dbStorage.FlushAndClose()
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantClose, dbStorage.close)
		})
	}
}

func TestDBStorage_Sync(t *testing.T) {
	testCases := []struct {
		name               string
		asyncInterval      int64
		checkCancelContext bool
		wantError          error
		getExecutor        func(t *testing.T) SQLExecutor
		getStorage         func(t *testing.T) IStorage
		close              bool
	}{
		{
			name:               "success_on_sync",
			asyncInterval:      1,
			checkCancelContext: false,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"g1": 9}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"c1": 9}, nil).AnyTimes()
				return store
			},
		},
		{
			name:               "error_on_sync",
			asyncInterval:      1,
			checkCancelContext: false,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
			close: true,
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"g1": 9}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"c1": 9}, nil).AnyTimes()
				return store
			},
			wantError: ErrorStorageDatabaseClosed,
		},
		{
			name:               "sync_cancelled_early",
			asyncInterval:      2,
			checkCancelContext: true,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"g1": 9}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"c1": 9}, nil).AnyTimes()
				return store
			},
		},
		{
			name:               "sync_cancelled_early_with_error",
			asyncInterval:      2,
			checkCancelContext: true,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				prepared := NewMockIStmt(ctrl)
				prepared.EXPECT().Close().Return(nil).AnyTimes()
				prepared.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				tx := NewMockITX(ctrl)
				tx.EXPECT().Commit().Return(nil).AnyTimes()
				tx.EXPECT().Rollback().Return(nil).AnyTimes()
				tx.EXPECT().PrepareContext(gomock.Any(), gomock.Any()).Return(prepared, nil).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).Return(tx, nil).AnyTimes()
				return executor
			},
			close: true,
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"g1": 9}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"c1": 9}, nil).AnyTimes()
				return store
			},
			wantError: ErrorStorageDatabaseClosed,
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
			err := dbStorage.Sync(start)
			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError, "Want error %s, got %s", tc.wantError, err)

			} else {
				assert.NoError(t, err, "Want no error, got %s", err)
			}
		})
	}
}

func TestDBStorage_getGauges(t *testing.T) {
	errorQueryContext := errors.New("QueryContext")
	errorRowsErr := errors.New("errorRowsErr")
	errorScan := errors.New("errorScan")
	errorClose := errors.New("errorClose")
	testCases := []struct {
		name           string
		err            error
		getExecutor    func(t *testing.T) SQLExecutor
		close          bool
		expectedGauges map[string]Gauge
	}{
		{
			name: "storage_close",
			err:  ErrorStorageDatabaseClosed,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return executor
			},
			expectedGauges: map[string]Gauge{"test": Gauge(123)},
			close:          true,
		},
		{
			name: "success",
			err:  nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				rows := NewMockIRows(ctrl)
				rows.EXPECT().Close().Return(nil).AnyTimes()
				rows.EXPECT().Err().Return(nil).AnyTimes()
				firstTimes := rows.EXPECT().Next().Return(true).Times(2)
				rows.EXPECT().Next().Return(false).After(firstTimes)
				first := rows.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(nil).Do(func(args ...interface{}) {
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
				rows.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(nil).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 2 {
						if ptr, ok := args[1].(*Gauge); ok {
							*ptr = Gauge(124)
						}
						if ptr, ok := args[0].(*string); ok {
							*ptr = "test1"
						}
					}
				}).After(first)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(rows, nil).AnyTimes()
				return executor
			},
			expectedGauges: map[string]Gauge{"test": Gauge(123), "test1": Gauge(124)},
		},
		{
			name: "error_while_query",
			err:  errorQueryContext,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, errorQueryContext).AnyTimes()
				return executor
			},
			expectedGauges: make(map[string]Gauge),
		},
		{
			name: "error_rows_error",
			err:  errorRowsErr,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				rows := NewMockIRows(ctrl)
				rows.EXPECT().Close().Return(nil).AnyTimes()
				rows.EXPECT().Err().Return(errorRowsErr).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(rows, nil).AnyTimes()
				return executor
			},
			expectedGauges: make(map[string]Gauge),
		},
		{
			name: "first_next_false",
			err:  nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				rows := NewMockIRows(ctrl)
				rows.EXPECT().Close().Return(nil).AnyTimes()
				rows.EXPECT().Err().Return(nil).AnyTimes()
				rows.EXPECT().Next().Return(false).Times(1)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(rows, nil).AnyTimes()
				return executor
			},
			expectedGauges: make(map[string]Gauge),
		},
		{
			name: "first_scan_error",
			err:  nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				rows := NewMockIRows(ctrl)
				rows.EXPECT().Close().Return(nil).AnyTimes()
				rows.EXPECT().Err().Return(nil).AnyTimes()
				firstTimes := rows.EXPECT().Next().Return(true).Times(1)
				rows.EXPECT().Next().Return(false).After(firstTimes)
				rows.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(errorScan).Do(func(args ...interface{}) {
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
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(rows, nil).AnyTimes()
				return executor
			},
			expectedGauges: make(map[string]Gauge),
		},
		{
			name: "success_with_error_close",
			err:  nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				rows := NewMockIRows(ctrl)
				rows.EXPECT().Close().Return(errorClose).AnyTimes()
				rows.EXPECT().Err().Return(nil).AnyTimes()
				firstTimes := rows.EXPECT().Next().Return(true).Times(2)
				rows.EXPECT().Next().Return(false).After(firstTimes)
				first := rows.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(nil).Do(func(args ...interface{}) {
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
				rows.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(nil).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 2 {
						if ptr, ok := args[1].(*Gauge); ok {
							*ptr = Gauge(124)
						}
						if ptr, ok := args[0].(*string); ok {
							*ptr = "test1"
						}
					}
				}).After(first)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(rows, nil).AnyTimes()
				return executor
			},
			expectedGauges: map[string]Gauge{"test": Gauge(123), "test1": Gauge(124)},
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
			res, err := dbStorage.getGauges()
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedGauges, res)
			}
		})
	}
}

func TestDBStorage_getCounters(t *testing.T) {
	errorQueryContext := errors.New("QueryContext")
	errorRowsErr := errors.New("errorRowsErr")
	errorScan := errors.New("errorScan")
	errorClose := errors.New("errorClose")
	testCases := []struct {
		name             string
		err              error
		getExecutor      func(t *testing.T) SQLExecutor
		close            bool
		expectedCounters map[string]Counter
	}{
		{
			name: "storage_close",
			err:  ErrorStorageDatabaseClosed,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return executor
			},
			expectedCounters: map[string]Counter{"test": Counter(123)},
			close:            true,
		},
		{
			name: "success",
			err:  nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
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
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(rows, nil).AnyTimes()
				return executor
			},
			expectedCounters: map[string]Counter{"test": Counter(123), "test1": Counter(124)},
		},
		{
			name: "error_while_query",
			err:  errorQueryContext,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, errorQueryContext).AnyTimes()
				return executor
			},
			expectedCounters: make(map[string]Counter),
		},
		{
			name: "error_rows_error",
			err:  errorRowsErr,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				rows := NewMockIRows(ctrl)
				rows.EXPECT().Close().Return(nil).AnyTimes()
				rows.EXPECT().Err().Return(errorRowsErr).AnyTimes()
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(rows, nil).AnyTimes()
				return executor
			},
			expectedCounters: make(map[string]Counter),
		},
		{
			name: "first_next_false",
			err:  nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				rows := NewMockIRows(ctrl)
				rows.EXPECT().Close().Return(nil).AnyTimes()
				rows.EXPECT().Err().Return(nil).AnyTimes()
				rows.EXPECT().Next().Return(false).Times(1)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(rows, nil).AnyTimes()
				return executor
			},
			expectedCounters: make(map[string]Counter),
		},
		{
			name: "first_scan_error",
			err:  nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				rows := NewMockIRows(ctrl)
				rows.EXPECT().Close().Return(nil).AnyTimes()
				rows.EXPECT().Err().Return(nil).AnyTimes()
				firstTimes := rows.EXPECT().Next().Return(true).Times(1)
				rows.EXPECT().Next().Return(false).After(firstTimes)
				rows.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(errorScan).Do(func(args ...interface{}) {
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
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(rows, nil).AnyTimes()
				return executor
			},
			expectedCounters: make(map[string]Counter),
		},
		{
			name: "success_with_error_close",
			err:  nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				rows := NewMockIRows(ctrl)
				rows.EXPECT().Close().Return(errorClose).AnyTimes()
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
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(rows, nil).AnyTimes()
				return executor
			},
			expectedCounters: map[string]Counter{"test": Counter(123), "test1": Counter(124)},
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
			res, err := dbStorage.getCounters()
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCounters, res)
			}
		})
	}
}

func TestDBStorage_GetGauges(t *testing.T) {
	errorQueryContext := errors.New("QueryContext")
	errorNoItems := errors.New("no items")
	errorPGConnection := &pgconn.PgError{Code: pgerrcode.ConnectionException}
	testCases := []struct {
		name           string
		err            error
		getExecutor    func(t *testing.T) SQLExecutor
		close          bool
		expectedGauges map[string]Gauge
		getStorage     func(t *testing.T) IStorage
	}{
		{
			name: "storage_close",
			err:  ErrorStorageDatabaseClosed,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return executor
			},
			expectedGauges: map[string]Gauge{"test": Gauge(123)},
			close:          true,
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"test1": 9}, nil).AnyTimes()
				return store
			},
		},
		{
			name: "success",
			err:  nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				rows := NewMockIRows(ctrl)
				rows.EXPECT().Close().Return(nil).AnyTimes()
				rows.EXPECT().Err().Return(nil).AnyTimes()
				firstTimes := rows.EXPECT().Next().Return(true).Times(2)
				rows.EXPECT().Next().Return(false).After(firstTimes)
				first := rows.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(nil).Do(func(args ...interface{}) {
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
				rows.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(nil).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 2 {
						if ptr, ok := args[1].(*Gauge); ok {
							*ptr = Gauge(124)
						}
						if ptr, ok := args[0].(*string); ok {
							*ptr = "test1"
						}
					}
				}).After(first)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(rows, nil).AnyTimes()
				return executor
			},
			expectedGauges: map[string]Gauge{"test": Gauge(123), "test1": Gauge(124)},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"test1": 9}, nil).AnyTimes()
				return store
			},
		},
		{
			name: "error_while_query_with_pg_error_first_time",
			err:  errorQueryContext,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				first := executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).Times(1)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, errorQueryContext).After(first)
				return executor
			},
			expectedGauges: make(map[string]Gauge),
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"test1": 9}, nil).AnyTimes()
				return store
			},
		},
		{
			name: "error_while_query_with_pg_error",
			err:  errorPGConnection,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).AnyTimes()
				return executor
			},
			expectedGauges: make(map[string]Gauge),
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"test1": 9}, nil).AnyTimes()
				return store
			},
		},
		{
			name: "error_while_get_from_storage",
			err:  errorNoItems,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).AnyTimes()
				return executor
			},
			expectedGauges: map[string]Gauge{"test1": 9},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"test1": 9}, errorNoItems).AnyTimes()
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
			res, err := dbStorage.GetGauges()
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedGauges, res)
			}
		})
	}
}

func TestDBStorage_GetCounters(t *testing.T) {
	errorQueryContext := errors.New("QueryContext")
	errorNoItems := errors.New("no items")
	errorPGConnection := &pgconn.PgError{Code: pgerrcode.ConnectionException}
	testCases := []struct {
		name             string
		err              error
		getExecutor      func(t *testing.T) SQLExecutor
		close            bool
		expectedCounters map[string]Counter
		getStorage       func(t *testing.T) IStorage
	}{
		{
			name: "storage_close",
			err:  ErrorStorageDatabaseClosed,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return executor
			},
			expectedCounters: map[string]Counter{"test": Counter(123)},
			close:            true,
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetCounters().Return(map[string]Counter{"test1": 9}, nil).AnyTimes()
				return store
			},
		},
		{
			name: "success",
			err:  nil,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
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
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(rows, nil).AnyTimes()
				return executor
			},
			expectedCounters: map[string]Counter{"test": Counter(123), "test1": Counter(124)},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetCounters().Return(map[string]Counter{"test1": 9}, nil).AnyTimes()
				return store
			},
		},
		{
			name: "error_while_query_with_pg_error_first_time",
			err:  errorQueryContext,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				first := executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).Times(1)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, errorQueryContext).After(first)
				return executor
			},
			expectedCounters: make(map[string]Counter),
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetCounters().Return(map[string]Counter{"test1": 9}, nil).AnyTimes()
				return store
			},
		},
		{
			name: "error_while_query_with_pg_error",
			err:  errorPGConnection,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).AnyTimes()
				return executor
			},
			expectedCounters: make(map[string]Counter),
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetCounters().Return(map[string]Counter{"test1": 9}, nil).AnyTimes()
				return store
			},
		},
		{
			name: "error_while_get_fromStorage",
			err:  errorNoItems,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, errorPGConnection).AnyTimes()
				return executor
			},
			expectedCounters: map[string]Counter{"test1": 9},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetCounters().Return(map[string]Counter{"test1": 9}, errorNoItems).AnyTimes()
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
			res, err := dbStorage.GetCounters()
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCounters, res)
			}
		})
	}
}

func TestDBStorage_restore(t *testing.T) {
	errorSetGauges := errors.New("errorSetGauges")
	errorAddCounters := errors.New("errorAddCounters")
	errorNoItems := errors.New("no items")
	testCases := []struct {
		name        string
		err         error
		getExecutor func(t *testing.T) SQLExecutor
		close       bool
		getStorage  func(t *testing.T) IStorage
	}{
		{
			name: "storage_close",
			err:  ErrorStorageDatabaseClosed,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return executor
			},
			close: true,
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetCounters().Return(map[string]Counter{"test1": 9}, nil).AnyTimes()
				return store
			},
		},
		{
			name: "success",
			err:  nil,
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
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"test1": 9}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"test1": 9}, nil).AnyTimes()
				store.EXPECT().SetGauges(gomock.Any()).Return(nil).AnyTimes()
				store.EXPECT().AddCounters(gomock.Any()).Return(nil).AnyTimes()
				return store
			},
		},
		{
			name: "error_while_get_gauges_from_storage",
			err:  errorNoItems,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"test1": 9}, errorNoItems).AnyTimes()
				return store
			},
		},
		{
			name: "error_while_set_gauges",
			err:  errorSetGauges,
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
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(rowsG, nil).Times(1)
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"test1": 9}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"test1": 9}, nil).AnyTimes()
				store.EXPECT().SetGauges(gomock.Any()).Return(errorSetGauges).AnyTimes()
				store.EXPECT().AddCounters(gomock.Any()).Return(errorAddCounters).AnyTimes()
				return store
			},
		},
		{
			name: "error_while_get_counters",
			err:  errorNoItems,
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
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(rowsG, nil).Times(1)
				return executor
			},
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"test1": 9}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"test1": 9}, errorNoItems).AnyTimes()
				store.EXPECT().SetGauges(gomock.Any()).Return(nil).AnyTimes()
				store.EXPECT().AddCounters(gomock.Any()).Return(errorAddCounters).AnyTimes()
				return store
			},
		},
		{
			name: "error_while_add_counters",
			err:  errorAddCounters,
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
			getStorage: func(t *testing.T) IStorage {
				ctrl := gomock.NewController(t)
				store := NewMockIStorage(ctrl)
				store.EXPECT().GetGauges().Return(map[string]Gauge{"test1": 9}, nil).AnyTimes()
				store.EXPECT().GetCounters().Return(map[string]Counter{"test1": 9}, nil).AnyTimes()
				store.EXPECT().SetGauges(gomock.Any()).Return(nil).AnyTimes()
				store.EXPECT().AddCounters(gomock.Any()).Return(errorAddCounters).AnyTimes()
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
			err := dbStorage.restore()
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDBStorage_getCounter(t *testing.T) {
	errorScan := errors.New("errorScan")
	testCases := []struct {
		name            string
		err             error
		getExecutor     func(t *testing.T) SQLExecutor
		close           bool
		expectedCounter Counter
		counterName     string
	}{
		{
			name: "storage_close",
			err:  ErrorStorageDatabaseClosed,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return executor
			},
			expectedCounter: Counter(0),
			close:           true,
			counterName:     "aboba",
		},
		{
			name: "success",
			err:  nil,
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
				}).Times(1)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(row).AnyTimes()
				return executor
			},
			expectedCounter: Counter(12),
			counterName:     "aboba",
		},
		{
			name: "error_scan",
			err:  errorScan,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				row := NewMockIRows(ctrl)
				row.EXPECT().Scan(gomock.Any()).Return(errorScan).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Counter); ok {
							*ptr = Counter(12)
						}
					}
				}).Times(1)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(row).AnyTimes()
				return executor
			},
			expectedCounter: Counter(12),
			counterName:     "aboba",
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
			res, err := dbStorage.getCounter(tc.counterName)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCounter, res)
			}
		})
	}
}

func TestDBStorage_getGauge(t *testing.T) {
	errorScan := errors.New("errorScan")
	testCases := []struct {
		name          string
		err           error
		getExecutor   func(t *testing.T) SQLExecutor
		close         bool
		expectedGauge Gauge
		gaugeName     string
	}{
		{
			name: "storage_close",
			err:  ErrorStorageDatabaseClosed,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryContext(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return executor
			},
			expectedGauge: Gauge(0),
			close:         true,
			gaugeName:     "aboba",
		},
		{
			name: "success",
			err:  nil,
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
				}).Times(1)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(row).AnyTimes()
				return executor
			},
			expectedGauge: Gauge(12),
			gaugeName:     "aboba",
		},
		{
			name: "error_scan",
			err:  errorScan,
			getExecutor: func(t *testing.T) SQLExecutor {
				ctrl := gomock.NewController(t)
				row := NewMockIRows(ctrl)
				row.EXPECT().Scan(gomock.Any()).Return(errorScan).Do(func(args ...interface{}) {
					// Устанавливаем значения в переданные аргументы
					if len(args) >= 1 {
						if ptr, ok := args[0].(*Gauge); ok {
							*ptr = Gauge(12)
						}
					}
				}).Times(1)
				executor := NewMockSQLExecutor(ctrl)
				executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).Return(row).AnyTimes()
				return executor
			},
			expectedGauge: Gauge(12),
			gaugeName:     "aboba",
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
			res, err := dbStorage.getGauge(tc.gaugeName)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedGauge, res)
			}
		})
	}
}
