package main

import (
	"fmt"

	"demo/internal/celorpc/rpc"
)

func main() {

	res, err := rpc.GetBlockByHash("0xbb3a336e3f823ec18197f1e13ee875700f08f03e2cab75f0d0b118dabb44cba0")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("response:", res)
}
