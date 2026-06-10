package main

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func connectToDB(ctx context.Context, connStr string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func createTable(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, "CREATE TABLE IF NOT EXISTS meteovalues (city TEXT, timestamp TIMESTAMPTZ, temperature FLOAT8)")
	return err
}
