package main

import (
	"context"
	"database/sql"
	"log"
	"math/big"
	"os"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/mattn/go-sqlite3"
	"github.com/metachris/eth-go-bindings/erc165"
	"github.com/metachris/eth-go-bindings/erc721"
)

const max_block uint64 = 200
const start_block uint64 = 0

type Data struct {
	Collections      []common.Address
	CollectionsMutex *sync.Mutex
}

type ERC721CollectionStruct struct {
	Address common.Address
	Name    string
	Symbol  string
}

type ERC721TxStruct struct {
	Hash       string
	Tag        string
	FromAddr   common.Address
	ToAddr     common.Address
	Value      string
	TokenId    string
	Collection common.Address
}

var insertionMutex = &sync.Mutex{}

func weiToEther(wei *big.Int) float64 {
	weiFloat := new(big.Float)
	weiFloat.SetString(wei.String())
	etherFloat := new(big.Float).Quo(weiFloat, big.NewFloat(1000000000000000000))
	ether, _ := etherFloat.Float64()
	return ether
}

var EVT_TRANSFER = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

func detectERC721Deployment(tx *types.Transaction, client *ethclient.Client) (common.Address, error) {
	signer := types.LatestSignerForChainID(tx.ChainId())
	sender, err := signer.Sender(tx)
	if err != nil {
		return common.Address{}, err
	}
	contractAddress := crypto.CreateAddress(sender, tx.Nonce())

	contract, err := erc165.NewErc165(contractAddress, client)
	if err != nil {
		return common.Address{}, err
	}
	isERC721, err := contract.SupportsInterface(nil, [4]byte{0x80, 0xac, 0x58, 0xcd})
	if err == nil && isERC721 {
		log.Println("ERC721 contract found !", contractAddress.Hex())
		return contractAddress, nil
	}
	return common.Address{}, err
}

// func getERC721TokenURI(contractAddress common.Address, client *ethclient.Client, tokenId int64) (string, error) {
// 	erc721, err := erc721.NewERC721(contractAddress, client)
// 	if err != nil {
// 		return "", err
// 	}
// 	log.Println("ERC721 contract found !", contractAddress.Hex())

// 	tokenURI, err := erc721.TokenURI(nil, big.NewInt(tokenId))
// 	if err != nil {
// 		return "", err
// 	}
// 	return tokenURI, nil
// }

func eventChecker(tx *types.Transaction, client *ethclient.Client, db *sql.DB) (common.Address, error) {
	// Get tx receipt
	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return common.Address{}, err
	}

	for _, vLog := range receipt.Logs {
		for _, topic := range vLog.Topics {
			if topic == EVT_TRANSFER {
				txTag := "transfer"
				if vLog.Topics[1].Hex()[26:] == common.HexToAddress("0x0").Hex()[2:] {
					txTag = "mint"
				} else if vLog.Topics[0].Hex()[26:] == common.HexToAddress("0x0").Hex()[2:] {
					txTag = "burn"
				}
				err := insertTx(db, ERC721TxStruct{
					Hash:       tx.Hash().Hex(),
					Tag:        txTag,
					FromAddr:   common.HexToAddress("0x"),
					ToAddr:     vLog.Address,
					Value:      tx.Value().String(),
					TokenId:    vLog.Topics[2].Big().String(),
					Collection: vLog.Address,
				})
				if err != nil {
					return common.Address{}, err
				}
			}
		}
	}
	return common.Address{}, nil
}

func blockAnalizer(block *types.Block, client *ethclient.Client, db *sql.DB) {
	for _, tx := range block.Transactions() {
		// if it's a deployment transaction, the to field will be nil
		if tx.To() == nil {
			addr, err := detectERC721Deployment(tx, client)
			if err != nil {
				continue
			}
			if addr != (common.Address{}) {
				// Get the collection Name and Symbol from the contract
				erc721, err := erc721.NewErc721(addr, client)
				if err != nil {
					log.Fatalln(err)
				}
				name, err := erc721.Name(nil)
				if err != nil {
					log.Fatalln(err)
				}
				symbol, err := erc721.Symbol(nil)
				if err != nil {
					log.Fatalln(err)
				}
				// Insert a collection
				inserCollection(db, ERC721CollectionStruct{
					Address: addr,
					Name:    name,
					Symbol:  symbol,
				})
			}
		}
		eventChecker(tx, client, db)
	}
	log.Println("Block", block.Number().Uint64(), "done")
}

func query(client *ethclient.Client, blockNb uint64, db *sql.DB) {
	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(blockNb)))
	if err != nil {
		log.Fatalln(err)
	}
	blockAnalizer(block, client, db)
}

