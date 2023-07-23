package dto

const ERC721_COLLECTION_TABLE string = `
CREATE TABLE IF NOT EXISTS ERC721Collection (
	deploy_timestamp text NOT NULL,
	block_number text NOT NULL,
	deploy_hash text NOT NULL,
	contract_address text NOT NULL PRIMARY KEY,
	contract_name text,
	contract_symbol text
);
`
const ERC721_TABLE string = `
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
`

const ERC721_TX_TABLE string = `
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
`

const STATE_TABLE string = `
CREATE TABLE IF NOT EXISTS State (
	block INTEGER NOT NULL PRIMARY KEY
);
`

const DROP_TABLES string = `
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
`

const DELETE_ROWS string = `
DELETE FROM ERC721Tx;
DELETE FROM ERC721Collection;
DELETE FROM ERC721;
DELETE FROM State;
`
