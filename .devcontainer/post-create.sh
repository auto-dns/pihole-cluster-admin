#!/usr/bin/env bash
set -e

# If .env is missing, copy example.env to .env
if [ ! -f .devcontainer/.env ]; then
  cp .devcontainer/example.env .devcontainer/.env
fi

# Install frontend dependencies
npm install --prefix frontend
