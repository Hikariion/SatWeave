package config

import (
	"github.com/mohae/deepcopy"
	"net"
	"os"
	"path"
	"satweave/sat-node/worker"
	"satweave/utils/common"
	"satweave/utils/config"
)

const rpcPort = 3267
const httpPort = 3268

type Config struct {
	IpAddr      string        `json:"IpAddr"`
	RpcPort     uint64        `json:"RpcPort"`
	HttpPort    uint64        `json:"HttpPort"`
	StoragePath string        `json:"StoragePath"`
	WorkConfig  worker.Config `json:"WorkConfig"`
}

var DefaultConfig Config

func init() {
	// 初始化 ip
	// TODO(qiu): 这里可能会有坑，会得到docker内部的地址，但是需要的是宿主机的ip
	_, ipAddr := getSelfIpAddr()
	DefaultConfig = Config{
		IpAddr: ipAddr,
	}
}

// InitConfig check config and init data dir and set some empty config value
func InitConfig(conf *Config) error {
	var defaultConfig Config
	// TODO(qiu): 看看是啥意思
	_ = config.GetDefaultConf(&defaultConfig)

	// TODO(qiu): 为什么要检查ip？
	// check ip address
	//if !isAvailableIpAdder(conf.WatcherConfig.SelfNodeInfo.IpAddr) {
	//	conf.WatcherConfig.SelfNodeInfo.IpAddr =
	//		defaultConfig.WatcherConfig.SelfNodeInfo.IpAddr
	//}

	// check storagePath
	if conf.StoragePath == "" {
		conf.StoragePath = defaultConfig.StoragePath
	}

	err := common.InitPath(conf.StoragePath)
	if err != nil {
		return err
	}

	// read persist config file in storage path
	storagePath := conf.StoragePath
	confPath := path.Join(storagePath + "/config/edge_node.json")
	s, err := os.Stat(confPath)
	persistConf := deepcopy.Copy(DefaultConfig).(Config)
	if err == nil && !s.IsDir() && s.Size() > 0 {
		_ = config.Read(confPath, &persistConf)
	}

	// save persist config file in storage path
	err = config.WriteToPath(conf, confPath)
	return err
}

func getSelfIpAddr() (error, string) {
	addrs, err := net.InterfaceAddrs()
	selfIp := ""
	if err != nil {
		return err, selfIp
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				selfIp = ipnet.IP.String()
			}
		}
	}
	return nil, selfIp
}

func isAvailableIpAdder(addr string) bool {
	if addr == "" {
		return false
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}
	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.String() == addr {
				return true
			}
		}
	}
	return false
}
