package main

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/metachris/eth-go-bindings/erc165"
)

var EVT_TRANSFER = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
var EVT_MINT = crypto.Keccak256Hash([]byte("Mint(address,uint256)"))
var EVT_BURN = crypto.Keccak256Hash([]byte("Burn(address,uint256)"))

var EVENTS_SIG = []common.Hash{
	EVT_TRANSFER,
	EVT_MINT,
	EVT_BURN,
}
var NIL_ADDR = common.HexToAddress("0x0")

func weiToEther(wei *big.Int) *big.Float {
	ether := new(big.Float)
	ether.SetString(wei.String())
	ether = ether.Quo(ether, big.NewFloat(1000000000000000000))
	return ether
}

func detectERC721Deployment(tx *types.Transaction, client *ethclient.Client) (common.Address, error) {
	signer := types.LatestSignerForChainID(tx.ChainId())
	sender, err := signer.Sender(tx)
	if err != nil {
		return NIL_ADDR, err
	}
	contractAddress := crypto.CreateAddress(sender, tx.Nonce())

	contract, err := erc165.NewErc165(contractAddress, client)
	if err != nil {
		return NIL_ADDR, err
	}
	isERC721, err := contract.SupportsInterface(nil, [4]byte{0x80, 0xac, 0x58, 0xcd})
	if err == nil && isERC721 {
		log.Println("ERC721 contract found !", contractAddress.Hex())
		return contractAddress, nil
	}
	return NIL_ADDR, err
}

func eventChecker(tx *types.Transaction, client *ethclient.Client) (common.Address, error) {
	// Get tx receipt
	receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return NIL_ADDR, err
	}

	for _, vLog := range receipt.Logs {
		for _, topic := range vLog.Topics {

			if topic == EVT_TRANSFER {
				log.Println("Event found ! -> Transfer")
				if vLog.Topics[1].Hex()[26:] == common.HexToAddress("0x0").Hex()[2:] {
					log.Println("Mint @@@@@@@@@@")
					log.Println("Token id :", vLog.Topics[3].Big().Uint64())
					// Wei to ether
					txValue := weiToEther(tx.Value())

					log.Println("mint price :", txValue, "ETH")
					log.Println("Tx :", tx.Hash().Hex())
				}
			}
			if topic == EVT_MINT {
				log.Println("Event found ! -> Mint")
				log.Println("Tx :", tx.Hash().Hex())
			}
			if topic == EVT_BURN {
				log.Println("Event found ! -> Burn")
			}
		}
	}
	return NIL_ADDR, nil
}

func blockAnalizer(block *types.Block, client *ethclient.Client) {
	transactions := block.Transactions()
	for _, tx := range transactions {
		// if it's a deployment transaction, the to field will be nil
		if tx.To() == nil {
			_, err := detectERC721Deployment(tx, client)
			if err != nil {
				log.Fatalln(err)
			}
		}

		_, err := eventChecker(tx, client)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func main() {
	endpoint := "wss://linea-mainnet.infura.io/ws/v3/11135f07f5c84261a8887926742776c6"
	client, err := ethclient.Dial(endpoint)
	if err != nil {
		log.Fatalln(err)
	}
	// w3client := w3.MustDial(endpoint)

	blockSync := big.NewInt(4541)
	// Syncing
	for {
		block, err := client.BlockByNumber(context.Background(), blockSync)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("block number :", block.Number())
		blockAnalizer(block, client)

		blockNumber, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("block number :", block.Number())

		if blockNumber == block.Number().Uint64() {
			return
		}
		blockSync.Add(blockSync, big.NewInt(1))
	}
}
