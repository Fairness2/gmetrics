package config

import (
	incnf "gmetrics/internal/config"
)

type FileConfig struct {
	Address       string         `json:"address"`
	RPCAddress    string         `json:"rpc_address"`
	Restore       bool           `json:"restore"`
	StoreInterval incnf.Duration `json:"store_interval"`
	StoreFile     string         `json:"store_file"`
	DatabaseDsn   string         `json:"database_dsn"`
	CryptoKey     string         `json:"crypto_key"`
	TrustedSubnet string         `json:"trusted_subnet"`
}
