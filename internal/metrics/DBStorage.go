package metrics

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"gmetrics/internal/logger"
	"time"
)

type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// DBStorage хранилище метрик в базе данных
// TODO Возможно сделать хранилище с промежуточным хранение в памяти
type DBStorage struct {
	// storeCtx контекст, который отвечает за запросы
	storeCtx context.Context
	// db пул соединений с базой данных, которыми может пользоваться хранилище
	db SQLExecutor
}

// NewDBStorage создание нового хранилища в базе данных
func NewDBStorage(ctx context.Context, db *sql.DB) (*DBStorage, error) {
	return &DBStorage{
		storeCtx: ctx,
		db:       db,
	}, nil
}

func (storage *DBStorage) SetGauge(name string, value Gauge) (err error) {
	pause := time.Second
	var pgErr *pgconn.PgError
	for i := 0; i < 3; i++ {
		err = storage.setGauge(name, value)
		if err == nil {
			break
		}
		logger.Log.Error(err)
		if !errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			break
		}

		<-time.After(pause)
		pause += 2 * time.Second
	}
	return err
}

func (storage *DBStorage) setGauge(name string, value Gauge) error {
	_, err := storage.db.ExecContext(storage.storeCtx, "INSERT INTO t_gauge (name, value) VALUES ($1, $2) on conflict (name) do update set value = $2, updated_at = $3", name, value, time.Now())
	return err
}

func (storage *DBStorage) AddCounter(name string, value Counter) (err error) {
	pause := time.Second
	var pgErr *pgconn.PgError
	for i := 0; i < 3; i++ {
		err = storage.addCounter(name, value)
		if err == nil {
			break
		}
		logger.Log.Error(err)
		if !errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			break
		}

		<-time.After(pause)
		pause += 2 * time.Second
	}
	return err
}

func (storage *DBStorage) addCounter(name string, value Counter) error {
	_, err := storage.db.ExecContext(storage.storeCtx, "INSERT INTO t_counter (name, value) VALUES ($1, $2) on conflict (name) do update set value = t_counter.value + $2, updated_at = $3", name, value, time.Now())
	return err
}

func (storage *DBStorage) GetGauge(name string) (g Gauge, ok bool) {
	pause := time.Second
	var err error
	var pgErr *pgconn.PgError
	for i := 0; i < 3; i++ {
		g, err = storage.getGauge(name)
		if err == nil {
			ok = true
			break
		}
		if errors.Is(err, sql.ErrNoRows) {
			ok = false
			break
		}
		logger.Log.Error(err)
		if !errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			ok = false
			break
		}

		<-time.After(pause)
		pause += 2 * time.Second
	}
	return g, ok
}

// Get Получение значения метрики из хранилища
// Parameters:
//
//	name: имя метрики
//
// Returns:
//
//	value: значение метрики
//	ok: флаг, указывающий на наличие метрики в хранилище
func (storage *DBStorage) getGauge(name string) (Gauge, error) {
	var value Gauge
	row := storage.db.QueryRowContext(storage.storeCtx, "SELECT value FROM t_gauge WHERE name = $1", name)
	if err := row.Scan(&value); err != nil {
		return value, err
	}

	return value, nil
}

func (storage *DBStorage) GetCounter(name string) (c Counter, ok bool) {
	pause := time.Second
	var err error
	var pgErr *pgconn.PgError
	for i := 0; i < 3; i++ {
		c, err = storage.getCounter(name)
		if err == nil {
			ok = true
			break
		}
		if errors.Is(err, sql.ErrNoRows) {
			ok = false
			break
		}
		logger.Log.Error(err)
		if !(errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)) {
			ok = false
			break
		}

		<-time.After(pause)
		pause += 2 * time.Second
	}
	return c, ok
}

func (storage *DBStorage) getCounter(name string) (Counter, error) {
	var value Counter
	row := storage.db.QueryRowContext(storage.storeCtx, "SELECT value FROM t_counter WHERE name = $1", name)
	if err := row.Scan(&value); err != nil {
		return value, err
	}

	return value, nil
}

func (storage *DBStorage) GetGauges() (gauges map[string]Gauge, err error) {
	pause := time.Second
	var pgErr *pgconn.PgError
	for i := 0; i < 3; i++ {
		gauges, err = storage.getGauges()
		if err == nil {
			break
		}
		logger.Log.Error(err)
		if !(errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)) {
			break
		}

		<-time.After(pause)
		pause += 2 * time.Second
	}
	return gauges, err
}

