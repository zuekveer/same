FROM golang:1.23.0 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy && go mod download
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o application cmd/main.go

FROM alpine:3.20.2

WORKDIR /app/
COPY --from=builder /app/application .
COPY --from=builder /go/bin/goose /usr/local/bin/goose

COPY internal/database/migrations /app/internal/database/migrations

COPY internal/config/config.yaml /app/internal/config/config.yaml
COPY .env /app/.env

CMD ["./application"]