func startClient() (*ethclient.Client, error) {
	endpoint := "wss://linea-mainnet.infura.io/ws/v3/531ec7a2e96642ceac54c241c236c580"
	client, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func syncDatabase(client *ethclient.Client, db *sql.DB) {
	// Current block
	currentBlock, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Current block :", currentBlock)
	syncedBlock, err := selectBlock(db)
	if err != nil {
		log.Fatalln(err)
	}

	var wg sync.WaitGroup
	// Sync the database
	for i := syncedBlock; i < currentBlock; i++ {
		wg.Add(1)
		go func(i uint64) {
			defer wg.Done()
			query(client, i, db)
		}(i)

		if i%25 == 0 {
			wg.Wait()
			log.Println("Synced up to block", i)
			err := updateBlock(db, i)
			if err != nil {
				log.Fatalln(err)
			}
			readDatabase(db)
		}

		if i == max_block-1 {
			currentBlock, err = client.BlockNumber(context.Background())
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
	wg.Wait()

	log.Println("Synced")

	// // Subscribe to new head
	// headers := make(chan *types.Header)
	// sub, err := client.SubscribeNewHead(context.Background(), headers)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// defer sub.Unsubscribe()

	// for {
	// 	select {
	// 	case err := <-sub.Err():
	// 		log.Fatalln(err)
	// 	case header := <-headers:
	// 		log.Println("got header", header.Number)
	// 	}
	// }
}

func main() {
	db, err := startDatabase()
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	client, err := startClient()
	if err != nil {
		log.Fatalln(err)
	}

	// Sync the database
	syncDatabase(client, db)
	readDatabase(db)
}

func selectBlock(db *sql.DB) (block uint64, err error) {
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
	insertState := `INSERT INTO State(block) VALUES (?)`
	_, err = db.Exec(insertState, start_block)
	if err != nil {
		return 0, err
	}
	return start_block, nil
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

func updateBlock(db *sql.DB, block uint64) (err error) {
	update := `UPDATE State SET block = ?`
	log.Println("Updating block to", block)
	return exec(db, update, block)
}

func inserCollection(db *sql.DB, toInsert ERC721CollectionStruct) (err error) {
	// Insert a collection
	insertCollection := `INSERT INTO ERC721Collection(address, name, symbol) VALUES (?, ?, ?)`
	log.Println("Inserting collection :", toInsert.Address.Hex())
	return exec(db, insertCollection, toInsert.Address.Hex(), toInsert.Name, toInsert.Symbol)
}

func insertTx(db *sql.DB, toInsert ERC721TxStruct) (err error) {
	// Insert a tx
	insertTx := `INSERT INTO ERC721Tx(hash, tag, fromAddr, toAddr, value, tokenId, collection) VALUES (?, ?, ?, ?, ?, ?, ?)`
	log.Println("Inserting tx :", toInsert.Hash)

	log.Println("Command :", insertTx)
	log.Println("Args :", toInsert.Hash, toInsert.Tag, toInsert.FromAddr.Hex(), toInsert.ToAddr.Hex(), toInsert.Value, toInsert.TokenId, toInsert.Collection.Hex())

	insertionMutex.Lock()
	err = exec(db, insertTx, toInsert.Hash, toInsert.Tag, toInsert.FromAddr.Hex(), toInsert.ToAddr.Hex(), toInsert.Value, toInsert.TokenId, toInsert.Collection.Hex())
	insertionMutex.Unlock()
	return err
}

// Database
func startDatabase() (database *sql.DB, e error) {
	const file string = "dtb.db"

	// Reset the database
	err := os.Remove(file)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}

	const ERC721Collection string = `CREATE TABLE IF NOT EXISTS ERC721Collection (
		address TEXT NOT NULL PRIMARY KEY,
		name TEXT,
		symbol TEXT
	);`

	const ERC721Tx string = `CREATE TABLE IF NOT EXISTS ERC721Tx (
		hash TEXT NOT NULL PRIMARY KEY,
		tag TEXT,
		fromAddr TEXT,
		toAddr TEXT,
		value TEXT,
		tokenId TEXT,
		collection TEXT,
		FOREIGN KEY(collection) REFERENCES ERC721Collection(address)
	);`

	const State string = `CREATE TABLE IF NOT EXISTS State (
		block INTEGER NOT NULL PRIMARY KEY
	);`

	for _, create := range []string{ERC721Collection, ERC721Tx, State} {
		_, err = db.Exec(create)
		if err != nil {
			return nil, err
		}
	}

	// If state already exists, return
	block, err := selectBlock(db)
	if err == nil {
		log.Println("State already exists, block :", block)
		return db, nil
	}

	insertState := `INSERT INTO State(block) VALUES (?)`
	_, err = db.Exec(insertState, start_block)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func readDatabase(db *sql.DB) {
	// Query all the collections
	rows, err := db.Query("SELECT address FROM ERC721Collection")
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
