package main

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/keidarcy/e1s/internal/utils"
    e1s "github.com/keidarcy/e1s/internal/view"
    "github.com/spf13/cobra"
)

var (
    readOnly    bool
    debug       bool
    logFilePath string
    json        bool
    refresh     int
    shell       string
    profile     string
)

func init() {
    defaultLogFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.log", utils.AppName))

    rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "sets debug mode")
    rootCmd.Flags().BoolVarP(&json, "json", "j", false, "log output json format")
    rootCmd.Flags().StringVarP(&logFilePath, "log-file-path", "l", defaultLogFilePath, "specify the log file path")
    rootCmd.Flags().BoolVar(&readOnly, "readonly", false, "sets readOnly mode")
    rootCmd.Flags().IntVarP(&refresh, "refresh", "r", 30, "specify the default refresh rate as an integer (sec) (default 30, set -1 to stop auto refresh)")
    rootCmd.Flags().StringVarP(&shell, "shell", "s", "/bin/sh", "specify ecs exec ssh shell")
    rootCmd.Flags().StringVarP(&profile, "profile", "p", "", "specify the AWS profile")
}

var rootCmd = &cobra.Command{
    Use:   "e1s",
    Short: "E1s - Easily Manage AWS ECS Resources in Terminal 🐱",
    Long: `E1s is a terminal application to easily browse and manage AWS ECS resources 🐱. 
Check https://github.com/keidarcy/e1s for more details.`,
    Run: func(cmd *cobra.Command, args []string) {
        if profile != "" {
            os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
            os.Setenv("AWS_PROFILE", profile)
            defer func() {
                os.Unsetenv("AWS_SDK_LOAD_CONFIG")
                os.Unsetenv("AWS_PROFILE")
            }()
        }

        logger, file := utils.GetLogger(logFilePath, json, debug)
        defer file.Close()

        option := e1s.Option{
            ReadOnly: readOnly,
            Logger:   logger,
            Refresh:  refresh,
            Shell:    shell,
        }

        if err := e1s.Start(option); err != nil {
            fmt.Printf("e1s failed to start, please check your aws cli credential and permission. error: %v\n", err)
            logger.Fatalf("Failed to start, error: %v\n", err) // will call os.Exit(1)
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