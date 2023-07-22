package database

import (
	"database/sql"
	"log"

	"workspace/config"
	"workspace/customTypes"
	"workspace/database/dto"
)

// Log the whole db
func readDatabase(db *sql.DB) {
	// Query all the collections
	rows, err := db.Query("SELECT contract_address FROM ERC721Collection")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	log.Println("Collections found :")
	for rows.Next() {
		var address string
		err = rows.Scan(&address)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("|", address)
	}

	// Query all the tx
	rows, err = db.Query("SELECT hash FROM ERC721Tx")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	log.Println("Tx found :")
	for rows.Next() {
		var hash string
		err = rows.Scan(&hash)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("|", hash)
	}
}

// ///////////////////////////////////// QUERIES ///////////////////////////////////////
// Insert a collection
func InserCollection(db *sql.DB, toInsert customTypes.ERC721CollectionStruct) (err error) {
	// Insert a collection
	insertCollection := `INSERT INTO ERC721Collection(deploy_timestamp, block_number, deploy_hash, contract_address, contract_name, contract_symbol) VALUES ($1, $2, $3, $4, $5, $6)`
	err = exec(db, insertCollection, toInsert.DeployTimestamp, toInsert.DeployBlockNumber, toInsert.DeployTxHash, toInsert.ContractAddress.Hex(), toInsert.ContractName, toInsert.ContractSymbol)
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
	log.Println("Updating block to", block)
	err = exec(db, update, block)
	return err
}

// Insert a tx
func InsertTx(db *sql.DB, toInsert customTypes.ERC721TxStruct) (err error) {
	// Insert a tx
	insertTx := `INSERT INTO ERC721Tx(timestamp, block_number, hash, tag, from_addr, to_addr, value, token_id, collection) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	err = exec(db, insertTx, toInsert.Timestamp, toInsert.BlockNumber, toInsert.TxHash, toInsert.Tag, toInsert.FromAddr.Hex(), toInsert.ToAddr.Hex(), toInsert.Value, toInsert.TokenId, toInsert.Collection.Hex())
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

	// // Drop tables
	// _, err = db.Exec("DROP TABLE IF EXISTS ERC721Tx")
	// if err != nil {
	// 	return nil, err
	// }
	// _, err = db.Exec("DROP TABLE IF EXISTS ERC721Collection")
	// if err != nil {
	// 	return nil, err
	// }
	// _, err = db.Exec("DROP TABLE IF EXISTS State")
	// if err != nil {
	// 	return nil, err
	// }

	for _, create := range []string{dto.ERC721_COLLECTION_TABLE, dto.ERC721_TABLE, dto.ERC721_TX_TABLE} {
		_, err = db.Exec(create)
		if err != nil {
			return nil, err
		}
	}

	// // Clean the database
	// _, err = db.Exec("DELETE FROM ERC721Tx")
	// if err != nil {
	// 	return nil, err
	// }
	// _, err = db.Exec("DELETE FROM ERC721Collection")
	// if err != nil {
	// 	return nil, err
	// }
	// _, err = db.Exec("DELETE FROM State")
	// if err != nil {
	// 	return nil, err
	// }

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
