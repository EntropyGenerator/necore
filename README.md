# Necore

High performance backend of [neco](https://github.com/EntropyGenerator/neco)(NMO Ecosystem) based on gofiber.

## Development

`direnv` can be used for development environment.

## Database and Backup

- Database: `./data/*.sqlite3`.

- Contents: `./contents/{id}/*`.

Please backup these files/folders when necessary.

## API

Please refer to Neco's [`API.md`](https://github.com/EntropyGenerator/neco/blob/main/API.md).

## Config

`.env` is used for configuration.

- `PORT`: Port of the server, default is `3000`.
- `SECRET`: Secret key for JWT.