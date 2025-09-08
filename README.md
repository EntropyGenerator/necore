# Necore

Backend of [neco](https://github.com/EntropyGenerator/neco), NMO Ecosystem.

## Installation

Same as a golang project. `go run dev` for development, `go build` for production.

We provide direnv for development environment.

## Database

Currently we use sqlite3 since it's lightweight and easy to backup.

We store article contents in `./contents/{id}/`.

## API

Refer to Neco's `API.md`.

Port: 3000