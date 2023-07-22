package utils

import "math/big"

func weiToEther(wei *big.Int) float64 {
	weiFloat := new(big.Float)
	weiFloat.SetString(wei.String())
	etherFloat := new(big.Float).Quo(weiFloat, big.NewFloat(1000000000000000000))
	ether, _ := etherFloat.Float64()
	return ether
}
