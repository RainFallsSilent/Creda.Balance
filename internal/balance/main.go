package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"math"
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

	taskChan := make(chan struct{}, 5)

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

	// 代币小数位
	// '0x617f3112bf5397D0467D315cC709EF968D9ba546': {name: 'USDT', coinID: 825, decimals: 6},
	// '0xef4229c8c3250C675F21BCefa42f58EfbfF6002a': {name: 'USDC', coinID: 3408,decimals: 6},
	// '0x37f750B7cC259A2f741AF45294f6a16572CF5cAd': {name:'USDC(WormHole)', coinID:20650,decimals:6},
	// '0xD629eb00dEced2a080B7EC630eF6aC117e614f1b': {name:'WBTC', coinID:3717,decimals:18},
	// '0x471EcE3750Da237f93B8E339c536989b8978a438': {name:'CELO', coinID:  5567,decimals:18},
	// '0x29dFce9c22003A4999930382Fd00f9Fd6133Acd1': {name:'SUSHI', coinID:  6758,decimals:18},
	// '0xB9C8F0d3254007eE4b98970b94544e473Cd610EC':{name:'MIMATIC', coinID:10238,decimals:18},
	// '0xD8763CBa276a3738E6DE85b4b3bF5FDed6D6cA73':{name:'cEUR', coinID:9467,decimals:18},
	// '0x9995cc8F20Db5896943Afc8eE0ba463259c931ed':{name:'ETHIX',coinID:8442,decimals:18},
	// '0x765DE816845861e75A25fCA122bb6898B8B1282a':{name:'cUSD',coinID:7236,decimals:18},
	// '0x1d18d0386F51ab03E7E84E71BdA1681EbA865F1f':{name:'JMPT',coinID:17334,decimals:18},
	// '0x27cd006548dF7C8c8e9fdc4A67fa05C2E3CA5CF9':{name:'PLASTIK',coinID:15575,decimals:9},
	// '0xEe9801669C6138E84bD50dEB500827b776777d28':{name:'O3',coinID:9588,decimals:18},
	// '0x6e512BFC33be36F2666754E996ff103AD1680Cc9':{name:'ABR',coinID:12212,decimals:18},
	// '0x00Be915B9dCf56a3CBE739D9B9c202ca692409EC':{name:'UBE',coinID:10808,decimals:18},
	// '0x17700282592D6917F6A73D0bF8AcCf4D578c131e':{name:'MOO',coinID:13021,decimals:18},
	// '0xe8537a3d056DA446677B9E9d6c5dB704EaAb4787':{name:'CREAL',coinID:16385,decimals:18},
	decimals := make(map[string]int) // key: coinID, value: decimals
	decimals["825"] = 6
	decimals["3408"] = 6
	decimals["20650"] = 6
	decimals["3717"] = 18
	decimals["5567"] = 18
	decimals["6758"] = 18
	decimals["10238"] = 18
	decimals["9467"] = 18
	decimals["8442"] = 18
	decimals["7236"] = 18
	decimals["17334"] = 18
	decimals["15575"] = 9
	decimals["9588"] = 18
	decimals["12212"] = 18
	decimals["10808"] = 18
	decimals["13021"] = 18
	decimals["16385"] = 18

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

		eventTableName := "event" + date.Format("20060102")
		g.Log().Info(ctx, "###### start get events")
		// 查询表
		query := fmt.Sprintf("SELECT coinid, fromaddress, toaddress, value FROM %s", eventTableName)
		rows, err := db.Query(query)
		if err != nil {
			g.Log().Error(ctx, err)
		}
		defer rows.Close()

		for rows.Next() {
			var coinID, fromAddress, toAddress, value string
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

		dateT := date

		taskChan <- struct{}{}

		go func() {
			tableName := "ods_balance_" + dateT.Format("20060102")

			exists, err := tableExists(db, tableName)
			if err != nil {
				g.Log().Error(ctx, err)
			} else {
				if exists {
					g.Log().Info(ctx, "table ", tableName, "exist skip")
					<-taskChan
					return
				} else {
					createTableSQL := `CREATE TABLE ` + tableName + ` (date VARCHAR(255),address VARCHAR(255),	balance VARCHAR(255));`
					_, err = db.Exec(createTableSQL)
					if err != nil {
						g.Log().Error(ctx, err)
					}
					g.Log().Info(ctx, "create table", tableName, "success")
				}
			}

			g.Log().Info(ctx, "###### start save to db:", tableName)

			// 插入结果到 ods_balance_fake 表
			tx, err := db.Begin()
			if err != nil {
				g.Log().Error(ctx, err)
			}
			defer tx.Rollback()

			dateStr := dateT.Format("2006-01-02")
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
					// decimail
					decimal := decimals[coinID]
					// balance with decimal * price
					balanceWithPrice := new(big.Float).SetInt(balance)
					balanceWithPrice.Quo(balanceWithPrice, new(big.Float).SetFloat64(math.Pow(float64(10), float64(decimal))))
					balanceWithPrice.Mul(balanceWithPrice, price)

					// g.Log().Info(ctx, "decimal", decimal, "price:", price, "balance:", balance, "balanceWithPrice:", balanceWithPrice)

					balanceF.Add(balanceF, balanceWithPrice)
				}

				// if balanceF equal 0, then continue
				if balanceF.Cmp(big.NewFloat(0)) == 0 {
					continue
				}

				// 插入结果到 ods_balance_fake 表
				// 创建预处理语句
				stmt, err := tx.Prepare("INSERT INTO " + tableName + "(date, address, balance) VALUES($1, $2, $3)")
				if err != nil {
					g.Log().Error(ctx, err)
				}
				defer stmt.Close()

				_, err = stmt.Exec(dateStr, address, balanceF.Text('f', 18))
				if err != nil {
					g.Log().Error(ctx, err)
				}
			}
			g.Log().Info(ctx, "###### start commit:", tableName)

			// 提交事务
			if err := tx.Commit(); err != nil {
				g.Log().Error(ctx, err)
			}
			g.Log().Info(ctx, "Finish processing date", dateStr)
			<-taskChan
		}()
	}

}

func tableExists(db *sql.DB, tableName string) (bool, error) {
	query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = '%s')", tableName)
	var exists bool
	err := db.QueryRow(query).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
