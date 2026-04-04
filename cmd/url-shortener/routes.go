package main

import (
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (app *application) routes() http.Handler {
	r := http.NewServeMux()

	r.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		res := struct {
			Status string `json:"status"`
		}{
			Status: "OK",
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(&res); err != nil {
			app.log.Error("json endcoder", "error", err)
		}
	})

	r.HandleFunc("POST /api/shorten", app.handler.ShortenURL)
	r.HandleFunc("GET /api/urls", app.handler.GetURLs)

	r.Handle("GET /metrics", promhttp.Handler())

	r.HandleFunc("GET /{shortCode}", app.handler.RedirectURL)

	return r
}
