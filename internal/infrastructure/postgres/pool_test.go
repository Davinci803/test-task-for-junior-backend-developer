package postgres

import (
	"context"
	"testing"
	"time"
)

func TestOpenValidation(t *testing.T) {
	_, err := Open(context.Background(), "")
	if err == nil {
		t.Fatalf("expected error for empty dsn")
	}

	_, err = Open(context.Background(), "://bad-dsn")
	if err == nil {
		t.Fatalf("expected parse error for invalid dsn")
	}
}

func TestOpenPingFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, err := Open(ctx, "postgres://postgres:postgres@127.0.0.1:65432/taskservice?sslmode=disable")
	if err == nil {
		t.Fatalf("expected connection/ping error")
	}
}

func TestOpenSuccess(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pool, err := Open(ctx, "postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable")
	if err != nil {
		t.Fatalf("expected successful open against local postgres: %v", err)
	}
	defer pool.Close()
}
