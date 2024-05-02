package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/keidarcy/e1s/internal/utils"
	e1s "github.com/keidarcy/e1s/internal/view"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFile string

func initConfig() {
	useFlag := true
	if configFile == "" {
		useFlag = false
		home, err := os.UserHomeDir()
		if err != nil {
			home = utils.EmptyText
		}
		configFile = filepath.Join(home, ".config", "e1s", "config.yml")
	}

	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			if useFlag {
				fmt.Println("Error reading config file:", err)
				os.Exit(1)
			}
			configFile = utils.EmptyText
		} else {
			fmt.Println("Error reading config file:", err)
		}
	}

}

func init() {
	cobra.OnInitialize(initConfig)

	defaultLogFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s.log", utils.AppName))

	rootCmd.Flags().StringVarP(&configFile, "config-file", "c", "", "config file (default \"$HOME/.config/e1s/config.yml\")")
	rootCmd.Flags().BoolP("debug", "d", false, "sets debug mode")
	rootCmd.Flags().BoolP("json", "j", false, "log output json format")
	rootCmd.Flags().Bool("read-only", false, "sets read only mode")
	rootCmd.Flags().StringP("log-file", "l", defaultLogFile, "specify the log file path")
	rootCmd.Flags().StringP("shell", "s", "/bin/sh", "specify interactive ecs exec shell")
	rootCmd.Flags().IntP("refresh", "r", 30, "specify the default refresh rate as an integer, sets -1 to stop auto refresh (sec)")
	rootCmd.Flags().String("profile", "", "specify the AWS profile")
	rootCmd.Flags().String("region", "", "specify the AWS region")

	viper.BindPFlags(rootCmd.Flags())
}

var rootCmd = &cobra.Command{
	Use:   "e1s",
	Short: "e1s - Easily Manage AWS ECS Resources in Terminal 🐱",
	Long: `e1s is a terminal application to easily browse and manage AWS ECS resources 🐱. 
Check https://github.com/keidarcy/e1s for more details.`,
	Run: func(cmd *cobra.Command, args []string) {
		profile := viper.GetString("profile")
		region := viper.GetString("region")
		if profile != "" {
			os.Setenv("AWS_PROFILE", profile)
			defer func() {
				os.Unsetenv("AWS_PROFILE")
			}()
		}
		if region != "" {
			os.Setenv("AWS_REGION", region)
			defer func() {
				os.Unsetenv("AWS_REGION")
			}()
		}

		logFile := viper.GetString("log-file")
		json := viper.GetBool("json")
		debug := viper.GetBool("debug")
		readOnly := viper.GetBool("read-only")
		refresh := viper.GetInt("refresh")
		shell := viper.GetString("shell")

		logger, file := utils.GetLogger(logFile, json, debug)
		defer file.Close()

		option := e1s.Option{
			Logger:     logger,
			ConfigFile: configFile,
			LogFile:    logFile,
			Debug:      debug,
			JSON:       json,
			ReadOnly:   readOnly,
			Refresh:    refresh,
			Shell:      shell,
		}

		logger.Debugf("ConfigFile: %s, LogFile: %s, Debug: %t, JSON: %t, ReadOnly: %t, Refresh: %d, Shell: %s", configFile, logFile, debug, json, readOnly, refresh, shell)

		if err := e1s.Start(option); err != nil {
			fmt.Printf("e1s failed to start, please check your aws cli credential and permission. error: %v\n", err)
			logger.Fatalf("Failed to start, error: %v\n", err)
			// will call os.Exit(1)
		}
	},
	Version: utils.ShowVersion(),
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
