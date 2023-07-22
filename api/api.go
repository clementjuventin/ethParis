package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type ERC721TxStruct struct {
	Hash       string
	Tag        string
	FromAddr   common.Address
	ToAddr     common.Address
	Value      string
	TokenId    string
	Collection common.Address
}

func main() {
	router := gin.Default()
	router.POST("/userTx/:addr", getUserTransactions)

	router.Run("localhost:8080")
}

func getDbInstance() (database *sql.DB, e error) {
	const connString string = "postgres://postgres:postgres@localhost/indexer_db?sslmode=disable"
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}
	return db, nil
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

	rows, err := db.Query("SELECT * FROM ERC721Tx WHERE fromAddr = $1 OR toAddr = $1 ORDER BY id DESC LIMIT 100 OFFSET $2", address, offset)
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
		err = rows.Scan(&id, &tx.Hash, &tx.Tag, &fromAddress, &toAddress, &tx.Value, &tx.TokenId, &collection)
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

func getContractTransactions(c *gin.Context) {
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

	rows, err := db.Query("SELECT * FROM ERC721Tx WHERE collection = $1 ORDER BY id DESC LIMIT 100 OFFSET $2", address, offset)
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
		err = rows.Scan(&id, &tx.Hash, &tx.Tag, &fromAddress, &toAddress, &tx.Value, &tx.TokenId, &collection)
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

func getNftPriceHistory(c *gin.Context) {

}

func getUserNFT(c *gin.Context) {
	address := c.Param("addr")
	offset := c.Param("offset")
	log.Println("Address :", offset)
	if offset == "" {
		offset = "0"
	}

	db, err := getDbInstance()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Select all received NFTs that was not sent after
	rows, err := db.Query("SELECT * FROM ERC721Tx WHERE toAddr = $1 AND  ORDER BY id DESC LIMIT 100 OFFSET $2", address, offset)
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
		err = rows.Scan(&id, &tx.Hash, &tx.Tag, &fromAddress, &toAddress, &tx.Value, &tx.TokenId, &collection)
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
