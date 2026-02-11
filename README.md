
[![ci](https://github.com/rhierlmeier/hargassner-monitor/actions/workflows/docker-image.yaml/badge.svg)](https://github.com/rhierlmeier/hargassner-monitor/actions/workflows/docker-image.yaml)

# Hargassner Monitor
Hargassner Monitor is a Go application that reads and parses status records from a Hargassner HSV heating system via a serial connection and publish them via MQTT. The application can be built and run locally or within a Docker container.

## Features

- Reads status records from a serial device
- Parses and displays status records
- Multi-architecture Docker image support (amd64 and arm64)

## Prerequisites

- Go 1.24 or later
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

To pull the Docker image for `amd64` architecture from GitHub Container Registry (GHCR):

```sh
docker pull ghcr.io/rhierlmeier/hargassner-monitor:amd64-latest
```

To pull the Docker image for `arm64` architecture from GHCR:

```sh
docker pull ghcr.io/rhierlmeier/hargassner-monitor:arm64-latest
```

#### Running the Docker Container

To run the Docker container from GHCR:

```sh
docker run --rm \
    -e HARGASSNER_SERIAL_PORT=
    -e HARGASSNER_MQTT_BROKER=tcp://mqtt.local
    ghcr.io/rhierlmeier/hargassner-monitor:latest
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

| **ID**                  | **Name**                   | **Type** | **Unit** |
|-------------------------|----------------------------|----------|----------|
| `rauchgasTemperatur`    | Rauchgas Temperatur        | float    | °C       |
| `boiler1Temperatur`     | Boiler 1 Temperatur        | float    | °C       |
| `aussenTemperaturAktuell` | Aussentemperatur aktuell  | float    | °C       |
| `aussenTemperaturGemittelt` | Aussentemperatur gemittelt | float | °C       |
| `kesselTemperatur`      | Kesseltemperatur           | float    | °C       |
| `kesselSollTemperatur`  | Kesselsolltemperatur       | float    | °C       |
| `meldung`               | Meldung                    | string   |          |
| `saugluftGeblaese`      | Saugluftgebläse            | float    | °C       |
| `primaerLuftGeblaese`   | Primärluftgebläse          | float    | °C       |
| `o2InAbgas`             | O2 in Abgas                | float    | %        |
| `foerderMenge`          | Fördermenge                | float    | %        |
| `stromRaumaustragung`   | Strom Raumaustragung       | float    | A        |
| `stromAscheaustragung`  | Strom Ascheaustragung      | float    | A        |
| `stromEinschub`         | Strom Einschub             | float    | A        |

#### Heizkreis 1

- **ID**: `heizkreis1`
- **Name**: `Heizkreis 1`
- **Type**: `Heizkreis 1`

##### Properties

| **ID**                  | **Name**                   | **Type** | **Unit** |
|-------------------------|----------------------------|----------|----------|
| `vorlaufTemperatur`     | Vorlauftemperatur          | float    | °C       |
| `vorlaufSollTemperatur` | Vorlauf Solltemperatur     | float    | °C       |

#### Heizkreis 2

- **ID**: `heizkreis2`
- **Name**: `Heizkreis 2`
- **Type**: `Heizkreis 2`

##### Properties

| **ID**                  | **Name**                   | **Type** | **Unit** |
|-------------------------|----------------------------|----------|----------|
| `vorlaufTemperatur`     | Vorlauftemperatur          | float    | °C       |
| `vorlaufSollTemperatur` | Vorlauf Solltemperatur     | float    | °C       |

#### Störung

- **ID**: `stoerung`
- **Name**: `Störung`
- **Type**: `Störung`

##### Properties

| **ID** | **Name** | **Type** |
|--------|----------|----------|
| `nr`   | Nummer   | integer  |
| `text` | Text     | string   |
| `active`| Aktiv | boolean |
| `lastChange` | Letzte Änderung | string |
