package moon

import "satweave/sat-node/infos"

type Config struct {
	ClusterInfo        infos.ClusterInfo
	RaftStoragePath    string
	RocksdbStoragePath string
}

var DefaultConfig Config

func init() {
	DefaultConfig = Config{
		ClusterInfo:        infos.ClusterInfo{},
		RaftStoragePath:    "./satweave-data/moon/raft/",
		RocksdbStoragePath: "./satweave-data/moon/rocksdb/",
	}
}
