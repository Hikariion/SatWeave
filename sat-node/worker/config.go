package worker

type Config struct {
	// 存储附件的路径
	AttachmentStoragePath string
	// 文件输出的路径
	OutputPath string
}

var DefaultConfig Config

func init() {
	DefaultConfig = Config{
		AttachmentStoragePath: "./satweave-data/attachment/",
		OutputPath:            "./satweave-data/output/",
	}
}
