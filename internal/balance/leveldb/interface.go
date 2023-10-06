package leveldb

// DB is the common interface to all database implementations.
type DB interface {
	// Clear delete all data in database.
	Clear() error

	// Close database.
	Close() error
}

// Batch is the common interface to all batch implementations.
type Batch interface {
	// Put sets the value for the given key. It overwrites any previous value
	Rollback() error

	// Delete deletes the value for the given key.
	Commit() error
}

type Balance interface {
	DB
	Batch

	// Put sets the value for the given key. It overwrites any previous value
	Put(bi BalanceInfo) error

	// BatchPut sets the value for the given key. It overwrites any previous value
	BatchPut(bis []BalanceInfo) error

	// Get returns the value for the given key. It returns nil if the DB does not contains the key.
	Get(address []byte, coinId uint32, day uint32) (balance []byte, err error)
}

type Price interface {
	DB
	Batch

	// Put sets the value for the given key. It overwrites any previous value
	Put(pi PriceInfo) error

	// BatchPut sets the value for the given key. It overwrites any previous value
	BatchPut(pis []PriceInfo) error

	// Get returns the value for the given key. It returns nil if the DB does not contains the key.
	Get(coinId uint32, day uint32) (price []byte, err error)
}

type DataStore interface {
	DB

	// Balance returns the Balance interface.
	Balance() Balance

	// Price returns the Price interface.
	Price() Price
}
