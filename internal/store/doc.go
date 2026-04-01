// Package store implements the SQLite persistence layer using
// modernc.org/sqlite (pure Go, no CGO). It provides migrations
// (tracked in schema_migrations so each .sql file runs once),
// CRUD operations, and FTS5 full-text search for sponsor lookups.
package store
