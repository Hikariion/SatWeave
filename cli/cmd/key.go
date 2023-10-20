package cmd

import (
	"os"
	"satweave/client"
	"satweave/client/config"
	configUtil "satweave/utils/config"
	"satweave/utils/logger"
)

func getClient() *client.Client {
	var conf config.ClientConfig
	_ = configUtil.GetConf(&conf)
	c, err := client.New(&conf)
	if err != nil {
		logger.Errorf("create client fail: %v", err)
		os.Exit(1)
	}
	return c
}
