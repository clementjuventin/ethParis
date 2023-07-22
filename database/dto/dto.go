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
CREATE TABLE IF NOT EXISTS ERC721Tx (
	id SERIAL PRIMARY KEY,
	mint_timestamp text,
	mint_block_number text,
	mint_hash text,
	uri text,
	token_id text,
	collection text,
	owner text,
	FOREIGN KEY (collection) REFERENCES ERC721Collection(contract_address)
);
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
	collection text,
	FOREIGN KEY (collection) REFERENCES ERC721Collection(contract_address)
);
`

const STATE_TABLE string = `
CREATE TABLE IF NOT EXISTS State (
	block INTEGER NOT NULL PRIMARY KEY
);
`
