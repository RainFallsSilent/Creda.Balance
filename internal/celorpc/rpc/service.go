package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type RpcConfig struct {
	HttpUrl string `json:"HttpUrl"`
	User    string `json:"User"`
	Pass    string `json:"Pass"`
}

type Response struct {
	ID      int64       `json:"id"`
	Version string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	*Error  `json:"error"`
}

type Error struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

func CallAndUnmarshal(method string, params Parameter, config *RpcConfig) (interface{}, error) {
	body, err := Call(method, params, config)
	if err != nil {
		return nil, err
	}

	resp := Response{}
	if err = json.Unmarshal(body, &resp); err != nil {
		return string(body), nil
	}

	if resp.Error != nil {
		return nil, errors.New(resp.Error.Message)
	}

	return resp.Result, nil
}

func Call(method string, params Parameter, config *RpcConfig) ([]byte, error) {
	url := config.HttpUrl
	var parm string
	parm = "["
	for _, p := range params {
		parm += "\"" + p + "\""

	}
	parm += "]"
	payload := []byte(`{
		"jsonrpc": "2.0",
		"method": "` + method + `",
		"params": ` + parm + `,
		"id": 1
	}`)

	fmt.Println("call:", string(payload))
	resp, err := post(url, "application/json", strings.NewReader(string(payload)))
	if err != nil {
		fmt.Println("POST requset err:", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func Unmarshal(result interface{}, target interface{}) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, target); err != nil {
		return err
	}
	return nil
}

func post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	client := *http.DefaultClient
	client.Timeout = time.Minute
	return client.Do(req)
}

type TransactionResult struct {
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
}