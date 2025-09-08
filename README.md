# Necore

Backend of [neco](https://github.com/EntropyGenerator/neco), NMO Ecosystem.

## Installation

Same as a golang project. `go run dev` for development, `go build` for production.

We provide direnv for development environment.

## Database and Backup

Currently we use sqlite3 since it's lightweight and easy to backup. It's stored in `./database.sqlite3`.

We store article contents in `./contents/{id}/`.

Please backup these two files/folders.

## API

Refer to Neco's `API.md`.

Port: 3000