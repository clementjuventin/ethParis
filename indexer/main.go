package main

import (
	"context"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/metachris/eth-go-bindings/erc165"
)

/*
var EVT_TRANSFER = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
var EVT_MINT = crypto.Keccak256Hash([]byte("Mint(address,uint256)"))
var EVT_BURN = crypto.Keccak256Hash([]byte("Burn(address,uint256)"))

	var EVENTS_SIG = []common.Hash{
		EVT_TRANSFER,
		EVT_MINT,
		EVT_BURN,
	}

	func weiToEther(wei *big.Int) *big.Float {
		ether := new(big.Float)
		ether.SetString(wei.String())
		ether = ether.Quo(ether, big.NewFloat(1000000000000000000))
		return ether
	}

	func eventChecker(tx *types.Transaction, client *ethclient.Client) (common.Address, error) {
		// Get tx receipt
		receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			return common.Address{}, err
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
		return common.Address{}, nil
	}
*/
type Data struct {
	Collections      []common.Address
	CollectionsMutex *sync.Mutex
}

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

func blockAnalizer(block *types.Block, client *ethclient.Client, data *Data) {
	for _, tx := range block.Transactions() {
		// if it's a deployment transaction, the to field will be nil
		if tx.To() == nil {
			addr, err := detectERC721Deployment(tx, client)
			if err != nil {
				continue
			}
			if addr != (common.Address{}) {
				data.CollectionsMutex.Lock()
				data.Collections = append(data.Collections, addr)
				data.CollectionsMutex.Unlock()
			}
		}
	}
}

func query(client *ethclient.Client, blockNb int, data *Data) {
	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(blockNb)))
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("got block :", block.Number())
	blockAnalizer(block, client, data)
}

func manager(client *ethclient.Client, blockHeight int) *Data {

	data := new(Data)
	data.CollectionsMutex = new(sync.Mutex)

	for i := 0; i < blockHeight; i++ {
		if (i+1)%75== 0 {
			time.Sleep(time.Second * 1)
			log.Println("batch completed at :", i)
		}
		go query(client, i, data)
	}

	return data
}

func main() {
	endpoint := "wss://linea-mainnet.infura.io/ws/v3/531ec7a2e96642ceac54c241c236c580"
	client, err := ethclient.Dial(endpoint)
	if err != nil {
		log.Fatalln(err)
	}
	contract, err := erc165.NewErc165(common.HexToAddress("0xb62c414abf83c0107db84f8de1c88631c05a8d7b"), client)
	if err != nil {
		log.Fatalln(err)
	}
	isERC721, err := contract.SupportsInterface(nil, [4]byte{0x80, 0xac, 0x58, 0xcd})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("test punks :", isERC721)

	/*
		height, err := client.BlockNumber(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("chain height :", height)

		data := manager(client, int(height))
	*/
	height := 2000
	data := manager(client, height)
	log.Println("data up to ", height, "collections amount found :", len(data.Collections))
	log.Println("displaying collecitons addresses :")
	for _, collec := range data.Collections {
		log.Println("|", collec)
	}
}
