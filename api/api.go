package main

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-contrib/cors"
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

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}                             // Replace with your React app's URL
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"} // Add other headers if needed
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}           // Add other methods your API supports
	router.Use(cors.New(config))

	router.POST("/nft/history/:collection/:tokenId", getNftHistory)
	router.POST("/nft/:collection/:tokenId", getNftData)

	router.POST("/collection/:addr", getCollectionNfts)
	router.POST("/collection/history/:addr", getCollectionHistory)
	router.POST("/collection/stats/:addr", getCollectionStats)

	router.POST("/address/history/:addr", getAddressHistory)
	router.POST("/address/:addr", getAddressNfts)

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

	rows, err := db.Query("SELECT * FROM ERC721Tx WHERE collection = $1 AND token_id = $2 ORDER BY timestamp DESC LIMIT 100 OFFSET $3", collection, tokenId, offset)
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

	rows, err := db.Query("SELECT * FROM ERC721 WHERE owner = $1 ORDER BY mint_timestamp DESC LIMIT 100 OFFSET $2", address, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var nfts []ERC721Struct
	for rows.Next() {
		var nft ERC721Struct
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
func getCollectionNfts(c *gin.Context) {
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
	rows, err := db.Query("SELECT * FROM ERC721 WHERE collection = $1 ORDER BY mint_timestamp DESC LIMIT 100 OFFSET $2", address, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var nfts []ERC721Struct
	for rows.Next() {
		var nft ERC721Struct
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
func getAddressHistory(c *gin.Context) {
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

	rows, err := db.Query("SELECT * FROM ERC721Tx WHERE from_addr = $1 OR to_addr = $1 ORDER BY timestamp DESC LIMIT 100 OFFSET $2", address, offset)
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
func getCollectionHistory(c *gin.Context) {
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

	rows, err := db.Query("SELECT * FROM ERC721Tx WHERE collection = $1 ORDER BY timestamp DESC LIMIT 100 OFFSET $2", address, offset)
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
func getNftData(c *gin.Context) {
	collection := c.Param("collection")
	tokenId := c.Param("tokenId")

	db, err := getDbInstance()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	collection = strings.ToLower(collection)
	tokenId = strings.ToLower(tokenId)

	rows, err := db.Query("SELECT * FROM ERC721 WHERE collection = $1 AND token_id = $2", collection, tokenId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var nft ERC721Struct
	for rows.Next() {
		owner := ""
		collection := ""
		err = rows.Scan(&nft.MintTimestamp, &nft.MintBlockNumber, &nft.MintTxHash, &nft.URI, &nft.TokenId, &collection, &owner)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		nft.Collection = common.HexToAddress(collection)
		nft.Owner = common.HexToAddress(owner)
	}

	c.JSON(http.StatusOK, gin.H{"data": nft})
}
func getCollectionStats(c *gin.Context) {
	address := c.Param("addr")
	since := c.Param("since")
	if since == "" {
		// Date.now - 24h
		since = strings.Split(time.Now().Add(-24*time.Hour).String(), " ")[0]
	}
	db, err := getDbInstance()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	const ownerCountQuery = `SELECT COUNT(DISTINCT owner) FROM ERC721 WHERE collection = $1`
	const txCountSinceQuery = `SELECT COUNT(DISTINCT HASH) FROM ERC721Tx WHERE collection = $1 AND timestamp > $2`
	// const volumeSinceQuery = `SELECT SUM(value)::numeric FROM ERC721Tx WHERE collection = $1 AND timestamp > $2`

	oCount := 0
	tCount := 0
	// vCount := 0

	rows, err := db.Query(ownerCountQuery, address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&oCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	rows, err = db.Query(txCountSinceQuery, address, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&tCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// rows, err = db.Query(volumeSinceQuery, address, since)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	// defer rows.Close()

	// for rows.Next() {
	// 	err = rows.Scan(&vCount)
	// 	if err != nil {
	// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 		return
	// 	}
	// }
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"ownerCount": oCount, "txCount": tCount, "volume": 0}})
}
