package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"satweave/utils/logger"
)

var clusterCmd = &cobra.Command{
	Use:   "cluster {describe}",
	Short: "Manage clusters",
}

var clusterDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe satweave cluster",
	Run: func(cmd *cobra.Command, args []string) {
		clusterDescribe()
	},
}

func clusterDescribe() {
	client := getClient()
	state, err := client.GetClusterOperator().State()
	if err != nil {
		logger.Errorf("Failed to get cluster state: %v", err)
		os.Exit(1)
	}
	fmt.Println(state)
}
