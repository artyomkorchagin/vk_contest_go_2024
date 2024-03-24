package utils

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"testing"
	"time"
)

func TestFloodControl(t *testing.T) {
	cfg, err := LoadDbConfig()
	if err != nil {
		t.Fatal(err)
	}
	connectionStr := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s", cfg.Host, cfg.Port, cfg.Database, cfg.Username, cfg.Password)
	pool, err := pgxpool.Connect(context.Background(), connectionStr)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
	})

	_, err = pool.Exec(context.Background(), "CREATE TABLE test_flood (user_id BIGINT NOT NULL, request_time TIMESTAMP NOT NULL);")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), "DROP TABLE test_flood")
	})

	fc, err := NewFloodControl(cfg)
	if err != nil {
		t.Fatal(err)
	}
	fc.MaxTries = 10
	fc.Duration = time.Minute

	ok, err := fc.Check(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected flood control to pass")
	}

	for i := 0; i < 9; i++ {
		_, err := fc.Check(context.Background(), 1)
		if err != nil {
			t.Fatal(err)
		}
	}

	ok, err = fc.Check(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected flood control to fail")
	}

	// checking if other users request will pass
	ok, err = fc.Check(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected flood control to pass")
	}

	time.Sleep(time.Minute)

	ok, err = fc.Check(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected flood control to pass")
	}
}
