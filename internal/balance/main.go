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
	"strconv"
	"sync"
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
		tableName := "event" + date.Format("20060102")

		// 查询表
		var blockNumber string
		query := fmt.Sprintf("SELECT coinid, fromaddress, toaddress, value, blocknumber FROM %s", tableName)
		rows, err := db.Query(query)
		if err != nil {
			g.Log().Error(ctx, err)
		}
		defer rows.Close()

		for rows.Next() {
			var coinID, fromAddress, toAddress, value string
			if err := rows.Scan(&coinID, &fromAddress, &toAddress, &value, &blockNumber); err != nil {
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
		negCeloBalances := make([]string, 0)
		for address, coinBalances := range addressMap {
			for coinID, balance := range coinBalances {
				if coinID == "5567" && balance.Sign() == -1 {
					negCeloBalances = append(negCeloBalances, address)
					continue
				}
			}
		}

		// 异步并行去调用api拿去负数的地址celo余额
		newBalances := getBalances(ctx, negCeloBalances, blockNumber)

		var index int
		for address, coinBalances := range addressMap {
			balanceF := new(big.Float)
			for coinID, balance := range coinBalances {
				price := coinPrices[coinID][dateStr]

				if coinID == "5567" && balance.Sign() == -1 {
					balance = newBalances[index]
					index++
				}

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
			insertQuery := fmt.Sprintf("INSERT INTO ods_balance_fake (date, address, balance) VALUES ('%s', '%s', '%s')", dateStr, address, balanceF.Text('f', 18))
			_, err := db.Exec(insertQuery)
			if err != nil {
				g.Log().Error(ctx, err)
			}
		}
		g.Log().Info(ctx, "Finish processing date", dateStr)
	}
}

func heightTo16(ctx context.Context, decimalString string) string {
	// decimalString := "872324"

	decimalValue, err := strconv.Atoi(decimalString)
	if err != nil {
		g.Log().Error(ctx, err)
		return ""
	}

	hexString := fmt.Sprintf("%X", decimalValue)
	return fmt.Sprint("0x" + hexString)
}

func getBalances(ctx context.Context, addreses []string, height string) map[int]*big.Int {
	url := "https://solitary-responsive-putty.celo-mainnet.quiknode.pro/40a3938f2f03f6ae973996eccf6106a9ab27c418/"

	hexHeight := heightTo16(ctx, height)

	var wg sync.WaitGroup

	results := make(map[int]*big.Int, 0)

	for i, address := range addreses {
		addr := address
		index := i
		wg.Add(1)
		go func() {
			defer wg.Done()

			rb := `{
				"method": "eth_getBalance",
				"params": ["` + addr + `", "` + hexHeight + `"],
				"id": 1,
				"jsonrpc": "2.0"
			}`
			requestBody := []byte(rb)

			resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
			if err != nil {
				g.Log().Error(ctx, "请求失败:", err)
				return
			}
			defer resp.Body.Close()

			buffer := new(bytes.Buffer)
			_, err = buffer.ReadFrom(resp.Body)
			if err != nil {
				g.Log().Error(ctx, "读取响应失败:", err)
				return
			}

			var response struct {
				Result string `json:"result"`
			}
			if err := json.Unmarshal(buffer.Bytes(), &response); err != nil {
				g.Log().Error(ctx, "JSON解析失败:", err)
				return
			}
			resultHex := response.Result
			resultBigInt := new(big.Int)
			resultBigInt.SetString(resultHex[2:], 16)

			results[index] = resultBigInt
		}()
	}

	wg.Wait()

	return results
}
