
include .env
export

migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1

migrate-force:
	migrate -path db/migrations -database "$(DATABASE_URL)" force $(VERSION)

migrate-version:
	migrate -path db/migrations -database "$(DATABASE_URL)" version

build:
	go build -o bin/server cmd/server/main.go

dev:
	air
