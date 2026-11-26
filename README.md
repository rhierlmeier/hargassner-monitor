
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
    -e HARGASSNER_SERIAL_DEVICE=/dev/ttyUSB0 \
    -e HARGASSNER_MQTT_BROKER=tcp://mqtt.local \
    ghcr.io/rhierlmeier/hargassner-monitor:latest
```

Replace `/dev/ttyUSB0` with the appropriate serial device on your system.

# Environment Variables

The application uses the following environment variables:

- `HARGASSNER_SERIAL_DEVICE`: Specifies the serial device to which the Hargassner heating system is connected. Default is `/dev/ttyUSB0`.
- `HARGASSNER_MQTT_BROKER`: Specifies the MQTT broker URL. Default is `tcp://localhost:1883`.
- `HARGASSNER_MQTT_CLIENT_ID`: Specifies the MQTT client ID. Default is `hargassner-monitor`.
- `HARGASSNER_MQTT_USER`: Specifies the username for MQTT broker authentication. Default is empty.
- `HARGASSNER_MQTT_PASSWORD`: Specifies the password for MQTT broker authentication. Default is empty.
- `HARGASSNER_MONITOR_PORT`: Port where the HTTP server first status request is listing
- `HARGASSNER_MAX_DATA_AGE`: Maximum time without data from the serial port before `/readiness` reports not ready, as a Go duration (e.g. `5m`). Default is `5m`.
- `HARGASSNER_LOG_RAW`: `true` logs every line read from the serial port, e.g. to identify fields of the `pm` record that are not parsed yet. Default is empty: no logging.
- `HARGASSNER_HA_DISCOVERY`: Publishes Home Assistant MQTT discovery configs, see below. `all` for every property, or a comma-separated list of `<node>/<property>` (e.g. `prozesswerte/aussenTemperaturAktuell`). Default is empty: no discovery.
- `HARGASSNER_HA_DISCOVERY_PREFIX`: Discovery prefix of Home Assistant. Default is `homeassistant`.
- `HARGASSNER_HA_LANGUAGE`: Language of the entity names in Home Assistant, `en` or `de`. Default is `en`.
- `TZ`: Time zone of the controller (e.g. `Europe/Berlin`). Controller records carry only the time of day; `stoerung/lastActive` is built from it in this time zone. Default is UTC.

## Readiness

`GET /readiness` returns 200 if the MQTT client is connected and a line was
read from the serial port within `HARGASSNER_MAX_DATA_AGE`. Otherwise it
returns 503 with the reasons. A silent serial port (controller switched off,
cable loose while the USB adapter is still present) causes no read error, so
it is only detected this way. After a start, the monitor allows
`HARGASSNER_MAX_DATA_AGE` for the first data. Use the endpoint as a readiness
probe; a restart does not fix a silent controller.


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

#### Kessel

- **ID**: `kessel`
- **Name**: `Kessel`
- **Type**: `Kessel`

##### Properties

| **ID**                        | **Name**                    | **Type** | **Unit** |
|-------------------------------|-----------------------------|----------|----------|
| `DauerLetzteZuendung`         | Dauer letzte Zündung        | integer  | s        |
| `DauerLetzterLeistungsbrand` | Dauer letzter Leistungsbrand | integer  | s        |
| `AnzahlZuendungen`            | Anzahl Zündungen            | integer  |          |

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
| `lastActive` | Letzte Aktivität | string |

`lastActive` is an RFC 3339 timestamp, e.g. `2026-11-15T08:29:10+01:00`
(before 1.3.0 only the time of day, `08:29:10`).

## Home Assistant

Home Assistant has no Homie integration. With `HARGASSNER_HA_DISCOVERY` set,
the monitor publishes a retained [MQTT discovery](https://www.home-assistant.io/integrations/mqtt/#mqtt-discovery)
config per property on `<prefix>/<component>/hargassner/<node>_<property>/config`
whenever it connects to the broker. Home Assistant then creates the entities
itself, grouped under one device `Hargassner Heizung`. The configs only point
at the Homie topics, which stay unchanged.

- Numbers become `sensor` entities with unit, `device_class` (from the unit)
  and `state_class: measurement`; booleans become `binary_sensor` with
  `payload_on: "true"`, `payload_off: "false"`; strings become `sensor`.
- An entity is available while `homie/hargassner/$state` is `ready`.
- Properties that are not selected get an empty retained config, which
  removes an entity discovered earlier. A property removed from the code
  leaves its config on the broker; clear it by hand with an empty retained
  message.

### Migrating from hand-written MQTT entities

The 23 entities of the former hand-written Home Assistant config keep their
`unique_id` (see `haOverrides` in `ha_discovery.go`), so Home Assistant keeps
their entity IDs and history. New properties get `hargassner_<node>_<property>`.
Home Assistant rejects a discovered entity while a YAML entity with the same
`unique_id` exists, so migrate like this:

1. Select one property with `HARGASSNER_HA_DISCOVERY`, remove its entity from
   the YAML config and restart Home Assistant. Check entity ID, history and
   dashboards.
2. Set `HARGASSNER_HA_DISCOVERY=all` and remove the YAML entities in the same
   step.

