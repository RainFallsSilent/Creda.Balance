package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"

	"demo/internal/balance/leveldb"
	"demo/internal/consts"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	ctx := gctx.New()
	SyncBalances(ctx)
}

func SyncBalances(ctx context.Context) {
	coinName := gcmd.GetArg(1, "eth").String()
	startHeight := gcmd.GetArg(2, "100").Uint32()
	// url := gcmd.GetArg(3, "https://api.elastos.io/esc")
	// url := gcmd.GetArg(3, "https://api–testnet.elastos.io/esc")
	url := gcmd.GetArg(3, "https://api.elastos.io/ela").String()
	g.Log().Info(ctx, "start sync balances")
	g.Log().Infof(ctx, "sync balances for coin: %s", coinName)
	g.Log().Infof(ctx, "sync balances from height: %d", startHeight)
	g.Log().Infof(ctx, "sync balances from url: %s", url)
	syncAndStore(ctx, coinName, url, startHeight)
	g.Log().Info(ctx, "end sync balances")
}

const (
	BTC CoinType = iota
	ETH
	Abr
	// 1000
)

type CoinType int

func syncAndStore(ctx context.Context, coinName, url string, startHeight uint32) {

	method := "POST"
	payload := map[string]interface{}{
		"method": "getblockbyheight",
		"params": map[string]interface{}{
			"height": "1479760",
		},
	}

	payloadBytes, _ := json.Marshal(payload)

	client := &http.Client{}
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(payloadBytes))

	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		g.Log().Error(ctx, err)
		return
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	g.Log().Info(ctx, result)

	// store to db
	db, err := leveldb.NewDataStore(consts.DBPath)
	if err != nil {
		g.Log().Error(ctx, err)
		return
	}

	db.Balance().BatchPut(
		[]leveldb.BalanceInfo{
			{
				Address: []byte("address"),
				CoinId:  1,
				Day:     2,
				Balance: big.NewInt(1024).Bytes(),
			},
		},
	)
	db.Balance().Commit()

	balance, err := db.Balance().Get([]byte("address"), 1, 1)
	if err != nil {
		g.Log().Error(ctx, err)
		return
	}
	g.Log().Info(ctx, "day 1 balance:", big.NewInt(0).SetBytes(balance))

	balance, err = db.Balance().Get([]byte("address"), 1, 2)
	if err != nil {
		g.Log().Error(ctx, err)
		return
	}
	g.Log().Info(ctx, "day 2 balance:", big.NewInt(0).SetBytes(balance))

	balance, err = db.Balance().Get([]byte("address"), 1, 3)
	if err != nil {
		g.Log().Error(ctx, err)
		return
	}
	g.Log().Info(ctx, "day 3 balance:", big.NewInt(0).SetBytes(balance))

	balance, err = db.Balance().Get([]byte("address"), 2, 3)
	if err != nil {
		if err.Error() == "leveldb: not found" {
			g.Log().Info(ctx, "day 3 but coin id 2 balance is nil")
			return
		}
		g.Log().Error(ctx, err)
		return
	}
}
