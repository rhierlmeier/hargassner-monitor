# Hargassner Monitor
Hargassner Monitor is a Go application that reads and parses status records from a Hargassner HSV heating system via a serial connection. The application can be built and run locally or within a Docker container.

## Features

- Reads status records from a serial device
- Parses and displays status records
- Multi-architecture Docker image support (amd64 and arm64)

## Prerequisites

- Go 1.17 or later
- Docker (for building and running the Docker image)

## Getting Started

### Clone the Repository

```sh
git clone https://github.com/rhierlmeier/hargassner-monitor.git
cd hargassner-monitor
```
Test
### Docker Images

The project provides pre-built Docker images for different architectures. You can pull the appropriate image for your system from Docker Hub.

#### Pulling the Docker Image

To pull the Docker image for `amd64` architecture:

```sh
docker pull rhierlmeier/hargassner-monitor:amd64-latest
```

To pull the Docker image for `arm64` architecture:

```sh
docker pull rhierlmeier/hargassner-monitor:arm64-latest
```

#### Running the Docker Container

To run the Docker container:

```sh
docker run --rm -it \
    --device=/dev/ttyUSB0 \
    rhierlmeier/hargassner-monitor:latest
```

Replace `/dev/ttyUSB0` with the appropriate serial device on your system.

# Environment Variables

The application uses the following environment variables:

- `HARGASSNER_SERIAL_PORT`: Specifies the serial port to which the Hargassner heating system is connected. Default is `/dev/ttyUSB0`.
- `HARGASSNER_MQTT_BROKER`: Specifies the MQTT broker URL. Default is `tcp://localhost:1883`.
- `HARGASSNER_MQTT_CLIENT_ID`: Specifies the MQTT client ID. Default is `hargassner-monitor`.
- `HARGASSNER_MQTT_USERNAME`: Specifies the username for MQTT broker authentication. Default is empty.
- `HARGASSNER_MQTT_PASSWORD`: Specifies the password for MQTT broker authentication. Default is empty.
