# Use the official Golang image to create a build artifact.
FROM golang:1.25.5 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Add the missing module and download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Build the Go app
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o build/hargassner-monitor main.go
# Run tests
RUN go test ./...

# Start a new stage from scratch
FROM alpine:3.23.2

# Set the Current Working Directory inside the container
WORKDIR /app/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/build/hargassner-monitor /app/

ENV HARGASSNER_MONITOR_PORT=8080
ENV HARGASSNER_SERIAL_DEVICE=/dev/ttyUSB0
ENV HARGASSNER_MQTT_CLIENT_ID=hargassner-monitor
# MQTT broker address (e.g tcp://localhost:1883 or ssl://localhost:8883)
ENV HARGASSNER_MQTT_BROKER=tcp://localhost:1883
# Optional username and password for MQTT
ENV HARGASSNER_MQTT_USER=""
ENV HARGASSNER_MQTT_PASSWORD=""

ENV HARGASSNER_MONITOR_PORT=8080

EXPOSE $HARGASSNER_MONITOR_PORT


# Command to run the executable
ENTRYPOINT ["/app/hargassner-monitor"]
