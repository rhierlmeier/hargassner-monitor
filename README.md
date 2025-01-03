
[![ci](https://github.com/rhierlmeier/hargassner-monitor/actions/workflows/docker-image.yaml/badge.svg)](https://github.com/rhierlmeier/hargassner-monitor/actions/workflows/docker-image.yaml)

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
docker run --rm \
    -e HARGASSNER_SERIAL_PORT=
    -e HARGASSNER_MQTT_BROKER=tcp://mqtt.local
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
- `HARGASSNER_MONITOR_PORT`: Port where the HTTP server first status request is listing


## MQTT Homie Devices, Nodes, and Properties

The application publishes the status values of the Hargassner heating system. It follows the [MQTT Homie specification](https://homieiot.github.io/specification/). Below is the structure of the Homie device, nodes, and properties used in this application.

### Device

- **ID**: `hargassner`
- **Name**: `Hargassner Heizung`

### Nodes

#### Prozesswerte

- **ID**: `prozesswerte`
- **Name**: `Prozesswerte`
- **Type**: `Prozesswerte`

##### Properties

- **ID**: `rauchgasTemperatur`
  - **Name**: `Rauchgas Temperatur`
  - **Type**: `float`
  - **Unit**: `°C`

- **ID**: `boiler1Temperatur`
  - **Name**: `Boiler 1 Temperatur`
  - **Type**: `float`
  - **Unit**: `°C`

- **ID**: `aussenTemperaturAktuell`
  - **Name**: `Aussentemperatur aktuell`
  - **Type**: `float`
  - **Unit**: `°C`

- **ID**: `aussenTemperaturGemittelt`
  - **Name**: `Aussentemperatur gemittelt`
  - **Type**: `float`
  - **Unit**: `°C`

- **ID**: `kesselTemperatur`
  - **Name**: `Kesseltemperatur`
  - **Type**: `float`
  - **Unit**: `°C`

- **ID**: `kesselSollTemperatur`
  - **Name**: `Kesselsolltemperatur`
  - **Type**: `float`
  - **Unit**: `°C`

- **ID**: `meldung`
  - **Name**: `Meldung`
  - **Type**: `string`

- **ID**: `saugluftGeblaese`
  - **Name**: `Saugluftgebläse`
  - **Type**: `float`
  - **Unit**: `°C`

- **ID**: `primaerLuftGeblaese`
  - **Name**: `Primärluftgebläse`
  - **Type**: `float`
  - **Unit**: `°C`

- **ID**: `o2InAbgas`
  - **Name**: `O2 in Abgas`
  - **Type**: `float`
  - **Unit**: `%`

- **ID**: `foerderMenge`
  - **Name**: `Fördermenge`
  - **Type**: `float`
  - **Unit**: `%`

- **ID**: `stromRaumaustragung`
  - **Name**: `Strom Raumaustragung`
  - **Type**: `float`
  - **Unit**: `A`

- **ID**: `stromAscheaustragung`
  - **Name**: `Strom Ascheaustragung`
  - **Type**: `float`
  - **Unit**: `A`

- **ID**: `stromEinschub`
  - **Name**: `Strom Einschub`
  - **Type**: `float`
  - **Unit**: `A`

#### Heizkreis 1

- **ID**: `heizkreis1`
- **Name**: `Heizkreis 1`
- **Type**: `Heizkreis 1`

##### Properties

- **ID**: `vorlaufTemperatur`
  - **Name**: `Vorlauftemperatur`
  - **Type**: `float`
  - **Unit**: `°C`

- **ID**: `vorlaufSollTemperatur`
  - **Name**: `Vorlauf Solltemperatur`
  - **Type**: `float`
  - **Unit**: `°C`

#### Heizkreis 2

- **ID**: `heizkreis2`
- **Name**: `Heizkreis 2`
- **Type**: `Heizkreis 2`

##### Properties

- **ID**: `vorlaufTemperatur`
  - **Name**: `Vorlauftemperatur`
  - **Type**: `float`
  - **Unit**: `°C`

- **ID**: `vorlaufSollTemperatur`
  - **Name**: `Vorlauf Solltemperatur`
  - **Type**: `float`
  - **Unit**: `°C`

#### Störung

- **ID**: `stoerung`
- **Name**: `Störung`
- **Type**: `Störung`

##### Properties

- **ID**: `nr`
  - **Name**: `Nummer`
  - **Type**: `integer`

- **ID**: `text`
  - **Name**: `Text`
  - **Type**: `string`