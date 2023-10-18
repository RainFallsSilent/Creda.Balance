package main

import (
	"fmt"
	"time"

	"demo/internal/celorpc/rpc"

	"github.com/celo-org/celo-blockchain/core/types"
)

func main() {

	block := types.Block{
		ReceivedAt:   time.Time{},
		ReceivedFrom: nil,
	}
	fmt.Println("block:", block)

	res, err := rpc.GetTransactionByHash("0xbb3a336e3f823ec18197f1e13ee875700f08f03e2cab75f0d0b118dabb44cba0")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("response:", res)
}
