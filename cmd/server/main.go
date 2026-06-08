package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Andk228/wether/internal/client/http/geocoding"
	openmeteo "github.com/Andk228/wether/internal/client/http/open_meteo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/go-co-op/gocron/v2"
)

const (
	httpPort = ":3000"
)

type GivenCity struct {
	city string
	mu   sync.RWMutex
}

type MeteoValues struct {
	Timestamp   time.Time
	Temperature float64
}

type Storage struct {
	data map[string][]MeteoValues
	mu   sync.RWMutex
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	storage := &Storage{
		data: make(map[string][]MeteoValues),
	}

	givencity := &GivenCity{}

	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {
		givencity.mu.Lock()
		defer givencity.mu.Unlock()

		givencity.city = chi.URLParam(r, "city")

		fmt.Printf("Requested city %s\n", givencity.city)

		storage.mu.RLock()
		defer storage.mu.RUnlock()

		mateoValues, ok := storage.data[givencity.city]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
			return
		}

		meteoraw, err := json.Marshal(mateoValues)
		if err != nil {
			log.Println(err)

		}

		_, err = w.Write(meteoraw)
		if err != nil {
			log.Print(err)
		}

	})

	s, err := gocron.NewScheduler()
	if err != nil {
		panic(err)
	}

	jobs, err := initJobs(s, storage, givencity)
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}

	wg.Add(2)
	go func() {
		defer wg.Done()
		fmt.Println("Start server")
		err := http.ListenAndServe(httpPort, r)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		fmt.Printf("Starting job: %v\n", jobs[0].ID())
		s.Start()
	}()

	wg.Wait()
}

func initJobs(scheduler gocron.Scheduler, storage *Storage, givencity *GivenCity) ([]gocron.Job, error) {
	httpClient := &http.Client{
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
				givencity.mu.RLock()
				defer givencity.mu.RUnlock()
				resp, err := geocodingClient.GetCoordinates(givencity.city)
				if err != nil {
					log.Print(err)
				}

				meteoresp, err := openmeteoClient.GetTemperature(resp.Latitude, resp.Longitude)
				if err != nil {
					log.Print(err)
				}

				storage.mu.Lock()
				defer storage.mu.Unlock()

				timestamp, err := time.Parse("2006-01-2T15:04", meteoresp.Current.Time)
				if err != nil {
					log.Println(err)
				}

				storage.data[givencity.city] = append(storage.data[givencity.city], MeteoValues{
					Timestamp:   timestamp,
					Temperature: meteoresp.Current.Temperature2m,
				})
				fmt.Println("Data was updated")
			},
		),
	)
	if err != nil {
		return nil, err
	}

	return []gocron.Job{j}, nil

}
