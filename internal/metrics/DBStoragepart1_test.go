package metrics

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDBStorage_syncGauges(t *testing.T) {
	errorBeginTX := errors.New("errorBeginTX")
	errorCommit := errors.New("error commit")
	errorPrepareContext := errors.New("PrepareContext")
	errorPrepareClose := errors.New("close")
	errorPrepareExec := errors.New("exec")
	errorRollback := errors.New("error rollback")
	errorRollbackOK := errors.New("sql: transaction has already been committed or rolled back")
	errorPGConnection := &pgconn.PgError{Code: pgerrcode.ConnectionException}
	testCases := []struct {
		name        string
		gauges      map[string]Gauge
		err         error
		getExecutor func(t *testing.T) SQLExecutor
	}{
		{
			name:   "single_gauge",
			gauges: map[string]Gauge{"g1": 9},
			err:    nil,
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
		},
		{
			name:   "multiple_gauges",
			gauges: map[string]Gauge{"g1": 5, "g2": 7, "g3": 4},
			err:    nil,
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
		},
		{
			name:   "error_when_begin_tx",
			gauges: map[string]Gauge{"g1": 3},
			err:    errorBeginTX,
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
		},
		{
			name:   "error_when_commit_tx",
			gauges: map[string]Gauge{"g1": 3},
			err:    errorCommit,
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
		},
		{
			name:   "error_when_prepare_context",
			gauges: map[string]Gauge{"g1": 3},
			err:    errorPrepareContext,
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
		},
		{
			name:   "error_when_stmt_close",
			gauges: map[string]Gauge{"g1": 3},
			err:    nil,
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
		},
		{
			name:   "error_when_stmt_exec",
			gauges: map[string]Gauge{"g1": 3},
			err:    errorPrepareExec,
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
		},
		{
			name:   "error_when_tx_rollback",
			gauges: map[string]Gauge{"g1": 3},
			err:    nil,
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
		},
		{
			name:   "error_when_tx_rollback_ok",
			gauges: map[string]Gauge{"g1": 3},
			err:    nil,
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
		},
		{
			name:   "error_pg_conn",
			gauges: map[string]Gauge{"g1": 3},
			err:    errorPGConnection,
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
		},
		{
			name:   "empty_gauges",
			gauges: make(map[string]Gauge),
			err:    nil,
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
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dbStorage := DBStorage{
				IStorage: NewMemStorage(),
				storeCtx: context.Background(),
				db:       tc.getExecutor(t),
			}
			err := dbStorage.syncGauges(tc.gauges)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDBStorage_SetGauges(t *testing.T) {
	errorBeginTX := errors.New("errorBeginTX")
	errorCommit := errors.New("error commit")
	errorPrepareContext := errors.New("PrepareContext")
	errorPrepareClose := errors.New("close")
	errorPrepareExec := errors.New("exec")
	errorRollback := errors.New("error rollback")
	errorRollbackOK := errors.New("sql: transaction has already been committed or rolled back")
	errorPGConnection := &pgconn.PgError{Code: pgerrcode.ConnectionException}
	testCases := []struct {
		name        string
		gauges      map[string]Gauge
		err         error
		getExecutor func(t *testing.T) SQLExecutor
		syncMode    bool
	}{
		{
			name:   "single_gauge_without_sync",
			gauges: map[string]Gauge{"g1": 9},
			err:    nil,
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
		},
		{
			name:   "multiple_gauges_without_sync",
			gauges: map[string]Gauge{"g1": 5, "g2": 7, "g3": 4},
			err:    nil,
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
		},
		{
			name:   "single_gauge",
			gauges: map[string]Gauge{"g1": 9},
			err:    nil,
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
			syncMode: true,
		},
		{
			name:   "multiple_gauges",
			gauges: map[string]Gauge{"g1": 5, "g2": 7, "g3": 4},
			err:    nil,
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
			syncMode: true,
		},
		{
			name:   "error_when_begin_tx",
			gauges: map[string]Gauge{"g1": 3},
			err:    errorBeginTX,
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
			syncMode: true,
		},
		{
			name:   "error_when_commit_tx",
			gauges: map[string]Gauge{"g1": 3},
			err:    errorCommit,
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
			syncMode: true,
		},
		{
			name:   "error_when_prepare_context",
			gauges: map[string]Gauge{"g1": 3},
			err:    errorPrepareContext,
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
			syncMode: true,
		},
		{
			name:   "error_when_stmt_close",
			gauges: map[string]Gauge{"g1": 3},
			err:    nil,
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
			syncMode: true,
		},
		{
			name:   "error_when_stmt_exec",
			gauges: map[string]Gauge{"g1": 3},
			err:    errorPrepareExec,
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
			syncMode: true,
		},
		{
			name:   "error_when_tx_rollback",
			gauges: map[string]Gauge{"g1": 3},
			err:    nil,
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
			syncMode: true,
		},
		{
			name:   "error_when_tx_rollback_ok",
			gauges: map[string]Gauge{"g1": 3},
			err:    nil,
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
			syncMode: true,
		},
		{
			name:   "error_pg_conn",
			gauges: map[string]Gauge{"g1": 3},
			err:    errorPGConnection,
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
			syncMode: true,
		},
		{
			name:   "empty_gauges",
			gauges: make(map[string]Gauge),
			err:    nil,
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
			syncMode: true,
		},
		{
			name:   "empty_gauges_without_sync",
			gauges: make(map[string]Gauge),
			err:    nil,
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
			syncMode: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dbStorage := DBStorage{
				IStorage: NewMemStorage(),
				storeCtx: context.Background(),
				db:       tc.getExecutor(t),
				syncMode: tc.syncMode,
			}
			err := dbStorage.SetGauges(tc.gauges)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDBStorage_AddCounters(t *testing.T) {
	errorBeginTX := errors.New("errorBeginTX")
	errorCommit := errors.New("error commit")
	errorPrepareContext := errors.New("PrepareContext")
	errorPrepareClose := errors.New("close")
	errorPrepareExec := errors.New("exec")
	errorRollback := errors.New("error rollback")
	errorRollbackOK := errors.New("sql: transaction has already been committed or rolled back")
	errorPGConnection := &pgconn.PgError{Code: pgerrcode.ConnectionException}
	testCases := []struct {
		name        string
		counters    map[string]Counter
		err         error
		getExecutor func(t *testing.T) SQLExecutor
		syncMode    bool
	}{
		{
			name:     "single_counter_without_sync",
			counters: map[string]Counter{"c1": 9},
			err:      nil,
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
			syncMode: false,
		},
		{
			name:     "multiple_counters_without_sync",
			counters: map[string]Counter{"c1": 5, "c2": 7, "c3": 4},
			err:      nil,
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
			syncMode: false,
		},
		{
			name:     "single_counter",
			counters: map[string]Counter{"c1": 9},
			err:      nil,
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
			syncMode: true,
		},
		{
			name:     "multiple_counters",
			counters: map[string]Counter{"c1": 5, "c2": 7, "c3": 4},
			err:      nil,
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
			syncMode: true,
		},
		{
			name:     "error_when_begin_tx",
			counters: map[string]Counter{"g1": 3},
			err:      errorBeginTX,
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
			syncMode: true,
		},
		{
			name:     "error_when_commit_tx",
			counters: map[string]Counter{"g1": 3},
			err:      errorCommit,
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
			syncMode: true,
		},
		{
			name:     "error_when_prepare_context",
			counters: map[string]Counter{"g1": 3},
			err:      errorPrepareContext,
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
			syncMode: true,
		},
		{
			name:     "error_when_stmt_close",
			counters: map[string]Counter{"g1": 3},
			err:      nil,
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
			syncMode: true,
		},
		{
			name:     "error_when_stmt_exec",
			counters: map[string]Counter{"g1": 3},
			err:      errorPrepareExec,
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
			syncMode: true,
		},
		{
			name:     "error_when_tx_rollback",
			counters: map[string]Counter{"g1": 3},
			err:      nil,
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
			syncMode: true,
		},
		{
			name:     "error_when_tx_rollback_ok",
			counters: map[string]Counter{"g1": 3},
			err:      nil,
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
			syncMode: true,
		},
		{
			name:     "error_pg_conn",
			counters: map[string]Counter{"g1": 3},
			err:      errorPGConnection,
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
			syncMode: true,
		},
		{
			name:     "empty_counters",
			counters: make(map[string]Counter),
			err:      nil,
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
			syncMode: true,
		},
		{
			name:     "empty_counters_without_sync",
			counters: make(map[string]Counter),
			err:      nil,
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
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dbStorage := DBStorage{
				IStorage: NewMemStorage(),
				storeCtx: context.Background(),
				db:       tc.getExecutor(t),
				syncMode: tc.syncMode,
			}
			err := dbStorage.AddCounters(tc.counters)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDBStorage_Flush(t *testing.T) {
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
			err := dbStorage.Flush()
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
