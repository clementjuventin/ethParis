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

## API

The endpoints are :

```
	/nft/history/:collection/:tokenId   // Get the NFT transaction history
	/nft/:collection/:tokenId           // Get NFT data (URI, Owner, etc..)

	/collection/:addr                   // Get collection NFTs
	/collection/history/:addr           // Get collection transaction history
	/collection/stats/:addr             // Get some stats on the collection

	/address/history/:addr              // Get all the ERC721 transactions of an address
	/address/:addr                      // Get the NFTs owned by an address
```

Check the Postman collection.

## Database script

```
DROP INDEX IF EXISTS ERC721_collection_idx;
DROP INDEX IF EXISTS ERC721_owner_idx;
DROP INDEX IF EXISTS ERC721Tx_collection_idx;
DROP INDEX IF EXISTS ERC721Tx_from_idx;
DROP INDEX IF EXISTS ERC721Tx_to_idx;
DROP INDEX IF EXISTS ERC721Tx_token_id_and_collection_idx;

DROP TABLE IF EXISTS ERC721Tx;
DROP TABLE IF EXISTS ERC721;
DROP TABLE IF EXISTS ERC721Collection;
DROP TABLE IF EXISTS State;

CREATE TABLE IF NOT EXISTS ERC721Collection (
	deploy_timestamp text NOT NULL,
	block_number text NOT NULL,
	deploy_hash text NOT NULL,
	contract_address text NOT NULL PRIMARY KEY,
	contract_name text,
	contract_symbol text
);
CREATE TABLE IF NOT EXISTS ERC721 (
	mint_timestamp text,
	mint_block_number text,
	mint_hash text,
	uri text,
	token_id text,
	collection text,
	owner text,
	PRIMARY KEY (token_id, collection)
);

CREATE INDEX IF NOT EXISTS ERC721_collection_idx ON ERC721(collection);
CREATE INDEX IF NOT EXISTS ERC721_owner_idx ON ERC721(owner);

CREATE TABLE IF NOT EXISTS ERC721Tx (
	id SERIAL PRIMARY KEY,
	timestamp text,
	block_number text,
	hash text,
	tag text,
	from_addr text,
	to_addr text,
	value text,
	token_id text,
	collection text
);

CREATE INDEX IF NOT EXISTS ERC721Tx_collection_idx ON ERC721Tx(collection);
CREATE INDEX IF NOT EXISTS ERC721Tx_from_idx ON ERC721Tx(from_addr);
CREATE INDEX IF NOT EXISTS ERC721Tx_to_idx ON ERC721Tx(to_addr);
CREATE INDEX IF NOT EXISTS ERC721Tx_token_id_and_collection_idx ON ERC721Tx(token_id, collection);

CREATE TABLE IF NOT EXISTS State (
	block INTEGER NOT NULL PRIMARY KEY
);
```

## Authors

Clément Juventin

## Licence

GPLv3
