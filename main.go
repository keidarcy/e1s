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
	json        bool
	refresh     int
)

func init() {
	defaultLogFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.log", util.AppName))

	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "sets debug mode")
	rootCmd.Flags().BoolVarP(&json, "json", "j", false, "log output json format")
	rootCmd.Flags().StringVarP(&logFilePath, "log-file-path", "l", defaultLogFilePath, "specify the log file path")
	rootCmd.Flags().BoolVarP(&staleData, "stale-data", "s", false, "sets stale data mode (only refetch data when hit ctrl + r)")
	rootCmd.Flags().BoolVar(&readOnly, "readonly", false, "sets readOnly mode")
	rootCmd.Flags().IntVarP(&refresh, "refresh", "r", 0, "specify the default refresh rate as an integer (sec) (default 0 no refresh)")
}

var rootCmd = &cobra.Command{
	Use:   "e1s",
	Short: "E1s - Easily Manage AWS ECS Resources in Terminal 🐱",
	Long: `E1s is a terminal application to easily browse and manage AWS ECS resources 🐱. 
Check https://github.com/keidarcy/e1s for more details.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if refresh < 0 {
			return fmt.Errorf("refresh can not be negative %d", refresh)
		}
		if staleData && refresh > 0 {
			return fmt.Errorf("stale data only refetch data when manually hit ctrl + r but refresh=%d refetch data every %d second(s)", refresh, refresh)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		logger, file := util.GetLogger(logFilePath, json, debug)
		defer file.Close()

		option := ui.Option{
			StaleData: staleData,
			ReadOnly:  readOnly,
			Logger:    logger,
			Refresh:   refresh,
		}

		if err := ui.Start(option); err != nil {
			fmt.Printf("e1s failed to start, please check your aws cli credential and permission. error: %v\n", err)
			logger.Fatalf("Failed to start, error: %v\n", err) // will call os.Exit(1)
		}
	},
	Version: util.ShowVersion(),
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
