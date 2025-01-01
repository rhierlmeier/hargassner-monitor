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
git clone https://github.com/yourusername/hargassner-monitor.git
cd hargassner-monitor