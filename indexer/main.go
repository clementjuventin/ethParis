package main

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

func checkSupportInterface(client *w3.Client, address common.Address, interfaceId string) bool {

	funcSupportsInterface := w3.MustNewFunc("supportsInterface(bytes4 interfaceId)", "bool")

	var resp bool

	err := client.Call(
		eth.CallFunc(funcSupportsInterface, address, w3.B(interfaceId)).Returns(&resp),
	)

	if err != nil {
		log.Fatalln("err w3 :", err)
	}
	return resp
}

func main() {

	endpoint := "wss://linea-mainnet.infura.io/ws/v3/ed46365c230a4510a18eed8799cf4f01"

	client, err := ethclient.Dial(endpoint)
	if err != nil {
		log.Fatalln(err)
	}

	client2 := w3.MustDial(endpoint)

	log.Println("passed both conn")

	/*
		newHeads := make(chan *types.Header)
		sub, err := client.SubscribeNewHead(context.Background(), newHeads)
		if err != nil {
			log.Fatalln(err)
		}

			for {
				select {
				case err := <-sub.Err():
					log.Fatalln("connection ended : ", err)

				case head := <-newHeads:
					log.Println("==========")
					log.Println("New header received :", head.Hash())

					block, err := client.BlockByHash(context.Background(), head.Hash())
					if err != nil {
						log.Fatalln(err)
					}
					log.Println("////////")
					log.Println("Transactions for block :")
					for pos, tx := range block.Transactions() {
						log.Println("tx pos / hash :", pos, "|", tx.Hash())
						log.Println("to :", tx.To())
						log.Println("data :", common.Bytes2Hex(tx.Data()))
						if tx.To() == nil {
							log.Println("CONTRACT DEPLOYEMENT")
							resp := checkSupportInterface(client2, tx.To(), "0x80ac58cd")
							log.Println("implements interface :", resp)
						}
					}
				}


			}
	*/

	latestBlock, err := client.BlockNumber(context.Background())

	for i := 0; i < int(latestBlock); i++ {

		block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(i)))
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("////////")
		log.Println("Transactions for block :")
		for pos, tx := range block.Transactions() {
			if tx.To() == nil {
				log.Println("tx pos / hash :", pos, "|", tx.Hash())
				log.Println("to :", tx.To())
				log.Println("data :", common.Bytes2Hex(tx.Data()))
				log.Println("CONTRACT DEPLOYEMENT")
				signer := types.LatestSignerForChainID(tx.ChainId())
				sender, err := signer.Sender(tx)
				if err != nil {
					log.Fatalln(err)
				}
				contractAddr := crypto.CreateAddress(sender, tx.Nonce())
				log.Println("contract address : ", contractAddr)
				log.Println("tx sender : ", sender)
				resp := checkSupportInterface(client2, contractAddr, "0x80ac58cd")
				log.Println("implements interface :", resp)
			}
		}
	}
}
