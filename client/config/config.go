package config

const (
	ConnectEdge = iota
	ConnectCloud
)

type ClientConfig struct {
	NodeAddr    string
	NodePort    uint64
	CloudAddr   string
	CloudPort   uint64
	ConnectType int
}

var DefaultConfig ClientConfig

func init() {
	DefaultConfig = ClientConfig{
		NodeAddr:    "satweave-sat-dev.satweave.svc.cluster.local",
		NodePort:    3267,
		ConnectType: ConnectEdge,
	}
}
