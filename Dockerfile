# Use the official Golang image to create a build artifact.
FROM golang:1.25.5 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Add the missing module and download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

ARG APP_VERSION=latest
ENV ENV_APP_VERSION=${APP_VERSION}
ARG COMMIT_ID=UNKNOWN
ENV ENV_COMMIT_ID=${COMMIT_ID}

# Build the Go app
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$ENV_APP_VERSION -X main.commit=$ENV_COMMIT_ID" -o build/hargassner-monitor main.go
# Run tests
RUN go test ./...

# Start a new stage from scratch
FROM alpine:3.23.2

# Set the Current Working Directory inside the container
WORKDIR /app/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/build/hargassner-monitor /app/

# OCI labels (Open Container Initiative)
# These provide metadata about the image and are compatible with the
# `org.opencontainers` label schema.
LABEL org.opencontainers.image.title="hargassner-monitor" \
	org.opencontainers.image.description="Reads and parses status records from a Hargassner HSV heating system via a serial connection" \
	org.opencontainers.image.url="https://github.com/rhierlmeier/hargassner-monitor" \
	org.opencontainers.image.source="https://github.com/rhierlmeier/hargassner-monitor" \
	org.opencontainers.image.licenses="GPL" \
	org.opencontainers.image.authors="rhierlmeier" \
	org.opencontainers.image.version="0.0.0" \
	org.opencontainers.image.revision="${VCS_REF:-unknown}" \
	org.opencontainers.image.created="${BUILD_DATE:-unknown}"

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
