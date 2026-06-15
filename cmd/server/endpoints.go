package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
)

func startEndpointUpdater(ctx context.Context, conn *pgx.Conn) http.Handler {

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {
		givencity.mu.Lock()
		defer givencity.mu.Unlock()

		givencity.city = chi.URLParam(r, "city")
		givencity.processingCities = append(givencity.processingCities, givencity.city)

		fmt.Printf("Requested city %s\n", givencity.city)

		var meteoValues MeteoValues
		err := conn.QueryRow(
			ctx,
			`SELECT city, timestamp, temperature, windspeed FROM meteovalues 
			WHERE city = $1 ORDER BY timestamp DESC LIMIT 1`,
			givencity.city).Scan(&meteoValues.Name, &meteoValues.Timestamp, &meteoValues.Temperature, &meteoValues.WindSpeed)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("not found"))
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal error"))
			return
		}

		raw, err := json.Marshal(meteoValues)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal error"))
			return
		}

		_, err = w.Write(raw)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal error"))
			return
		}
	})

	return r
}
