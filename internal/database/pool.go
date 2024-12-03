package database

import (
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB глобальный пул подключений к базе данных для приложения
var DB *sql.DB

var (
	ErrorCantPing = errors.New("can't ping db")
	ErrorCantOpen = errors.New("can't open db")
)

// NewDB создаёт новое подключение к базе данных
func NewDB(driver string, dsn string) (*sql.DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, errors.Join(ErrorCantOpen, err)
	}
	// Если дсн не передан, то просто возвращаем созданный пул, он не работоспособен
	if dsn == "" {
		return db, nil
	}
	// Сразу проверим работоспособность соединения
	if err = db.Ping(); err != nil {
		return nil, errors.Join(ErrorCantPing, err)
	}
	return db, nil
}
