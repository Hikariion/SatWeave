package worker

type Config struct {
	// 存储附件的路径
	AttachmentStoragePath string `json:"AttachmentStoragePath"`
	// 文件输出的路径
	OutputPath string `json:"OutputPath"`
}

var DefaultConfig Config

func init() {
	DefaultConfig = Config{
		AttachmentStoragePath: "./satweave-data/attachment/",
		OutputPath:            "./satweave-data/output/",
	}
}
