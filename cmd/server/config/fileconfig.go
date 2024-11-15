package config

import (
	incnf "gmetrics/internal/config"
)

type FileConfig struct {
	Address       string         `json:"address"`
	Restore       bool           `json:"restore"`
	StoreInterval incnf.Duration `json:"store_interval"`
	StoreFile     string         `json:"store_file"`
	DatabaseDsn   string         `json:"database_dsn"`
	CryptoKey     string         `json:"crypto_key"`
}
