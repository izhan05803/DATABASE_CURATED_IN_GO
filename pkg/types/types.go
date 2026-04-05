package types

import "errors"

// Sentinel errors
var (
	ErrKeyNotFound   = errors.New("key not found")
	ErrPageFull      = errors.New("page is full")
	ErrInvalidFormat = errors.New("invalid file format")
)

// PageType represents the type of a storage page
type PageType uint

const (
	PageTypeLeaf PageType = iota
	PageTypeInternal
	PageTypeOverflow
)

// Record represents a key-value pair with metadata
type Record struct {
	Key       string
	Value     []byte
	Timestamp int64
	Deleted   bool
}

// Page represents a 4KB storage unit
type Page struct {
	PageID   uint32
	PageType PageType
	Records  []Record
	NextPage uint32
}

// Command represents a parsed CLI command
type Command struct {
	Type string
	Args []string
}

// Result represents the result of executing a command
type Result struct {
	Success bool
	Message string
	Data    interface{}
}

// Storage defines the storage layer contract
type Storage interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	Delete(key string) error
	Scan(prefix string) ([]Record, error)
	Close() error
}

// Index defines the indexing contract
type Index interface {
	Search(key string) (uint32, bool)
	Insert(key string, pageID uint32) error
	Delete(key string) error
}
