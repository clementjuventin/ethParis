package main

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type ERC721CollectionStruct struct {
	ContractAddress   common.Address
	ContractName      string
	ContractSymbol    string
	DeployTimestamp   uint64
	DeployBlockNumber uint64
	DeployTxHash      string
}

type ERC721TxStruct struct {
	Timestamp   uint64
	BlockNumber uint64
	TxHash      string
	Tag         string
	FromAddr    common.Address
	ToAddr      common.Address
	Value       string
	TokenId     string
	Collection  common.Address
}

type ERC721Struct struct {
	MintTimestamp   uint64
	MintBlockNumber uint64
	MintTxHash      string
	URI             string
	TokenId         string
	Collection      common.Address
	Owner           common.Address
}

func main() {
	router := gin.Default()
	router.POST("/nft/history/:collection/:tokenId", getNftHistory)
	router.POST("/nft/:addr", getAddressNfts)

	router.POST("/userTx/:addr", getUserTransactions)

	router.Run("localhost:8080")
}

func getDbInstance() (database *sql.DB, e error) {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost/indexer_v2?sslmode=disable")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func getNftHistory(c *gin.Context) {
	collection := c.Param("collection")
	tokenId := c.Param("tokenId")
	offset := c.Param("offset")
	if offset == "" {
		offset = "0"
	}

	db, err := getDbInstance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	collection = strings.ToLower(collection)
	tokenId = strings.ToLower(tokenId)

	rows, err := db.Query("SELECT * FROM ERC721Tx WHERE collection = $1 AND token_id = $2 ORDER BY id DESC LIMIT 100 OFFSET $3", collection, tokenId, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var txs []ERC721TxStruct
	for rows.Next() {
		var tx ERC721TxStruct
		id := 0
		fromAddress := ""
		toAddress := ""
		collection := ""
		err = rows.Scan(&id, &tx.Timestamp, &tx.BlockNumber, &tx.TxHash, &tx.Tag, &fromAddress, &toAddress, &tx.Value, &tx.TokenId, &collection)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		tx.FromAddr = common.HexToAddress(fromAddress)
		tx.ToAddr = common.HexToAddress(toAddress)
		tx.Collection = common.HexToAddress(collection)
		txs = append(txs, tx)
	}

	c.JSON(http.StatusOK, gin.H{"data": txs})
}
func getAddressNfts(c *gin.Context) {
	address := c.Param("addr")
	offset := c.Param("offset")
	if offset == "" {
		offset = "0"
	}

	db, err := getDbInstance()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := db.Query("SELECT * FROM ERC721 WHERE owner = $1 ORDER BY id DESC LIMIT 100 OFFSET $2", address, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var nfts []ERC721Struct
	for rows.Next() {
		var nft ERC721Struct
		// The first parameter is the id, we don't need it
		owner := ""
		collection := ""
		err = rows.Scan(&nft.MintTimestamp, &nft.MintBlockNumber, &nft.MintTxHash, &nft.URI, &nft.TokenId, &collection, &owner)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		nft.Collection = common.HexToAddress(collection)
		nft.Owner = common.HexToAddress(owner)
		nfts = append(nfts, nft)
	}

	c.JSON(http.StatusOK, gin.H{"data": nfts})
}

func getUserTransactions(c *gin.Context) {
	address := c.Param("addr")
	offset := c.Param("offset")
	if offset == "" {
		offset = "0"
	}

	db, err := getDbInstance()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := db.Query("SELECT * FROM ERC721Tx WHERE from_addr = $1 OR to_addr = $1 ORDER BY id DESC LIMIT 100 OFFSET $2", address, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var txs []ERC721TxStruct
	for rows.Next() {
		var tx ERC721TxStruct
		id := 0
		// The first parameter is the id, we don't need it
		fromAddress := ""
		toAddress := ""
		collection := ""
		err = rows.Scan(&id, &tx.Timestamp, &tx.BlockNumber, &tx.TxHash, &tx.Tag, &fromAddress, &toAddress, &tx.Value, &tx.TokenId, &collection)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		tx.FromAddr = common.HexToAddress(fromAddress)
		tx.ToAddr = common.HexToAddress(toAddress)
		tx.Collection = common.HexToAddress(collection)
		txs = append(txs, tx)
	}

	c.JSON(http.StatusOK, gin.H{"data": txs})
}
