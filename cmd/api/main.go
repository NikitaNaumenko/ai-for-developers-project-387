package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nikitanaumenko/calendar/internal/config"
	"github.com/nikitanaumenko/calendar/internal/db"
	"github.com/nikitanaumenko/calendar/internal/handler"
	"github.com/nikitanaumenko/calendar/internal/httpserver"
	"github.com/nikitanaumenko/calendar/internal/storage"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("load config", zap.Error(err))
	}

	var store handler.Store
	if cfg.DatabaseURL == "" {
		log.Info("database url is not configured; using in-memory storage")
		store = storage.NewMemoryStore()
	} else {
		pool, err := storage.OpenPostgres(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatal("open postgres", zap.Error(err))
		}
		defer pool.Close()
		store = db.New(pool)
	}

	apiHandler := handler.New(store)
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpserver.NewRouter(log, apiHandler, cfg.StaticDir),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("http server started", zap.String("addr", cfg.HTTPAddr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("http server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("http server shutdown failed", zap.Error(err))
	}
}
