package metrics

import (
	"context"
	"database/sql"
)

// ITX Интерфейс для скрытия реализации sql.TX
type ITX interface {
	Rollback() error
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Commit() error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// IRows Интерфейс для скрытия реализации sql.Rows
type IRows interface {
	Close() error
	Err() error
	Next() bool
	Scan(dest ...any) error
}

// IRow Интерфейс для скрытия реализации sql.Row
type IRow interface {
	Scan(dest ...any) error
}

// IResult Интерфейс для скрытия реализации sql.Result
type IResult interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

// DBAdapter Структура которая скрывает реализацию подключения к бд за функциями, которые можно замокировать
type DBAdapter struct {
	real *sql.DB
}

// NewDBAdapter Создание адаптера для sql.DB
func NewDBAdapter(real *sql.DB) *DBAdapter {
	return &DBAdapter{real}
}

// BeginTx Адаптер к sql.DB.BeginTx
func (dba DBAdapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (ITX, error) {
	return dba.real.BeginTx(ctx, opts)
}

// QueryContext Адаптер к sql.DB.QueryContext
func (dba DBAdapter) QueryContext(ctx context.Context, query string, args ...any) (IRows, error) {
	return dba.real.QueryContext(ctx, query, args...)
}

// QueryRowContext Адаптер к sql.DB.QueryRowContext
func (dba DBAdapter) QueryRowContext(ctx context.Context, query string, args ...any) IRow {
	return dba.real.QueryRowContext(ctx, query, args...)
}

// ExecContext Адаптер к sql.DB.ExecContext
func (dba DBAdapter) ExecContext(ctx context.Context, query string, args ...any) (IResult, error) {
	return dba.real.ExecContext(ctx, query, args...)
}
