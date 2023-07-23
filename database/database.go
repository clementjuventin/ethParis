package database

import (
	"database/sql"
	"log"
	"strings"

	"workspace/config"
	"workspace/customTypes"
	"workspace/database/dto"
)

// ///////////////////////////////////// QUERIES ///////////////////////////////////////
// Insert a collection
func InserCollection(db *sql.DB, toInsert customTypes.ERC721CollectionStruct) (err error) {
	// Insert a collection
	insertCollection := `INSERT INTO ERC721Collection(deploy_timestamp, block_number, deploy_hash, contract_address, contract_name, contract_symbol) VALUES ($1, $2, $3, $4, $5, $6)`
	err = exec(db, insertCollection, toInsert.DeployTimestamp, toInsert.DeployBlockNumber, toInsert.DeployTxHash, strings.ToLower(toInsert.ContractAddress.Hex()), toInsert.ContractName, toInsert.ContractSymbol)
	if err != nil {
		log.Println("Error :", err)
	}
	return err
}

// Get the last block of the db
func SelectBlock(db *sql.DB) (block uint64, err error) {
	// Query the state
	rows, err := db.Query("SELECT block FROM State")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&block)
		if err != nil {
			return 0, err
		}
		return block, nil
	}

	// Insert a block
	insertState := `INSERT INTO State(block) VALUES ($1)`
	_, err = db.Exec(insertState, config.START_BLOCK)
	if err != nil {
		return config.START_BLOCK, err
	}
	return config.START_BLOCK, nil
}

// Update the last block
func UpdateBlock(db *sql.DB, block uint64) (err error) {
	update := `UPDATE State SET block = $1`
	err = exec(db, update, block)
	return err
}

// Insert a tx
func InsertTx(db *sql.DB, toInsert customTypes.ERC721TxStruct) (err error) {
	// Insert a tx
	insertTx := `INSERT INTO ERC721Tx(timestamp, block_number, hash, tag, from_addr, to_addr, value, token_id, collection) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	err = exec(db, insertTx, toInsert.Timestamp, toInsert.BlockNumber, toInsert.TxHash, toInsert.Tag, strings.ToLower(toInsert.FromAddr.Hex()), strings.ToLower(toInsert.ToAddr.Hex()), toInsert.Value, toInsert.TokenId, strings.ToLower(toInsert.Collection.Hex()))
	if err != nil {
		log.Println("Error :", err)
	}
	return err
}

// Update Owner
func UpdateOwner(db *sql.DB, toUpdate customTypes.ERC721TxStruct) (err error) {
	// Update owner
	updateOwner := `UPDATE ERC721 SET owner = $1 WHERE token_id = $2 AND collection = $3`
	err = exec(db, updateOwner, strings.ToLower(toUpdate.ToAddr.Hex()), toUpdate.TokenId, strings.ToLower(toUpdate.Collection.Hex()))
	if err != nil {
		log.Println("Error :", err)
	}
	return err
}

// Insert a mint
func InsertMint(db *sql.DB, toInsert customTypes.ERC721Struct) (err error) {
	// Insert a mint
	insertMint := `INSERT INTO ERC721(mint_timestamp, mint_block_number, mint_hash, uri, token_id, collection, owner) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	err = exec(db, insertMint, toInsert.MintTimestamp, toInsert.MintBlockNumber, toInsert.MintTxHash, toInsert.URI, toInsert.TokenId, strings.ToLower(toInsert.Collection.Hex()), strings.ToLower(toInsert.Owner.Hex()))
	if err != nil {
		log.Println("Error :", err)
	}
	return err
}

// ///////////////////////////////////// UTILS ///////////////////////////////////////
// Start db
func StartDatabase() (database *sql.DB, e error) {
	db, err := sql.Open("postgres", config.POSTGRE_URI)
	if err != nil {
		return nil, err
	}

	// Drop indexes
	_, err = db.Exec("DROP INDEX IF EXISTS ERC721_collection_idx")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("DROP INDEX IF EXISTS ERC721_owner_idx")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("DROP INDEX IF EXISTS ERC721Tx_collection_idx")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("DROP INDEX IF EXISTS ERC721Tx_from_idx")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("DROP INDEX IF EXISTS ERC721Tx_to_idx")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("DROP INDEX IF EXISTS ERC721Tx_token_id_and_collection_idx")
	if err != nil {
		return nil, err
	}
	// Drop tables
	_, err = db.Exec("DROP TABLE IF EXISTS ERC721Tx")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("DROP TABLE IF EXISTS ERC721")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("DROP TABLE IF EXISTS ERC721Collection")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("DROP TABLE IF EXISTS State")
	if err != nil {
		return nil, err
	}

	for _, create := range []string{dto.ERC721_COLLECTION_TABLE, dto.ERC721_TABLE, dto.ERC721_TX_TABLE, dto.STATE_TABLE} {
		_, err = db.Exec(create)
		if err != nil {
			return nil, err
		}
	}

	// Clean the database
	_, err = db.Exec("DELETE FROM ERC721Tx")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("DELETE FROM ERC721Collection")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("DELETE FROM State")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("DELETE FROM ERC721")
	if err != nil {
		return nil, err
	}

	// If state already exists, return
	block, err := SelectBlock(db)
	if err == nil {
		log.Println("State already exists, block :", block)
		return db, nil
	}

	insertState := `INSERT INTO State (block) VALUES ($1)`
	_, err = db.Exec(insertState, config.START_BLOCK)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func exec(db *sql.DB, query string, args ...any) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
