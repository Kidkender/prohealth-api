FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/server cmd/server/main.go && \
    go build -o bin/seed cmd/seed/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /go/bin/migrate .
COPY --from=builder /app/bin/server .
COPY --from=builder /app/bin/seed .
COPY --from=builder /app/db/migrations ./migrations
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

EXPOSE 8080

CMD ["./entrypoint.sh"]
