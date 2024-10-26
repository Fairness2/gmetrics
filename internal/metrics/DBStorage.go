package metrics

import (
	"context"
	"database/sql"
	"errors"
	"gmetrics/internal/contextkeys"
	"gmetrics/internal/logger"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrorStorageDatabaseClosed ошибка, указывающая, что хранилище базы данных уже закрыто.
var ErrorStorageDatabaseClosed = errors.New("DB storage is already closed")

// SQLExecutor интерфейс с нужными функциями из sql.DB
type SQLExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// DBStorage хранилище метрик в базе данных
// TODO Подумать об очищении значений из памяти по наступлению какого-то события
type DBStorage struct {
	IStorage
	// storeCtx контекст, который отвечает за запросы
	storeCtx context.Context
	// db пул соединений с базой данных, которыми может пользоваться хранилище
	db SQLExecutor
	// syncMode сохранять ли сразу данные в базу
	syncMode bool
	// close закрыто ли хранилище
	close bool
}

// NewDBStorage создание нового хранилища в базе данных
func NewDBStorage(ctx context.Context, db SQLExecutor, restore bool, syncMode bool) (*DBStorage, error) {
	storage := NewMemStorage()
	dbStorage := &DBStorage{
		IStorage: storage,
		storeCtx: ctx,
		db:       db,
		syncMode: syncMode,
		close:    false,
	}
	if restore {
		// Восстанавливаем хранилище из файла, возвращаем ошибку, если чтение вернуло ошибку не с типом несуществующего файла или пустого файла
		if err := dbStorage.restore(); err != nil {
			logger.Log.Infow("Restore store failed", "error", err)
			return dbStorage, err
		}
	} else {
		if err := dbStorage.clean(); err != nil {
			logger.Log.Infow("Clean store failed", "error", err)
			return dbStorage, err
		}
	}
	return dbStorage, nil
}

// SetGauge устанавливаем gauge
func (storage *DBStorage) SetGauge(name string, value Gauge) error {
	err := storage.IStorage.SetGauge(name, value)
	if err != nil {
		return err
	}
	if storage.syncMode {
		if err = storage.syncGauge(name, value); err != nil {
			return err
		}
	}
	return nil
}

// syncGauge синхронизируем gauge в базу с повторными попытками сохранения
func (storage *DBStorage) syncGauge(name string, value Gauge) (err error) {
	if storage.close {
		return ErrorStorageDatabaseClosed
	}
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

// setGauge записываем Gauge в бд
func (storage *DBStorage) setGauge(name string, value Gauge) error {
	_, err := storage.db.ExecContext(storage.storeCtx, "INSERT INTO t_gauge (name, value) VALUES ($1, $2) on conflict (name) do update set value = $2, updated_at = $3", name, value, time.Now())
	return err
}

// AddCounter добавляем каунтер
func (storage *DBStorage) AddCounter(name string, value Counter) error {
	err := storage.IStorage.AddCounter(name, value)
	if err != nil {
		return err
	}
	if storage.syncMode {
		if err = storage.syncCounter(name, value); err != nil {
			return err
		}
	}
	return nil
}

// syncCounter синхронизируем Counter в базу с повторными попытками сохранения
func (storage *DBStorage) syncCounter(name string, value Counter) (err error) {
	if storage.close {
		return ErrorStorageDatabaseClosed
	}
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

// addCounter сохраняем Counter в бд
func (storage *DBStorage) addCounter(name string, value Counter) error {
	_, err := storage.db.ExecContext(storage.storeCtx, "INSERT INTO t_counter (name, value) VALUES ($1, $2) on conflict (name) do update set value = t_counter.value + $2, updated_at = $3", name, value, time.Now())
	return err
}

// GetGauge получение отдельного gauge
func (storage *DBStorage) GetGauge(name string) (g Gauge, ok bool) {
	// Ищем в памяти значение
	g, ok = storage.IStorage.GetGauge(name)
	if ok {
		return g, ok
	}
	// Если нет в памяти данных, то идём в базу данных
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

// getGauge Получение значения Gauge из бд
func (storage *DBStorage) getGauge(name string) (Gauge, error) {
	var value Gauge
	if storage.close {
		return value, ErrorStorageDatabaseClosed
	}
	row := storage.db.QueryRowContext(storage.storeCtx, "SELECT value FROM t_gauge WHERE name = $1", name)
	if err := row.Scan(&value); err != nil {
		return value, err
	}

	return value, nil
}

// GetCounter получение отдельного counter
func (storage *DBStorage) GetCounter(name string) (c Counter, ok bool) {
	// Ищем в памяти значение
	c, ok = storage.IStorage.GetCounter(name)
	if ok {
		return c, ok
	}
	// Если нет в памяти данных, то идём в базу данных
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

// getCounter получаем Counter из бд
func (storage *DBStorage) getCounter(name string) (Counter, error) {
	var value Counter
	if storage.close {
		return value, ErrorStorageDatabaseClosed
	}
	row := storage.db.QueryRowContext(storage.storeCtx, "SELECT value FROM t_counter WHERE name = $1", name)
	if err := row.Scan(&value); err != nil {
		return value, err
	}

	return value, nil
}

// GetGauges получение всех gauge
func (storage *DBStorage) GetGauges() (gauges map[string]Gauge, err error) {
	// Ищем в памяти значение
	gauges, err = storage.IStorage.GetGauges()
	if err != nil {
		return gauges, err
	}
	// Если нет в памяти данных, то идём в базу данных
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
	if err != nil {
		return gauges, err
	}
	return gauges, err
}

// getGauges получение всех Gauge из БД
func (storage *DBStorage) getGauges() (map[string]Gauge, error) {
	gauges := make(map[string]Gauge)
	if storage.close {
		return gauges, ErrorStorageDatabaseClosed
	}
	rows, err := storage.db.QueryContext(storage.storeCtx, "SELECT name, value FROM t_gauge")
	if err != nil {
		return gauges, err
	}
	// Закроем строки, чтобы освободить соединение
	defer func() {
		if rErr := rows.Close(); rErr != nil {
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

// GetCounters получение всех counter
func (storage *DBStorage) GetCounters() (counters map[string]Counter, err error) {
	// Ищем в памяти значение
	counters, err = storage.IStorage.GetCounters()
	if err != nil {
		return counters, err
	}
	// Если нет в памяти данных, то идём в базу данных
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

// getCounters получение всех Counter из бд
func (storage *DBStorage) getCounters() (map[string]Counter, error) {
	counters := make(map[string]Counter)
	if storage.close {
		return counters, ErrorStorageDatabaseClosed
	}
	rows, err := storage.db.QueryContext(storage.storeCtx, "SELECT name, value FROM t_counter")
	if err != nil {
		return counters, err
	}
	// Закроем строки, чтобы освободить соединение
	defer func() {
		if rErr := rows.Close(); rErr != nil {
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

// SetGauges массовое обновление метрик Гауге
func (storage *DBStorage) SetGauges(gauges map[string]Gauge) (err error) {
	err = storage.IStorage.SetGauges(gauges)
	if err != nil || !storage.syncMode {
		return err
	}
	// Записываем в базу данных
	return storage.syncGauges(gauges)
}

// syncGauges запись в бд нескольких Gauge с ретраями
func (storage *DBStorage) syncGauges(gauges map[string]Gauge) (err error) {
	if storage.close {
		return ErrorStorageDatabaseClosed
	}
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

// setGauges массовое обновление гауге в базе
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

// AddCounters массовое обновление метрик Каунтер
func (storage *DBStorage) AddCounters(counters map[string]Counter) (err error) {
	err = storage.IStorage.AddCounters(counters)
	if err != nil || !storage.syncMode {
		return err
	}
	// Записываем в базу данных
	return storage.syncCounters(counters, false)
}

// syncCounters запись в бд нескольких Counter с ретраями
func (storage *DBStorage) syncCounters(counters map[string]Counter, clearAndSet bool) (err error) {
	if storage.close {
		return ErrorStorageDatabaseClosed
	}
	pause := time.Second
	var pgErr *pgconn.PgError
	for i := 0; i < 3; i++ {
		err = storage.addCounters(counters, clearAndSet)
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

// addCounters массовое обновление каунтер в базе
func (storage *DBStorage) addCounters(counters map[string]Counter, clearAndSet bool) error {
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
	queryString := "INSERT INTO t_counter (name, value) VALUES ($1, $2) on conflict (name) do update set value = t_counter.value + $2, updated_at = $3"
	if clearAndSet {
		queryString = "INSERT INTO t_counter (name, value) VALUES ($1, $2) on conflict (name) do update set value = $2, updated_at = $3"
	}
	prepared, err := tx.PrepareContext(storage.storeCtx, queryString)
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

// restore восстанавливаем данные из базы данных
func (storage *DBStorage) restore() error {
	if storage.close {
		return ErrorStorageDatabaseClosed
	}
	gauges, err := storage.GetGauges()
	if err != nil {
		return err
	}
	err = storage.IStorage.SetGauges(gauges)
	if err != nil {
		return err
	}
	counters, err := storage.GetCounters()
	if err != nil {
		return err
	}
	err = storage.IStorage.AddCounters(counters)
	if err != nil {
		return err
	}
	return nil
}

// clean удаляем данные из базы данных перед стартом без восстановления данных
func (storage *DBStorage) clean() error {
	if storage.close {
		return ErrorStorageDatabaseClosed
	}
	tx, err := storage.db.BeginTx(storage.storeCtx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if tErr := tx.Rollback(); tErr != nil && tErr.Error() != "sql: transaction has already been committed or rolled back" {
			logger.Log.Error(tErr)
		}
	}()
	_, err = tx.ExecContext(storage.storeCtx, "TRUNCATE TABLE t_gauge")
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(storage.storeCtx, "TRUNCATE TABLE t_counter")
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Flush записываем данные в базу данных перед закрытием
func (storage *DBStorage) Flush() error {
	if storage.close {
		return ErrorStorageDatabaseClosed
	}
	gauges, err := storage.IStorage.GetGauges()
	if err != nil {
		return err
	}
	if err = storage.syncGauges(gauges); err != nil {
		return err
	}
	counters, err := storage.IStorage.GetCounters()
	if err != nil {
		return err
	}
	if err = storage.syncCounters(counters, true); err != nil {
		return err
	}

	return nil
}

// Sync синхронизация данных хранилища в базу данных по таймеру
func (storage *DBStorage) Sync(ctx context.Context) error {
	interval := time.Duration(ctx.Value(contextkeys.SyncInterval).(int64)) * time.Second
	logger.Log.Infof("Sync metrics process starts. Period is %d seconds", interval/time.Second)
	ticker := time.NewTicker(interval)
	for {
		// Ловим закрытие контекста, чтобы завершить обработку
		select {
		case <-ticker.C:
			logger.Log.Debug("Sync metrics")
			if err := storage.Flush(); err != nil {
				logger.Log.Error(err)
			}
		case <-ctx.Done():
			logger.Log.Debug("Sync metrics before end")
			ticker.Stop()
			if err := storage.Flush(); err != nil {
				logger.Log.Warn("Error while syncing metrics")
				return err
				//logger.Log.Error(err)
			}
			logger.Log.Debug("Synced")
			return nil
		}
	}
}

// Close Закрытие хранилища
func (storage *DBStorage) Close() error {
	if storage.close {
		return ErrorStorageDatabaseClosed
	}
	storage.close = true
	return nil
}

// FlushAndClose синхронизация данных и закрытие хранилища
func (storage *DBStorage) FlushAndClose() error {
	if storage.close {
		return ErrorStorageDatabaseClosed
	}
	if err := storage.Flush(); err != nil {
		return err
	}
	return nil
}

// IsSyncMode открыто ли хранилище в синхронном режиме
func (storage *DBStorage) IsSyncMode() bool {
	return storage.syncMode
}
