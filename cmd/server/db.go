package main

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

func connectToDB(ctx context.Context, connStr string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func connectToDBWithRetry(ctx context.Context, connStr string, attempts int, delay time.Duration) (*pgx.Conn, error) {
	var conn *pgx.Conn
	var err error

	for i := 1; i <= attempts; i++ {
		conn, err = connectToDB(ctx, connStr)
		if err == nil {
			return conn, nil
		}

		log.Printf("Postgres not ready yet (%d/%d): %v", i, attempts, err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
	}

	return nil, err
}

func createTable(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS meteovalues (city TEXT, timestamp TIMESTAMPTZ, temperature FLOAT8, windspeed FLOAT8)")
	return err
}
