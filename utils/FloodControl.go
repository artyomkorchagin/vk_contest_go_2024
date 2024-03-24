package utils

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

var (
	floodcfg  = ".internal/config/floodcfg.yaml"
	servercfg = ".internal/config/servercfg.yaml"
)

type PostgresConnectionConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
}

type FloodControl struct {
	MaxTries int `yaml:"maxTries"` // K - количество вызовов
	Interval int `yaml:"interval"` // N - секунд
	Db       *pgx.Conn
	Duration time.Duration
}

func LoadDbConfig() (*PostgresConnectionConfig, error) {
	cfg := PostgresConnectionConfig{}
	file, err := os.Open(servercfg)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data := yaml.NewDecoder(file)

	if err := data.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func NewFloodControl(cfg *PostgresConnectionConfig) (*FloodControl, error) {
	fc := FloodControl{}
	connectionStr := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s", cfg.Host, cfg.Port, cfg.Database, cfg.Username, cfg.Password)
	conn, err := pgx.Connect(context.Background(), connectionStr)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(floodcfg)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data := yaml.NewDecoder(file)

	if err := data.Decode(&fc); err != nil {
		return nil, err
	}
	fc.Duration = time.Second * time.Duration(fc.Interval)
	fc.Db = conn
	return &fc, nil
}

func (fc *FloodControl) Check(ctx context.Context, userID int64) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, fc.Duration)
	defer cancel()

	now := time.Now()
	startTime := now.Add(-fc.Duration)
	endTime := now

	var count int
	err := fc.Db.QueryRow(ctx, "SELECT COUNT(*) FROM requests WHERE user_id=$1 AND request_time BETWEEN $2 AND $3",
		userID, startTime, endTime).Scan(&count)
	if err != nil {
		return false, err
	}

	if count > fc.MaxTries {
		return false, nil
	}

	_, err = fc.Db.Exec(ctx, "INSERT INTO requests (user_id, request_time) VALUES ($1, $2)",
		userID, now)
	if err != nil {
		return false, err
	}

	return true, nil
}
