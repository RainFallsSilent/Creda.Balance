package leveldb

import (
	"bytes"
	"demo/utility"
	"math/big"
	"sync"

	"github.com/elastos/Elastos.ELA/common"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// Ensure balanceDB implement BalanceDB interface.
var _ Balance = (*BalanceDB)(nil)

type BalanceInfo struct {
	Address []byte
	CoinId  uint32
	Day     uint32 // example 20230101
	Balance []byte
}

func (b *BalanceInfo) Key() []byte {
	return toKey(LDBBalanceKey, append(b.Address, utility.Uint32ToBytes(b.CoinId)...)...)
}

func (b *BalanceInfo) Value() []byte {
	buf := new(bytes.Buffer)
	common.WriteUint32(buf, b.Day)
	common.WriteVarBytes(buf, b.Balance)
	return buf.Bytes()
}

type BalanceDB struct {
	Batch
	sync.RWMutex
	db *leveldb.DB
	b  *leveldb.Batch
}

func NewBalanceDB(db *leveldb.DB) *BalanceDB {
	return &BalanceDB{
		db: db,
		b:  new(leveldb.Batch),
	}
}

func (c *BalanceDB) Put(bi BalanceInfo) error {
	c.Lock()
	defer c.Unlock()
	if err := c.batchPut([]BalanceInfo{bi}, c.b); err != nil {
		return err
	}
	c.db.Write(c.b, nil)
	return nil
}

func (c *BalanceDB) batchPut(bis []BalanceInfo, batch *leveldb.Batch) error {
	for _, bi := range bis {
		batch.Put(bi.Key(), bi.Value())
	}

	return nil
}

func (c *BalanceDB) BatchPut(bis []BalanceInfo) error {
	c.Lock()
	defer c.Unlock()
	return c.batchPut(bis, c.b)
}

func (c *BalanceDB) Get(address []byte, coinId uint32, day uint32) (balance []byte, err error) {
	c.RLock()
	defer c.RUnlock()
	return c.get(address, coinId, day)
}

func (c *BalanceDB) get(address []byte, coinId uint32, day uint32) (balance []byte, err error) {
	bi := BalanceInfo{
		Address: address,
		CoinId:  coinId,
		Day:     day,
	}
	balanceData, err := c.db.Get(bi.Key(), nil)
	if err != nil {
		println("get balance error:", err)
		return nil, err
	}

	if len(balanceData) == 0 {
		return nil, nil
	}

	// serialize balanceData to balances
	buf := bytes.NewBuffer(balanceData)
	for {
		dbDay, err := common.ReadUint32(buf)
		if err != nil {
			// todo: make it better, maybe should store count of balance
			return big.NewInt(0).Bytes(), nil
		}
		balance, err := common.ReadVarBytes(buf, 64, "balance")
		if err != nil {
			return nil, err
		}
		if bi.Day >= dbDay {
			return balance, nil
		}
	}
}

func (c *BalanceDB) Close() error {
	c.db.Close()
	return nil
}

func (c *BalanceDB) Clear() error {
	c.Lock()
	defer c.Unlock()
	it := c.db.NewIterator(util.BytesPrefix(LDBBalanceKey), nil)
	defer it.Release()
	for it.Next() {
		c.b.Delete(it.Key())
	}
	it.Release()
	return c.db.Write(c.b, nil)
}

func (c *BalanceDB) Commit() error {
	return c.db.Write(c.b, nil)
}

func (c *BalanceDB) Rollback() error {
	c.b.Reset()
	return nil
}

func (c *BalanceDB) CommitBatch(batch *leveldb.Batch) error {
	return c.db.Write(batch, nil)
}

func (c *BalanceDB) RollbackBatch(batch *leveldb.Batch) error {
	batch.Reset()
	return nil
}

func toKey(bucket []byte, index ...byte) []byte {
	return append(bucket, index...)
}
