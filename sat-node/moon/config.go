package moon

type Config struct {
	RaftStoragePath    string
	RocksdbStoragePath string
}

var DefaultConfig Config
