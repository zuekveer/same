FROM golang:1.21.0 AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o application cmd/main.go

# Start a new stage from scratch
FROM alpine:3.20.2

# Set the working directory inside the container
WORKDIR /root/

# Copy the pre-built binary from the previous stage
COPY --from=builder /app/application .

# Expose port 8088 to the outside world
EXPOSE 8088

# Command to run the executable
CMD ["./application"]