package leveldb

import (
	"demo/utility"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type PriceInfo struct {
	CoinId uint32
	Day    uint32 // example 20230101
	Price  []byte
}

func (p *PriceInfo) Key() []byte {
	return toKey(LDBPriceKey, append(utility.Uint32ToBytes(p.CoinId), utility.Uint32ToBytes(p.Day)...)...)
}

func (p *PriceInfo) Value() []byte {
	return p.Price
}

type PriceDB struct {
	Batch
	sync.RWMutex
	db *leveldb.DB
	b  *leveldb.Batch
}

func NewPriceDB(db *leveldb.DB) *PriceDB {
	return &PriceDB{
		db: db,
		b:  new(leveldb.Batch),
	}
}

func (c *PriceDB) Put(pi PriceInfo) error {
	c.Lock()
	defer c.Unlock()
	if err := c.batchPut([]PriceInfo{pi}, c.b); err != nil {
		return err
	}
	c.db.Write(c.b, nil)
	return nil
}

func (c *PriceDB) batchPut(pis []PriceInfo, batch *leveldb.Batch) error {
	for _, pi := range pis {
		batch.Put(pi.Key(), pi.Value())
	}
	return nil
}

func (c *PriceDB) BatchPut(pis []PriceInfo) error {
	c.Lock()
	defer c.Unlock()
	return c.batchPut(pis, c.b)
}

func (c *PriceDB) Get(coinId uint32, day uint32) (price []byte, err error) {
	c.RLock()
	defer c.RUnlock()
	return c.get(coinId, day)
}

func (c *PriceDB) get(coinId uint32, day uint32) (price []byte, err error) {
	key := toKey(LDBPriceKey, append(utility.Uint32ToBytes(coinId), utility.Uint32ToBytes(day)...)...)
	return c.db.Get(key, nil)
}

func (c *PriceDB) Clear() error {
	c.Lock()
	defer c.Unlock()
	iter := c.db.NewIterator(util.BytesPrefix(LDBPriceKey), nil)
	for iter.Next() {
		c.b.Delete(iter.Key())
	}
	iter.Release()
	return c.db.Write(c.b, nil)
}

func (c *PriceDB) Close() error {
	c.db.Close()
	return nil
}

func (c *PriceDB) Rollback() error {
	c.b.Reset()

	return nil
}

func (c *PriceDB) Commit() error {
	return c.db.Write(c.b, nil)
}


func (c *PriceDB) CommitBatch(batch *leveldb.Batch) error {
	return c.db.Write(batch, nil)
}

func (c *PriceDB) RollbackBatch(batch *leveldb.Batch) error {
	batch.Reset()
	return nil
}