func (storage *DBStorage) getGauges() (map[string]Gauge, error) {
	gauges := make(map[string]Gauge)
	rows, err := storage.db.QueryContext(storage.storeCtx, "SELECT name, value FROM t_gauge")
	if err != nil {
		return gauges, err
	}
	// Закроем строки, чтобы освободить соединение
	defer func() {
		if rErr := rows.Close(); rErr == nil {
			logger.Log.Error(rErr)
		}
	}()
	if err = rows.Err(); err != nil {
		return gauges, err
	}
	var (
		name  string
		value Gauge
	)
	for rows.Next() {
		err = rows.Scan(&name, &value)
		if err != nil {
			logger.Log.Error(err)
			continue
		}
		gauges[name] = value
	}

	return gauges, nil
}

func (storage *DBStorage) GetCounters() (counters map[string]Counter, err error) {
	pause := time.Second
	var pgErr *pgconn.PgError
	for i := 0; i < 3; i++ {
		counters, err = storage.getCounters()
		if err == nil {
			break
		}
		logger.Log.Error(err)
		if !(errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)) {
			break
		}

		<-time.After(pause)
		pause += 2 * time.Second
	}
	return counters, err
}

func (storage *DBStorage) getCounters() (map[string]Counter, error) {
	counters := make(map[string]Counter)
	rows, err := storage.db.QueryContext(storage.storeCtx, "SELECT name, value FROM t_counter")
	if err != nil {
		return counters, err
	}
	// Закроем строки, чтобы освободить соединение
	defer func() {
		if rErr := rows.Close(); rErr == nil {
			logger.Log.Error(rErr)
		}
	}()
	if err = rows.Err(); err != nil {
		return counters, err
	}
	var (
		name  string
		value Counter
	)
	for rows.Next() {
		err = rows.Scan(&name, &value)
		if err != nil {
			logger.Log.Error(err)
			continue
		}
		counters[name] = value
	}

	return counters, nil
}

func (storage *DBStorage) SetGauges(gauges map[string]Gauge) (err error) {
	pause := time.Second
	var pgErr *pgconn.PgError
	for i := 0; i < 3; i++ {
		err = storage.setGauges(gauges)
		if err == nil {
			break
		}
		logger.Log.Error(err)
		if !(errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)) {
			break
		}

		<-time.After(pause)
		pause += 2 * time.Second
	}
	return err
}

// SetGauges массовое обновление гауге в базе
func (storage *DBStorage) setGauges(gauges map[string]Gauge) error {
	nowTime := time.Now()
	tx, err := storage.db.BeginTx(storage.storeCtx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if tErr := tx.Rollback(); tErr != nil && tErr.Error() != "sql: transaction has already been committed or rolled back" {
			logger.Log.Error(tErr)
		}
	}()
	prepared, err := tx.PrepareContext(storage.storeCtx, "INSERT INTO t_gauge (name, value) VALUES ($1, $2) on conflict (name) do update set value = $2, updated_at = $3")
	if err != nil {
		return err
	}
	defer func() {
		if pErr := prepared.Close(); pErr != nil {
			logger.Log.Error(pErr)
		}
	}()

	for name, gauge := range gauges {
		if _, err = prepared.Exec(name, gauge, nowTime); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (storage *DBStorage) AddCounters(counters map[string]Counter) (err error) {
	pause := time.Second
	var pgErr *pgconn.PgError
	for i := 0; i < 3; i++ {
		err = storage.addCounters(counters)
		if err == nil {
			break
		}
		logger.Log.Error(err)
		if !(errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code)) {
			break
		}

		<-time.After(pause)
		pause += 2 * time.Second
	}
	return err
}

// AddCounters массовое обновление каунтер в базе
func (storage *DBStorage) addCounters(counters map[string]Counter) error {
	nowTime := time.Now()
	tx, err := storage.db.BeginTx(storage.storeCtx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if tErr := tx.Rollback(); tErr != nil && tErr.Error() != "sql: transaction has already been committed or rolled back" {
			logger.Log.Error(tErr)
		}
	}()
	prepared, err := tx.PrepareContext(storage.storeCtx, "INSERT INTO t_counter (name, value) VALUES ($1, $2) on conflict (name) do update set value = t_counter.value + $2, updated_at = $3")
	if err != nil {
		return err
	}
	defer func() {
		if pErr := prepared.Close(); pErr != nil {
			logger.Log.Error(pErr)
		}
	}()

	for name, counter := range counters {
		if _, err = prepared.Exec(name, counter, nowTime); err != nil {
			return err
		}
	}
	return tx.Commit()
}
