package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denniskbijo/visa-tracker/internal/config"
	"github.com/denniskbijo/visa-tracker/internal/handlers"
	"github.com/denniskbijo/visa-tracker/internal/ingest"
	"github.com/denniskbijo/visa-tracker/internal/store"
)

func main() {
	migrateOnly := flag.Bool("migrate", false, "run migrations and exit")
	ingestOnly := flag.Bool("ingest", false, "run data ingestion and exit")
	flag.Parse()

	cfg := config.Load()

	db, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.MigrateFromDir(cfg.MigrationsDir); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	if *migrateOnly {
		log.Println("migrations complete")
		return
	}

	ingester := ingest.New(db, cfg)
	if err := ingester.LoadSeedData(); err != nil {
		log.Fatalf("failed to load seed data: %v", err)
	}

	if *ingestOnly {
		if err := ingester.RefreshSponsors(); err != nil {
			log.Fatalf("sponsor ingestion failed: %v", err)
		}
		log.Println("ingestion complete")
		return
	}

	go ingester.StartScheduler()

	mux := http.NewServeMux()
	h := handlers.New(db, cfg)
	h.Register(mux)

	addr := ":" + cfg.Port
	srv := &http.Server{
		Addr:              addr,
		Handler:           h.Wrapped(mux),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	log.Printf("visa-tracker listening on http://localhost%s", addr)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
