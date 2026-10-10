package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                 string
	OenexaRPCURL         string
	DatabaseDSN          string
	DefaultCommissionPct float64
	MinCommissionPct     float64
	MaxCommissionPct     float64
	OenEurExchangeRate   float64
	LegacyAggregatorFee  float64
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	rpcURL := os.Getenv("OENEXA_RPC_URL")
	if rpcURL == "" {
		rpcURL = "http://localhost:8545"
	}

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=oengo port=5432 sslmode=disable"
	}

	commPct := 5.0
	if val := os.Getenv("DEFAULT_COMMISSION_PCT"); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			commPct = f
		}
	}

	return &Config{
		Port:                 port,
		OenexaRPCURL:         rpcURL,
		DatabaseDSN:          dsn,
		DefaultCommissionPct: commPct,
		MinCommissionPct:     0.0,
		MaxCommissionPct:     30.0,
		OenEurExchangeRate:   13.60,
		LegacyAggregatorFee:  30.0,
	}
}
