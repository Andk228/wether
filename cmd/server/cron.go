package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Andk228/wether/internal/client/http/geocoding"
	openmeteo "github.com/Andk228/wether/internal/client/http/open_meteo"
	"github.com/go-co-op/gocron/v2"
	"github.com/jackc/pgx/v5"
)

func startCron(ctx context.Context, conn *pgx.Conn) error {
	s, err := gocron.NewScheduler()
	if err != nil {
		log.Fatal(err)
	}

	jobs, err := initJobs(ctx, s, conn)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		fmt.Printf("Starting job: %v\n", jobs[0].ID())
		s.Start()
	}()

	return nil
}

func initJobs(ctx context.Context, scheduler gocron.Scheduler, conn *pgx.Conn) ([]gocron.Job, error) {
	// transport := &http.Transport{
	// 	DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
	// 		return (&net.Dialer{
	// 			Timeout:   10 * time.Second,
	// 			KeepAlive: 30 * time.Second,
	// 		}).DialContext(ctx, "tcp4", addr)
	// 	},
	// 	TLSHandshakeTimeout: 15 * time.Second,
	// }

	httpClient := &http.Client{
		// Transport: transport,
		Timeout: 10 * time.Second,
	}

	geocodingClient := geocoding.NewClient(httpClient)
	openmeteoClient := openmeteo.NewClient(httpClient)
	j, err := scheduler.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func() {
				givencity.mu.Lock()
				cities := append([]string(nil), givencity.processingCities...)
				givencity.processingCities = nil
				givencity.mu.Unlock()

				for _, city := range cities {
					resp, err := geocodingClient.GetCoordinates(city)
					if err != nil {
						log.Print(err)
						continue
					}

					meteoresp, err := openmeteoClient.GetMeteoResp(resp.Latitude, resp.Longitude)
					if err != nil {
						log.Print(err)
						continue
					}

					timestamp, err := time.Parse("2006-01-2T15:04", meteoresp.Current.Time)
					if err != nil {
						log.Println(err)
						continue
					}

					_, err = conn.Exec(ctx, "INSERT INTO meteovalues (city, timestamp, temperature, windspeed) VALUES ($1, $2, $3, $4)",
						city, timestamp, meteoresp.Current.Temperature2m, meteoresp.Current.WindSpeed)
					if err != nil {
						log.Print(err)
					}
				}

				fmt.Println("Data was updated")
			},
		),
	)
	if err != nil {
		return nil, err
	}

	return []gocron.Job{j}, nil

}
