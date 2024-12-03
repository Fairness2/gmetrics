package database

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDB(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		driver  string
		wantErr error
	}{
		{
			name:    "valid_dsn",
			dsn:     ":memory:",
			driver:  "sqlite3",
			wantErr: nil,
		},
		{
			name:    "valid_dsn_but_not_pingable",
			dsn:     "postgresql://postgres:example@127.0.0.1:5432/gmetrics123",
			driver:  "pgx",
			wantErr: ErrorCantPing,
		},
		{
			name:    "empty dsn",
			dsn:     "",
			driver:  "pgx",
			wantErr: nil,
		},
		{
			name:    "invalid dsn",
			dsn:     "invalidDSN",
			driver:  "pgx",
			wantErr: ErrorCantPing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewDB(tt.driver, tt.dsn)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				defer db.Close()
				assert.NoError(t, err)
				assert.NotNil(t, db)
			}
		})
	}
}
