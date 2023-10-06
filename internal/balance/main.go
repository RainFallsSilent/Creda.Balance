package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

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

}
