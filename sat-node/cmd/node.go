package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"satweave/messenger"
	"satweave/sat-node/config"
	"satweave/sat-node/worker"
	configUtil "satweave/utils/config"
	"satweave/utils/logger"
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
	// TODO(qiu): 这个 config 是哪来的？
	confPath := cmd.Flag("config").Value.String()
	conf := config.DefaultConfig
	// TODO(qiu): 仔细看看这是怎么用的
	configUtil.Register(&conf, confPath)
	configUtil.ReadAll()

	// TODO(qiu): 这是什么
	// read history config
	_ = configUtil.GetConf(&conf)
	err := config.InitConfig(&conf)
	if err != nil {
		logger.Errorf("init config fail: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Gen Rpc
	rpc := messenger.NewRpcServer(conf.RpcPort)

	// Gen Worker
	wk := worker.NewWorker(ctx, rpc, &conf.WorkConfig)

	// Run
	go func() {
		err := rpc.Run()
		if err != nil {
			logger.Errorf("RPC server run error: %v", err)
		}
	}()
	go wk.Run()
}
