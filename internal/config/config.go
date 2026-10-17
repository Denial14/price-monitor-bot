package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser          string
	DBPassword      string
	DBName          string
	DBHost          string
	DBPort          string
	BotToken        string
	ScanInterval    time.Duration
	Workers         int
	RateRPS         float64
	RateBurst       int
	ShutdownTimeout time.Duration
	PollInterval    time.Duration
}

func getDefaultEnv(key string, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	host := getDefaultEnv("DB_HOST", "localhost")
	port := getDefaultEnv("DB_PORT", "5432")

	workersStr := getDefaultEnv("WORKERS", "10")
	workers, err := strconv.Atoi(workersStr)
	if err != nil {
		return nil, fmt.Errorf("WORKERS must be a number, but got: %q: %w", workersStr, err)
	}

	intervalStr := getDefaultEnv("SCAN_INTERVAL", "15m")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		return nil, fmt.Errorf("SCAN_INTERVAL must be valid time duration, but got: %q: %w", intervalStr, err)
	}

	rpsStr := getDefaultEnv("RATE_RPS", "2")
	rpc, err := strconv.ParseFloat(rpsStr, 64)
	if err != nil {
		return nil, fmt.Errorf("RPC must be a valid double, but got: %q: %w", rpsStr, err)
	}

	burstStr := getDefaultEnv("RATE_BURST", "5")
	burst, err := strconv.Atoi(burstStr)
	if err != nil {
		return nil, fmt.Errorf("BURST must be a valid number, but got: %q: %w", burstStr, err)
	}

	botToken := os.Getenv("BOT_TOKEN")

	shutdownStr := getDefaultEnv("SHUTDOWN_TIMEOUT", "10s")
	shutdown, err := time.ParseDuration(shutdownStr)
	if err != nil {
		return nil, fmt.Errorf("shutdown timeout must be valid time duration, but got %q: %w", shutdownStr, err)
	}

	pollStr := getDefaultEnv("POLL_INTERVAL", "1s")
	poll, err := time.ParseDuration(pollStr)
	if err != nil {
		return nil, fmt.Errorf("POLL_INTERVAL must be a duration, got %q: %w", pollStr, err)
	}

	cfg := &Config{
		DBUser:          user,
		DBName:          name,
		DBPassword:      password,
		DBHost:          host,
		DBPort:          port,
		BotToken:        botToken,
		ScanInterval:    interval,
		RateRPS:         rpc,
		RateBurst:       burst,
		Workers:         workers,
		ShutdownTimeout: shutdown,
		PollInterval:    poll,
	}

	return cfg, validation(cfg)
}

func validation(c *Config) error {
	errs := make([]error, 0)
	if c.DBPassword == "" {
		errs = append(errs, errors.New("db_password is empty"))
	}
	if c.DBName == "" {
		errs = append(errs, errors.New("database name is empty"))
	}
	if c.DBUser == "" {
		errs = append(errs, errors.New("database user is empty"))
	}
	if c.Workers <= 0 || c.Workers > 1000 {
		errs = append(errs, errors.New("workers must be a range 1..1000"))
	}
	if c.PollInterval <= 0 {
		errs = append(errs, errors.New("pollInterval must be > 0"))
	}
	if c.RateBurst <= 0 {
		errs = append(errs, errors.New("rateBurst must be > 0"))
	}
	if c.ScanInterval <= 0 {
		errs = append(errs, errors.New("scanInterval must be > 0"))
	}
	if c.RateRPS <= 0 {
		errs = append(errs, errors.New("rateRPC must be > 0"))
	}
	if c.ShutdownTimeout <= 0 {
		errs = append(errs, errors.New("shutdown must be > 0"))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}
