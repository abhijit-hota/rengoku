package main

import (
	"os"

	"github.com/abhijit-hota/rengoku/server/db"
	"github.com/abhijit-hota/rengoku/server/utils"

	"github.com/joho/godotenv"
)

func main() {
	// Get default rengoku path
	rengokuPath := GetRengokuPath()

	// Check if this is the first run
	if _, err := os.Stat(rengokuPath + rengokuInit); os.IsNotExist(err) {
		// If yes, then instantiate config files and directories
		Instantiate(rengokuPath)
	}

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
	CreateServer().Run(":" + os.Getenv("PORT"))
}
