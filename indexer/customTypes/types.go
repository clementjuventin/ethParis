package customTypes

import (
	"github.com/ethereum/go-ethereum/common"
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
