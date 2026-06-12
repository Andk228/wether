package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

const httpPort = ":3000"

func envOrDefault(key, def string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return def
}

func buildConnStr() string {
	dbHost := envOrDefault("DB_HOST", "localhost")
	dbPort := envOrDefault("DB_PORT", "5433")
	dbUser := envOrDefault("DB_USER", "andk228")
	dbPassword := envOrDefault("DB_PASSWORD", "password")
	dbName := envOrDefault("DB_NAME", "weather")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbHost, dbPort, dbName)
}

type GivenCity struct {
	city             string
	processingCities []string
	mu               sync.RWMutex
}

type MeteoValues struct {
	Name        string    `db:"name"`
	Timestamp   time.Time `db:"timestamp"`
	Temperature float64   `db:"temperature"`
}

var givencity = &GivenCity{}

func main() {
	var conn *pgx.Conn
	var err error

	ctx := context.Background()
	conn, err = connectToDB(ctx, buildConnStr())
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	err = createTable(ctx, conn)
	if err != nil {
		panic(err)
	}

	r := startEndpointUpdater(ctx, conn)

	err = startCron(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Start server")
		err := http.ListenAndServe(httpPort, r)
		if err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}
