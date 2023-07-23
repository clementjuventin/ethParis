package main

import (
	"context"
	"database/sql"
	"log"
	"math/big"
	"sync"

	"workspace/config"
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
		return contractAddress, nil
	}
	return common.Address{}, err
}

func eventChecker(tx *types.Transaction, block *types.Block, client *ethclient.Client, db *sql.DB) (common.Address, error) {
	// Get tx receipt
	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return common.Address{}, err
	}

	for _, vLog := range receipt.Logs {
		topic := vLog.Topics[0]
		if topic == EVT_TRANSFER {
			offset := 0
			// Check if the len is two (=> it's not a transfer event)
			if len(vLog.Topics) <= 3 {
				continue
			}

			// Check if there is a fourth argument in the topic
			if len(vLog.Topics) == 5 {
				offset = 1
			}

			txTag := "transfer"
			if vLog.Topics[1+offset].Hex()[26:] == common.HexToAddress("0x0").Hex()[2:] {
				txTag = "mint"
			} else if vLog.Topics[2+offset].Hex()[26:] == common.HexToAddress("0x0").Hex()[2:] {
				txTag = "burn"
			}

			tokenId := vLog.Topics[3+offset].Big()

			tx := customTypes.ERC721TxStruct{
				Timestamp:   block.Time(),
				BlockNumber: block.Number().Uint64(),
				TxHash:      tx.Hash().Hex(),
				Tag:         txTag,
				FromAddr:    common.HexToAddress(vLog.Topics[1+offset].Hex()[26:]),
				ToAddr:      common.HexToAddress(vLog.Topics[2+offset].Hex()[26:]),
				Value:       tx.Value().String(),
				TokenId:     tokenId.String(),
				Collection:  common.HexToAddress(vLog.Address.Hex()),
			}
			err := database.InsertTx(db, tx)

			if txTag == "transfer" {
				database.UpdateOwner(db, tx)
			}
			if txTag == "mint" {
				erc721, err := erc721.NewErc721(tx.Collection, client)
				if err != nil {
					log.Fatalln(err)
				}
				uri, err := erc721.TokenURI(nil, tokenId)
				if err != nil {
					uri = ""
				}

				nft := customTypes.ERC721Struct{
					MintTimestamp:   block.Time(),
					MintBlockNumber: block.Number().Uint64(),
					MintTxHash:      tx.TxHash,
					URI:             uri,
					TokenId:         tx.TokenId,
					Collection:      tx.Collection,
					Owner:           tx.ToAddr,
				}

				database.InsertMint(db, nft)
			}

			if err != nil {
				return common.Address{}, err
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
					ContractAddress:   addr,
					ContractName:      name,
					ContractSymbol:    symbol,
					DeployTimestamp:   block.Time(),
					DeployBlockNumber: block.Number().Uint64(),
					DeployTxHash:      tx.Hash().Hex(),
				})
			}
		}
		eventChecker(tx, block, client, db)
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
	endpoint := config.INFURA_KEY
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

	// tx 0x6e8f8ee72ef5d07c91d78876e916be3b017c6f28358dd0543336019f75f9024f
	// for i := 500; i < 600; i++ {
	// 	bi := big.NewInt(int64(i))
	// 	block, _ := client.BlockByNumber(context.Background(), bi)
	// 	blockAnalizer(block, client, db)
	// }
	// // tx, _, _ := client.TransactionByHash(context.Background(), common.HexToHash("0xede8f5012c5115534103b9ed25cc1cdfaeb41e661a7a4813a4e6e721922b616d"))
	// // block, _ := client.BlockByNumber(context.Background(), big.NewInt(4228))
	// // eventChecker(tx, block, client, db)

	// log.Fatal("Done")

	// Sync the database
	syncDatabase(client, db)

	// Listen for new blocks
	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatalln(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatalln(err)
		case header := <-headers:
			go func() {
				block, err := client.BlockByHash(context.Background(), header.Hash())
				if err != nil {
					log.Fatalln(err)
				}
				blockAnalizer(block, client, db)
			}()
		}
	}
}
