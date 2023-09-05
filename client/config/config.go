package config

type ClientConfig struct {
	// 客户端的 IP
	ClientIp string
	// rpc 端口
	RpcPort uint64
	// 卫星返回的文件的存储路径
	FileFromSatelliteStoragePath string
}
