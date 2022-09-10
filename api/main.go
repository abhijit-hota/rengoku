package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/abhijit-hota/rengoku/server/db"
	"github.com/abhijit-hota/rengoku/server/utils"
	"github.com/shirou/gopsutil/v3/process"

	"github.com/joho/godotenv"
)

//go:embed frontend-dist
var distFolder embed.FS

var initCmd = flag.NewFlagSet("init", flag.ExitOnError)
var overwrite = flag.Bool("overwrite", false, "Pass this to reinitialize the app.\n WARNING: You will lose all your data. Run `maby export` if you want to back up.")

func RunSubcommand() {
	rengokuPath := GetRengokuPath()
	switch os.Args[1] {
	case "init":
		initCmd.Parse(os.Args[2:])

		_, err := os.Stat(rengokuPath + rengokuInit)
		firstRun := os.IsNotExist(err)

		if !firstRun && !(*overwrite) {
			return
		}

		Instantiate(rengokuPath)
	case "start":

		cmd := exec.Cmd{
			Path: os.Args[0],
		}
		err := cmd.Start()
		utils.Must(err)

		proc, err := process.NewProcess(int32(cmd.Process.Pid))
		utils.Must(err)

		if status, err := proc.Status(); err != nil || status[0] == "zombie" {
			log.Fatal("Couldn't start server.")
		}

		err = os.WriteFile(rengokuPath+"maby.pid", []byte(fmt.Sprint(cmd.Process.Pid)), 0644)
		utils.Must(err)

		return
	case "stop":
		pid, err := os.ReadFile(rengokuPath + "maby.pid")
		if os.IsNotExist(err) {
			log.Fatal("It's not even started brah")
		}

		pidNo := int32(utils.MustGet(strconv.Atoi(string(pid))))
		mabyProc, err := process.NewProcess(pidNo)
		if err != nil {
			log.Fatal("You killed it yourself didn't you?")
		}

		mabyProc.Terminate()
		os.Remove(rengokuPath + "maby.pid")
		fmt.Println("Closed.")

		return
	case "export":
	case "auth":

	}
}

func main() {

	if len(os.Args) == 2 {
		RunSubcommand()
		return
	}

	rengokuPath := GetRengokuPath()

	// TODO: Allow to config this during initialize

	/*
		Load the .env file.
		It contains the runtime configuration for the application like database path,
		user settings path, where to save offline pages, port etc.
		These settings are usually not needed more than once and are not the application features.
		Hence, these are "hidden" from the users.
	*/
	godotenv.Load(rengokuPath + rengokuEnv)

	/*
		Load the user settings
		It contains settings which should be shown to users.
		These are the feature settings of the application.
	*/
	utils.LoadConfig()

	// Initialize the database. Connect and create tables if not made.
	db.InitializeDB()

	// Create and run the API/Static server
	err := CreateServer().Run(":" + os.Getenv("PORT"))
	utils.Must(err)
}
