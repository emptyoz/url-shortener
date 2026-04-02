package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Vadim-Makhnev/url-shortener/internal/config"
	"github.com/Vadim-Makhnev/url-shortener/internal/handler"
	"github.com/Vadim-Makhnev/url-shortener/internal/metrics"
	"github.com/Vadim-Makhnev/url-shortener/internal/repository"
	"github.com/Vadim-Makhnev/url-shortener/internal/service"
)

type application struct {
	log     *slog.Logger
	handler *handler.Handler
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "./config.yaml", "application config path")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	metrics.InitMetrics()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	storage, err := repository.NewPostgreSQL(log, cfg.DBAddress)
	if err != nil {
		log.Error("connect to PostgreSQL", "error", err)
		os.Exit(1)
	}

	cache, err := repository.NewRedis(log, cfg.RedisURL)
	if err != nil {
		log.Error("connect to Redis", "error", err)
		os.Exit(1)
	}

	urlService := service.NewService(log, storage, cache)
	urlHandler := handler.NewHandler(log, urlService, cfg.BaseURL)

	app := application{
		log:     log,
		handler: urlHandler,
	}

	srv := &http.Server{
		Handler:      app.routes(),
		Addr:         cfg.SrvAddress,
		ErrorLog:     slog.NewLogLogger(log.Handler(), slog.LevelError),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Info("server starting", "address", srv.Addr)

	err = srv.ListenAndServe()
	log.Error("server connection", "address", srv.Addr, "error", err)
	os.Exit(1)
}
