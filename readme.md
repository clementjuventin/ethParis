# Segment

Segment is an NFT indexer on Linea. It is made in Golang.

## Description

This NFT indexer makes it possible to fully index a compatible EVM blockchain from a node using a database.

## Architecture

![alt text](./doc/architecture.png)

- The indexer can connect to any SQL database and any neud indexing an EVM blockchain.
- Highly portable
- Comes with an API

## Getting Started

```
package config

const INFURA_KEY string = "wss://linea-mainnet.infura.io/ws/v3/YOUR_API_KEY"
const POSTGRE_URI string = "postgres://user:password@localhost/dbname?sslmode=disable"
```

This project require a Golang development environment and a postgres database.
With a golang development environment, simply install the dependencies and run `go run .` at the root level of the project.
This will start the indexer that will directly start syncing.

For the api you will need run `go run .` in the `/api` folder.

You have the postman collection for the REST API endpoints.

## Authors

Clément Juventin

## Licence

GPLv3
