package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type TransactionResult struct {
	ID      int    `json:"id"`
	JsonRPC string `json:"jsonrpc"`
	Result  struct {
		BlockHash        string `json:"blockHash"`
		BlockNumber      string `json:"blockNumber"`
		ChainId          string `json:"chainId"`
		From             string `json:"from"`
		Gas              string `json:"gas"`
		GasPrice         string `json:"gasPrice"`
		Hash             string `json:"hash"`
		Input            string `json:"input"`
		Nonce            string `json:"nonce"`
		R                string `json:"r"`
		S                string `json:"s"`
		To               string `json:"to"`
		TransactionIndex string `json:"transactionIndex"`
		Type             string `json:"type"`
		V                string `json:"v"`
		Value            string `json:"value"`
	} `json:"result"`
}

func main() {
	// Your Infura API key and the Celo RPC URL
	apiKey := "76891b8517e248fe9a49473d68f8f7f7"
	rpcURL := "https://mainnet.infura.io/v3/" + apiKey

	// Prepare the JSON-RPC request payload
	payload := []byte(`{
		"jsonrpc": "2.0",
		"method": "eth_getTransactionByHash",
		"params": ["0xbb3a336e3f823ec18197f1e13ee875700f08f03e2cab75f0d0b118dabb44cba0"],
		"id": 1
	}`)

	// Send the POST request to the Celo RPC
	resp, err := http.Post(rpcURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	// Read and print the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	fmt.Println(string(body))

	// Parse the JSON response into the struct
	var response TransactionResult
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	// Access and print transaction information
	fmt.Println("Celo RPC Response:")
	fmt.Println("Block Hash:", response.Result.BlockHash)
	fmt.Println("From:", response.Result.From)
	fmt.Println("Gas Price:", response.Result.GasPrice)
	// ... and so on for other fields
}
