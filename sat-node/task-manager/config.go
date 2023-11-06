package task_manager

type Config struct {
	// Slot 数量
	SlotNum     int
	StoragePath string
	SunAddr     string
	CloudAddr   string
	CloudPort   uint64
}

var DefaultConfig Config

func init() {
	DefaultConfig = Config{
		SlotNum: 5,
	}
}
