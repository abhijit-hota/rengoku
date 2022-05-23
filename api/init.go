package main

import (
	"fmt"
	"os"

	"github.com/abhijit-hota/rengoku/server/utils"
)

const (
	rengokuDir   = ".rengoku/"
	rengokuDBDir = "db/"

	rengokuOfflineDir = "saved_pages/"
	rengokuEnv        = "rengoku.env"
	rengokuConfig     = "rengoku.config.json"
	rengokuDBName     = "rengoku.db"

	rengokuInit = ".init_do_not_delete"
	rengokuPort = 8080
	CREATE_NEW  = os.O_CREATE | os.O_RDWR | os.O_APPEND | os.O_TRUNC
)

func GetRengokuPath() string {
	homeDir := utils.MustGet(os.UserHomeDir()) + "/"
	rengokuPath := homeDir + rengokuDir
	return rengokuPath
}

func Instantiate(rengokuPath string) {
	err := os.MkdirAll(rengokuPath, os.ModePerm)

	// Create .init file
	initFile, err := os.OpenFile(rengokuPath+rengokuInit, CREATE_NEW, os.ModePerm)
	utils.Must(err)
	defer initFile.Close()
	fmt.Fprintln(initFile, "DO NOT DELETE THIS. DOING SO WILL CAUSE ALL YOUR CONFIGURATION TO BE RESET, WHICH MAY INCLUDE LOSS OF DATA.")

	// Create .env file and fill default values
	envFile, err := os.OpenFile(rengokuPath+rengokuEnv, CREATE_NEW, os.ModePerm)
	utils.Must(err)
	defer envFile.Close()
	fmt.Fprintf(
		envFile,
		"DB_PATH=%s\nCONFIG_PATH=%s\nSAVE_OFFLINE_PATH=%s\nPORT=%v",
		rengokuPath+rengokuDBDir+rengokuDBName, rengokuPath+rengokuConfig, rengokuPath+rengokuOfflineDir, rengokuPort,
	)

	// Create config file and fill default values
	configFile, err := os.OpenFile(rengokuPath+rengokuConfig, CREATE_NEW, os.ModePerm)
	utils.Must(err)
	defer configFile.Close()
	fmt.Fprintln(configFile, "{}")

	// Create default db dir
	err = os.MkdirAll(rengokuPath+rengokuDBDir, os.ModePerm)
	utils.Must(err)

	// Create default directory for saving websites offline
	err = os.MkdirAll(rengokuPath+rengokuOfflineDir, os.ModePerm)
	utils.Must(err)
}
