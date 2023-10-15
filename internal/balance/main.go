package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"time"

	"demo/internal/consts"
	"demo/internal/leveldb"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"

	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func main() {
	ctx := gctx.New()
	processBlanceEvent(ctx, "coin_history.txt")
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

func processBlanceEvent(ctx context.Context, coin_history_file string) {
	// 数据库连接信息
	dbName := "postgres"
	user := "creda"
	password := "20231011"
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s dbname=%s sslmode=disable password=%s", user, dbName, password))
	if err != nil {
		g.Log().Error(ctx, err)
	}
	defer db.Close()

	// 创建地址余额映射
	addressMap := make(map[string]map[string]*big.Int)

	// 读取历史价格数据并存储在coinPrices映射中
	coinPrices := make(map[string]map[string]*big.Float)
	coinHistoryFile, err := os.Open(coin_history_file)
	if err != nil {
		g.Log().Error(ctx, err)
	}
	defer coinHistoryFile.Close()

	scanner := bufio.NewScanner(coinHistoryFile)
	for scanner.Scan() {
		line := scanner.Text()
		var coinID, dateStr, priceStr string
		_, err := fmt.Sscanf(line, "%s %s %s", &coinID, &dateStr, &priceStr)
		if err != nil {
			g.Log().Error(ctx, err)
		}

		_, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			g.Log().Error(ctx, err)
		}

		price, ok := new(big.Float).SetString(priceStr)
		if !ok {
			g.Log().Error(ctx, err)
		}

		if _, exists := coinPrices[coinID]; !exists {
			coinPrices[coinID] = make(map[string]*big.Float)
		}
		coinPrices[coinID][dateStr] = price
	}

	// 表名范围
	startDate := time.Date(2020, 4, 23, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2023, 10, 13, 0, 0, 0, 0, time.UTC)

	for date := startDate; date.Before(endDate); date = date.AddDate(0, 0, 1) {
		tableName := "event" + date.Format("20060102")

		// 查询表
		query := fmt.Sprintf("SELECT coinid, fromaddress, toaddress, value FROM %s", tableName)
		rows, err := db.Query(query)
		if err != nil {
			g.Log().Error(ctx, err)
		}
		defer rows.Close()

		for rows.Next() {
			var coinID, fromAddress, toAddress string
			var value []uint8
			if err := rows.Scan(&coinID, &fromAddress, &toAddress, &value); err != nil {
				g.Log().Error(ctx, err)
			}

			intValue, ok := new(big.Int).SetString(string(value), 10)
			if !ok {
				g.Log().Error(ctx, "Failed to convert value to *big.Int")
			}

			// 更新addressMap
			if _, exists := addressMap[fromAddress]; !exists {
				addressMap[fromAddress] = make(map[string]*big.Int)
			}
			if _, exists := addressMap[toAddress]; !exists {
				addressMap[toAddress] = make(map[string]*big.Int)
			}

			// 扣除 fromAddress 的余额
			if balance, exists := addressMap[fromAddress][coinID]; exists {
				balance.Sub(balance, intValue)
			} else {
				addressMap[fromAddress][coinID] = new(big.Int).Neg(intValue)
			}

			// 增加 toAddress 的余额
			if balance, exists := addressMap[toAddress][coinID]; exists {
				balance.Add(balance, intValue)
			} else {
				addressMap[toAddress][coinID] = new(big.Int).Set(intValue)
			}
		}

		// 插入结果到 ods_balance_fake 表
		dateStr := date.Format("2006-01-02")
		for address, coinBalances := range addressMap {
			balanceF := new(big.Float)
			for coinID, balance := range coinBalances {
				price := coinPrices[coinID][dateStr]

				// 计算balance*历史价格
				if price == nil {
					price = new(big.Float)
				}
				// change balance to positive
				if balance.Sign() == -1 {
					balance = new(big.Int).Neg(balance)
				}
				balanceWithPrice := new(big.Float).SetInt(balance)
				balanceWithPrice.Mul(balanceWithPrice, price)

				balanceF.Add(balanceF, balanceWithPrice)
			}

			// 插入结果到 ods_balance_fake 表
			insertQuery := fmt.Sprintf("INSERT INTO ods_balance_fake (date, address, balance) VALUES ('%s', '%s', '%s')", dateStr, address, balanceF.Text('f', 18))
			_, err := db.Exec(insertQuery)
			if err != nil {
				g.Log().Error(ctx, err)
			}
		}
		g.Log().Info(ctx, "Finish processing date", dateStr)
	}
}
