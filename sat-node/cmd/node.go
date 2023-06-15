package cmd

import (
	"github.com/spf13/cobra"
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

//func nodeRun(cmd *cobra.Command, _ []string)  {
//	// read config
//	confPath := cmd.Flag("config").Value.String()
//	conf := config.DefaultConfig
//}
