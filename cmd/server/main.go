package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	httpPort = ":3000"
	connStr  = "postgres://andk228:password@localhost:5433/weather"
)

type GivenCity struct {
	city string
	mu   sync.RWMutex
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
	conn, err = connectToDB(ctx, connStr)
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
