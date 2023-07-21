package main

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/metachris/eth-go-bindings/erc165"
)

func main() {
	endpoint := "wss://linea-mainnet.infura.io/ws/v3/11135f07f5c84261a8887926742776c6"
	client, err := ethclient.Dial(endpoint)
	if err != nil {
		log.Fatalln(err)
	}
	// w3client := w3.MustDial(endpoint)

	blockSync := big.NewInt(0)
	// Syncing
	for {
		block, err := client.BlockByNumber(context.Background(), blockSync)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("block number :", block.Number())
		transactions := block.Transactions()

		for _, tx := range transactions {
			log.Println("----------------")
			// if it's a deployment transaction, the to field will be nil
			if tx.To() == nil {
				signer := types.LatestSignerForChainID(tx.ChainId())
				sender, err := signer.Sender(tx)
				if err != nil {
					log.Fatalln(err)
				}
				contractAddress := crypto.CreateAddress(sender, tx.Nonce())

				contract, err := erc165.NewErc165(contractAddress, client)
				if err != nil {
					log.Fatalln(err)
				}
				isERC721, err := contract.SupportsInterface(nil, [4]byte{0x80, 0xac, 0x58, 0xcd})
				if err != nil {
					// log.Fatalln(err)
				}

				if isERC721 {
					log.Println("ERC721 contract found !")
					log.Println("hash :", tx.Hash())
					log.Println("sender :", sender.Hex())
					log.Println("contract address :", contractAddress.Hex())
				}
			}
		}

		blockNumber, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Fatalln(err)
		}

		if blockNumber == block.Number().Uint64() {
			break
		}
		blockSync.Add(blockSync, big.NewInt(1))
	}
}
