package task_manager

type Config struct {
	SunAddr   string
	CloudAddr string
	CloudPort uint64

	// Self
	SlotNum     uint64
	StoragePath string
	IpAddr      string
	RpcPort     uint64
}

var DefaultConfig Config

func init() {
	DefaultConfig = Config{
		SlotNum: 1000,
	}
}
