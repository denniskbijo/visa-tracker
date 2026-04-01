// Package main is the entry point for the visa-tracker server.
// It initialises the database, runs migrations, loads seed data,
// starts the sponsor data refresh scheduler, and serves HTTP.
//
// Usage:
//
//	visa-tracker              # start the web server (default :8080)
//	visa-tracker -migrate     # run migrations and exit
//	visa-tracker -ingest      # run data ingestion and exit
package main
