#!/bin/bash


# Exit immediately if a command exits with a non-zero status
set -e

# Create a new builder instance
docker buildx create --use

# Inspect the builder instance
docker buildx inspect --bootstrap

# Build and push the multi-arch image
docker buildx build --platform linux/amd64,linux/arm64 -t rhierlmeier/hargassner-monitor:latest --push .