package metrics

import (
	"context"
	"database/sql"
	"errors"
	"gmetrics/internal/logger"
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
	_, err := storage.db.ExecContext(storage.storeCtx, "INSERT INTO t_gauge (name, value) VALUES ($1, $2) on conflict (name) do update set value = $2", name, value)
	return err
}

func (storage *DBStorage) AddCounter(name string, value Counter) error {
	_, err := storage.db.ExecContext(storage.storeCtx, "INSERT INTO t_counter (name, value) VALUES ($1, $2) on conflict (name) do update set value = t_counter.value + $2", name, value)
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
