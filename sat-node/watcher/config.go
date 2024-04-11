package watcher

import (
	"satweave/sat-node/infos"
	"time"
)

type Config struct {
	SunAddr                string
	SelfNodeInfo           infos.NodeInfo
	ClusterInfo            infos.ClusterInfo
	NodeInfoCommitInterval time.Duration
	ClusterName            string
	CloudAddr              string
	CloudPort              uint64
	TaskFileStoragePath    string
	OssHost                string
}

var DefaultConfig Config

func init() {
	DefaultConfig = Config{
		SunAddr:                "",
		ClusterInfo:            infos.ClusterInfo{},
		SelfNodeInfo:           *infos.NewSelfInfo(1, "127.0.0.1", 0),
		NodeInfoCommitInterval: time.Second * 2,
		ClusterName:            "satweave_dev",
		OssHost:                "http://192.168.6.100:60402",
	}
}
