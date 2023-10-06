package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"

	"demo/internal/consts"
	"demo/internal/leveldb"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://sandbox-api.coinmarketcap.com/v1/cryptocurrency/listings/latest", nil)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	q := url.Values{}
	q.Add("start", "1")
	q.Add("limit", "5000")
	q.Add("convert", "USD")

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", "b54bcf4d-1bca-4e8e-9a24-22ff2c3d462c")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request to server")
		os.Exit(1)
	}
	fmt.Println(resp.Status)
	respBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(respBody))

	ctx := gctx.New()
	// store to db
	db, err := leveldb.NewDataStore(consts.DBPath)
	if err != nil {
		g.Log().Error(ctx, err)
		return
	}

	db.Price().BatchPut(
		[]leveldb.PriceInfo{
			{
				CoinId: 1,
				Day:    20210101,
				Price:  big.NewInt(1000000000000000000).Bytes(),
			},
			{
				CoinId: 1,
				Day:    20210102,
				Price:  big.NewInt(2000000000000000000).Bytes(),
			},
			{
				CoinId: 1,
				Day:    20210103,
				Price:  big.NewInt(3000000000000000000).Bytes(),
			},
		},
	)
	db.Price().Commit()

	price, err := db.Price().Get(1, 20210101)
	if err != nil {
		g.Log().Error(ctx, err)
		return
	}
	g.Log().Info(ctx, "day 20210101 price:", big.NewInt(0).SetBytes(price))

	price, err = db.Price().Get(1, 20210102)
	if err != nil {
		g.Log().Error(ctx, err)
		return
	}
	g.Log().Info(ctx, "day 20210102 price:", big.NewInt(0).SetBytes(price))

}
