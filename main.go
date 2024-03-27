package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/keidarcy/e1s/ui"
	"github.com/keidarcy/e1s/util"
)

var (
	readOnly    bool
	staleData   bool
	debug       bool
	logFilePath string
)

func init() {
	defaultLogFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.log", util.AppName))

	rootCmd.Flags().BoolVarP(&readOnly, "readonly", "r", false, "Enable read only mode")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode")
	rootCmd.Flags().BoolVar(&staleData, "stale-data", false, "Enable stale data mode(only refetch data when hit ctrl + r)")
	rootCmd.Flags().StringVar(&logFilePath, "log-file-path", defaultLogFilePath, "Custom e1s log file path")
}

var rootCmd = &cobra.Command{
	Use:   "e1s",
	Short: "E1S - Easily Manage AWS ECS Resources in Terminal 🐱",
	Long: `E1S is a terminal application to easily browse and manage AWS ECS resources, with a focus on Fargate. 
Inspired by k9s.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger, file := util.GetLogger(logFilePath, debug)
		defer file.Close()

		option := ui.Option{
			StaleData: staleData,
			ReadOnly:  readOnly,
			Logger:    logger,
		}

		if err := ui.Show(option); err != nil {
			fmt.Printf("e1s failed to start, please check your aws cli credential and permission. error: %v\n", err)
			logger.Fatalf("Failed to start, error: %v\n", err) // will call os.Exit(1)
		}
	},
	Version: fmt.Sprintf("v%s", util.AppVersion),
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
