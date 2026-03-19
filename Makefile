.PHONY: run-memory run-postgres down test build clean

run-memory:
	docker compose up --build app-memory

run-postgres:
	docker compose up --build app-postgres db

down:
	docker compose down -v

test:
	go test ./...

build:
	go build -o bin/shortener ./cmd/shortener

clean:
	rm -rf bin/
