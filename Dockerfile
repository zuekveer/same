FROM golang:1.23.0 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy && go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o application cmd/main.go

FROM alpine:3.20.2

WORKDIR /app/
COPY --from=builder /app/application .

COPY internal/config/config.yaml /app/internal/config/config.yaml
COPY .env /app/.env

CMD ["./application"]
