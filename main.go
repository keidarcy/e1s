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
	version   = false
	readOnly  = false
	staleData = false
)

func main() {
	defaultLogFilePath := filepath.Join(os.Getenv("HOME"), fmt.Sprintf(".%s.log", util.AppName))

	flag.BoolVar(&version, "version", false, "Print e1s version")
	flag.BoolVar(&readOnly, "readonly", false, "Enable readonly mode")
	logFilePath := flag.String("log-file-path", defaultLogFilePath, "The e1s debug log file path")
	flag.BoolVar(&staleData, "stale-data", false, "Only fetch data in the first run(update status when hit ctrl + r)")
	flag.Parse()

	if version {
		fmt.Println("v" + util.AppVersion)
		os.Exit(0)
	}

	logger, logFile := util.GetLogger(*logFilePath)
	defer logFile.Close()

	option := ui.Option{
		StaleData: staleData,
		ReadOnly:  readOnly,
		Logger:    logger,
	}

	if err := ui.Show(option); err != nil {
		logger.Printf("e1s - failed to start, error: %v\n", err)
		fmt.Println("e1s failed to start, please check your aws cli credential or permission.")
		fmt.Println(err)
		os.Exit(1)
	}
}
