.PHONY: test test-race lint run migrate migrate-up migrate-down docker-build

# Windows without GNU Make:  make.cmd test   OR   pwsh -File scripts/ps/dev.ps1 test
# Optional: winget install GnuWin32.Make

# Default test target avoids -race on Windows without CGO; CI runs -race on Linux.
test:
	go test ./...

test-race:
	go test -race ./...

lint:
	golangci-lint run

run:
	go run ./cmd/server

migrate: migrate-up

migrate-up:
	go run ./cmd/migrate -direction up

migrate-down:
	go run ./cmd/migrate -direction down

docker-build:
	docker build -t internal-comm-backend:local .
