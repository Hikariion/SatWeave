package task_manager

type Config struct {
	// Slot 数量
	SlotNum int
}

var DefaultConfig Config

func init() {
	DefaultConfig = Config{
		SlotNum: 5,
	}
}
