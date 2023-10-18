package rpc

import "fmt"

func GetBlockByHash(hash string) (*TransactionResult, error) {
	resp, err := CallAndUnmarshal("eth_getTransactionByHash", ParamList(hash), CeloRpcConfig)
	if err != nil {
		return nil, err
	}
	var res TransactionResult
	if err = Unmarshal(&resp, &res); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err
	}

	return &res, nil
}
