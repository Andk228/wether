package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Andk228/wether/internal/client/http/geocoding"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/go-co-op/gocron/v2"
)

const httpPort = ":3000"

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	geocodingClient := geocoding.NewClient(httpClient)

	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")

		resp, err := geocodingClient.GetCity(city)
		if err != nil {
			log.Print(err)
		}

		raw, err := json.Marshal(resp)
		if err != nil {
			log.Println(err)
		}

		_, err = w.Write(raw)
		if err != nil {
			log.Print(err)
		}
	})

	s, err := gocron.NewScheduler()
	if err != nil {
		panic(err)
	}

	jobs, err := initJobs(s)
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

func initJobs(scheduler gocron.Scheduler) ([]gocron.Job, error) {

	// add a job to the scheduler
	j, err := scheduler.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func() {
				fmt.Println("New task")
			},
		),
	)
	if err != nil {
		return nil, err
	}

	return []gocron.Job{j}, nil

}

// func runCron() error {
// 	// each job has a unique id
// 	fmt.Println(j.ID())

// 	// start the scheduler
// 	s.Start()
// 	// block until you are ready to shut down
// 	select {
// 	case <-time.After(time.Minute):
// 	}

// 	// when you're done, shut it down
// 	err = s.Shutdown()
// 	// or for context-aware teardown:
// 	// err = s.ShutdownWithContext(ctx)
// 	if err != nil {
// 		// handle error
// 	}
// }
