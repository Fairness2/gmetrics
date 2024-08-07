package metrics

import (
	"context"
	"database/sql"
	"errors"
	"gmetrics/internal/logger"
	"time"
)

// DBStorage хранилище метрик в базе данных
// TODO Возможно сделать хранилище с промежуточным хранение в памяти
type DBStorage struct {
	// storeCtx контекст, который отвечает за запросы
	storeCtx context.Context
	// db пул соединений с базой данных, которыми может пользоваться хранилище
	db *sql.DB
}

// NewDBStorage создание нового хранилища в базе данных
func NewDBStorage(ctx context.Context, db *sql.DB) (*DBStorage, error) {
	return &DBStorage{
		storeCtx: ctx,
		db:       db,
	}, nil
}

func (storage *DBStorage) SetGauge(name string, value Gauge) error {
	_, err := storage.db.ExecContext(storage.storeCtx, "INSERT INTO t_gauge (name, value) VALUES ($1, $2) on conflict (name) do update set value = $2, updated_at = $3", name, value, time.Now())
	return err
}

func (storage *DBStorage) AddCounter(name string, value Counter) error {
	_, err := storage.db.ExecContext(storage.storeCtx, "INSERT INTO t_counter (name, value) VALUES ($1, $2) on conflict (name) do update set value = t_counter.value + $2, updated_at = $3", name, value, time.Now())
	return err
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
func (storage *DBStorage) GetGauge(name string) (Gauge, bool) {
	var value Gauge
	row := storage.db.QueryRowContext(storage.storeCtx, "SELECT value FROM t_gauge WHERE name = $1", name)
	if err := row.Scan(&value); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Log.Error(err)
		}
		return value, false
	}

	return value, true
}

func (storage *DBStorage) GetCounter(name string) (Counter, bool) {
	var value Counter
	row := storage.db.QueryRowContext(storage.storeCtx, "SELECT value FROM t_counter WHERE name = $1", name)
	if err := row.Scan(&value); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.Log.Error(err)
		}
		return value, false
	}

	return value, true
}

func (storage *DBStorage) GetGauges() (map[string]Gauge, error) {
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

func (storage *DBStorage) GetCounters() (map[string]Counter, error) {
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

// SetGauges массовое обновление гауге в базе
func (storage *DBStorage) SetGauges(gauges map[string]Gauge) error {
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

// AddCounters массовое обновление каунтер в базе
func (storage *DBStorage) AddCounters(counters map[string]Counter) error {
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
