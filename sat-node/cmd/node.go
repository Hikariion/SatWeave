package cmd

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"net/http"
	"satweave/messenger"
	"satweave/sat-node/config"
	"satweave/sat-node/worker"
	configUtil "satweave/utils/config"
	"satweave/utils/logger"
	"strconv"
)

var nodeCmd = &cobra.Command{
	Use:   "node {run}",
	Short: "sat not operate",
}

var nodeRunCmd = &cobra.Command{
	Use:   "run",
	Short: "start run sat node",
	Run:   nodeRun,
}

func init() {

}

func nodeRun(cmd *cobra.Command, _ []string) {
	// read config
	confPath := cmd.Flag("config").Value.String()
	conf := config.DefaultConfig

	err := configUtil.TransferJsonToConfig(&conf, confPath)
	if err != nil {
		logger.Errorf("fail to load config, err: %v", err)
	}

	logger.Infof("config: %v", conf)

	// TODO(qiu): 这是什么
	//// read history config
	//_ = configUtil.GetConf(&conf)
	//err := config.InitConfig(&conf)
	//if err != nil {
	//	logger.Errorf("init config fail: %v", err)
	//}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Gen Rpc
	rpc := messenger.NewRpcServer(conf.RpcPort)

	// Gen Worker
	logger.Infof("begin to gen worker")
	wk := worker.NewWorker(ctx, rpc, &conf.WorkerConfig)

	// Run rpc
	logger.Infof("begin to run rpc")
	go func() {
		err := rpc.Run()
		if err != nil {
			logger.Errorf("RPC server run error: %v", err)
		}
	}()

	// Run worker
	logger.Infof("begin to run worker")
	go wk.Run()

	router := gin.Default()
	port := strconv.FormatUint(conf.HttpPort, 10)

	// 返回一个健康状态
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "Healthy",
		})
	})

	_ = router.Run(":" + port)
}

func health() {

}
