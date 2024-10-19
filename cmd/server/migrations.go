package main

import (
	"database/sql"

	"github.com/lopezator/migrator"
)

func migrations() (*migrator.Migrator, error) {
	// Configure migrations
	m, err := migrator.New(
		migrator.Migrations(
			&migrator.Migration{
				Name: "Create metric tables",
				Func: func(tx *sql.Tx) error {
					if _, err := tx.Exec("CREATE TABLE t_gauge (name VARCHAR PRIMARY KEY, value double precision);"); err != nil {
						return err
					}
					if _, err := tx.Exec("CREATE TABLE t_counter (name VARCHAR PRIMARY KEY, value bigint);"); err != nil {
						return err
					}
					return nil
				},
			},
			&migrator.Migration{
				Name: "Add timestamps to metrics",
				Func: func(tx *sql.Tx) error {
					if _, err := tx.Exec("alter table public.t_counter add created_at timestamp without time zone default now();"); err != nil {
						return err
					}
					if _, err := tx.Exec("alter table public.t_counter add updated_at timestamp without time zone default now();"); err != nil {
						return err
					}
					return nil
				},
			},
			&migrator.Migration{
				Name: "Add timestamps to gauge",
				Func: func(tx *sql.Tx) error {
					if _, err := tx.Exec("alter table public.t_gauge add created_at timestamp without time zone default now();"); err != nil {
						return err
					}
					if _, err := tx.Exec("alter table public.t_gauge add updated_at timestamp without time zone default now();"); err != nil {
						return err
					}
					return nil
				},
			},
		),
	)

	return m, err
}
