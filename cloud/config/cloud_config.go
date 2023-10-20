package config

type CloudConfig struct {
	BasePath string
}

var DefaultCloudConfig CloudConfig

func init() {
	DefaultCloudConfig = CloudConfig{BasePath: "./sat-data/cloud/"}
}
