package leveldb

import (
	"path/filepath"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

// Ensure dataStore implement DataStore interface.
var _ DataStore = (*dataStore)(nil)

type dataStore struct {
	sync.RWMutex
	db      *leveldb.DB
	balance *BalanceDB
	price   *PriceDB
}

func NewDataStore(dataDir string) (*dataStore, error) {
	db, err := leveldb.OpenFile(filepath.Join(dataDir, "store"), nil)
	if err != nil {
		return nil, err
	}

	return &dataStore{
		db:      db,
		balance: NewBalanceDB(db),
		price:   NewPriceDB(db),
	}, nil
}

func (ds *dataStore) Balance() Balance {
	return ds.balance
}

func (ds *dataStore) Price() Price {
	return ds.price
}

func (ds *dataStore) Batch() *leveldb.Batch {
	return new(leveldb.Batch)
}

func (d *dataStore) Clear() error {
	d.Lock()
	defer d.Unlock()

	it := d.db.NewIterator(nil, nil)
	batch := new(leveldb.Batch)
	for it.Next() {
		batch.Delete(it.Key())
	}
	it.Release()

	return d.db.Write(batch, nil)
}

func (ds *dataStore) Close() error {
	ds.Lock()
	defer ds.Unlock()
	ds.balance.Close()
	ds.price.Close()
	return ds.db.Close()
}
