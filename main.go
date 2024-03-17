package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/keidarcy/e1s/ui"
	"github.com/keidarcy/e1s/util"
)

var (
	showVersion = false
	readOnly    = false
	staleData   = false
	debug       = false
)

func main() {
	defaultLogFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.log", util.AppName))

	flag.BoolVar(&showVersion, "version", false, "Print e1s version")
	flag.BoolVar(&readOnly, "readonly", false, "Enable read only mode")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.BoolVar(&staleData, "stale-data", false, "Enable stale data mode(only refetch data when hit ctrl + r)")
	logFilePath := flag.String("log-file-path", defaultLogFilePath, "Custom e1s log file path")
	flag.Parse()

	if showVersion {
		fmt.Println("v" + util.AppVersion)
		os.Exit(0)
	}

	logger, file := util.GetLogger(*logFilePath, debug)
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
}
