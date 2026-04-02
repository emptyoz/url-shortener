package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (app *application) routes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		res := struct {
			Status string `json:"status"`
		}{
			Status: "OK",
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(&res); err != nil {
			app.log.Error("json endcoder", "error", err)
		}
	}).Methods("GET")

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/shorten", app.handler.ShortenURL).Methods("POST")
	api.HandleFunc("/urls", app.handler.GetURLs).Methods("GET")

	r.Handle("/metrics", promhttp.Handler())

	r.HandleFunc("/{shortCode}", app.handler.RedirectURL).Methods("GET")

	return r
}
