package config

import (
	incnf "gmetrics/internal/config"
)

type FileConfig struct {
	Address        string         `json:"address"`
	ReportInterval incnf.Duration `json:"report_interval"`
	PollInterval   incnf.Duration `json:"poll_interval"`
	CryptoKey      string         `json:"crypto_key"`
}
