package worker

type Config struct {
	// 存储附件的路径
	AttachmentStoragePath string
}

var DefaultConfig Config

func init() {
	DefaultConfig = Config{
		AttachmentStoragePath: "./satweave-data/attachment/",
	}
}
