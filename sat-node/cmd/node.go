package cmd

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"satweave/messenger"
	"satweave/sat-node/config"
	"satweave/sat-node/infos"
	"satweave/sat-node/moon"
	"satweave/sat-node/watcher"
	"satweave/utils/logger"
	"strconv"
	"syscall"
	"time"
)

var nodeCmd = &cobra.Command{
	Use:   "node {run}",
	Short: "satellite node operate",
}

var nodeRunCmd = &cobra.Command{
	Use:   "run",
	Short: "start run satellite node",
}

func init() {
	nodeRunCmd.Flags().StringP("sunAddr", "s", "", "ground station addr")
}

func nodeRun(cmd *cobra.Command, _ []string) {
	// read config
	confPath := cmd.Flag("config").Value.String()
	conf := config.DefaultConfig
	conf.WatcherConfig.SunAddr = cmd.Flag("sunAddr").Value.String()

	// open config file
	configFile, err := os.Open(confPath)
	if err != nil {
		logger.Errorf("open config file fail: %v", err)
	}
	defer configFile.Close()

	data, err := ioutil.ReadAll(configFile)
	if err != nil {
		logger.Errorf("read config file fail: %v", err)
	}

	if err := json.Unmarshal(data, &conf); err != nil {
		logger.Errorf("unmarshal config file fail: %v", err)
	}
	// TODO(qiu): read history config

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Print config
	logger.Infof("sat node config: %v", conf)

	// Gen Rpc
	rpc := messenger.NewRpcServer(conf.WatcherConfig.SelfNodeInfo.RpcPort)

	// Gen Moon
	// TODO(qiu): change to stable memory (maybe useless?)
	storageFactory := infos.NewMemoryInfoFactory()
	storageRegisterBuilder := infos.NewStorageRegisterBuilder(storageFactory)
	storageRegister := storageRegisterBuilder.GetStorageRegister()
	m := moon.NewMoon(ctx, &conf.WatcherConfig.SelfNodeInfo, &conf.MoonConfig, rpc, storageRegister)

	// Gen Watcher
	w := watcher.NewWatcher(ctx, &conf.WatcherConfig, rpc, m, storageRegister)

	// Run
	go func() {
		err := rpc.Run()
		if err != nil {
			logger.Errorf("Run rpcServer err: %v", err)
		}
	}()

	go w.Run()

	// 监听系统信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range c {
			logger.Infof("receive signal from os: %v, sat node start stop", s)
			cancel()
			time.Sleep(time.Second)
			os.Exit(0)
		}
	}()

	// init Gin
	router := newRouter(w)
	port := strconv.FormatUint(conf.HttpPort, 10)
	_ = router.Run(":" + port)
}

func newRouter(w *watcher.Watcher) *gin.Engine {
	router := gin.Default()
	router.GET("/", getClusterInfo(w))
	return router
}

func getClusterInfo(w *watcher.Watcher) func(c *gin.Context) {
	return func(c *gin.Context) {
		if w == nil {
			c.String(http.StatusBadGateway, "Sat Node Not Ready")
		}
		clusterInfo := w.GetCurrentClusterInfo()
		info := clusterInfo.GetNodesInfo()
		c.JSONP(http.StatusOK, info)
	}
}
