package database

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB глобальный пул подключений к базе данных для приложения
var DB *sql.DB

// NewPgDB создаёт новое подключение к базе данных
func NewPgDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	// Если дсн не передан, то просто возвращаем созданный пул, он не работоспособен
	if dsn == "" {
		return db, nil
	}
	// Сразу проверим работоспособность соединения
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
