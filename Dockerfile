# Use the official Golang image to create a build artifact.
FROM golang:1.17 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Add the missing module and download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN mkdir -p build && go build -o build/hargassner-monitor main.go
# Run tests
RUN go test ./...
# Start a new stage from scratch
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/build/hargassner-monitor .

# Command to run the executable
CMD ["./hargassner-monitor"]
