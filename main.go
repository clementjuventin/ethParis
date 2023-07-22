package main

import (
	"context"
	"database/sql"
	"log"
	"math/big"
	"sync"

	"workspace/customTypes"

	"workspace/database"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/lib/pq"
	"github.com/metachris/eth-go-bindings/erc165"
	"github.com/metachris/eth-go-bindings/erc721"
)

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
				err := database.InsertTx(db, customTypes.ERC721TxStruct{
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
				database.InserCollection(db, customTypes.ERC721CollectionStruct{
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
	endpoint := "wss://linea-mainnet.infura.io/ws/v3/d34b021a1e8e4219a919faa2265b62e3"
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
	syncedBlock, err := database.SelectBlock(db)
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

		if i%1000 == 0 {
			wg.Wait()
			log.Println("Synced up to block", i)
			err := database.UpdateBlock(db, i)
			if err != nil {
				log.Fatalln(err)
			}
		}

		if i == currentBlock-1 {
			currentBlock, err = client.BlockNumber(context.Background())
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
	wg.Wait()

	log.Println("Synced")
}

func main() {
	db, err := database.StartDatabase()
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
}
