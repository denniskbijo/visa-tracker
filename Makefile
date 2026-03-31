.PHONY: build run dev test clean ingest migrate

BINARY := visa-tracker
DB     := visa-tracker.db

build:
	go build -o $(BINARY) ./cmd/server

run: build
	./$(BINARY)

dev:
	go run ./cmd/server

test:
	go test ./...

clean:
	rm -f $(BINARY) $(DB)

ingest: build
	./$(BINARY) -ingest

migrate: build
	./$(BINARY) -migrate
