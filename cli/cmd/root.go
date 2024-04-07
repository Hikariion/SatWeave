package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"satweave/client/config"
	configUtil "satweave/utils/config"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "satweave-client",
	Short: "CLI for satweave-client",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.AddCommand(clusterCmd)
	clusterCmd.AddCommand(clusterDescribeCmd)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func readConfig(confPath string) {
	conf := config.DefaultConfig
	configUtil.Register(&conf, confPath)
	configUtil.ReadAll()
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.client.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringP("config", "c", "./client.json",
		"config file path")
	path, err := rootCmd.PersistentFlags().GetString("config")
	if err == nil {
		readConfig(path)
	}
}
