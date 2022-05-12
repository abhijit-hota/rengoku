package main

import (
	"bingo/api/db"
	"bingo/api/handlers"
	"bingo/api/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	utils.LoadConfig()
	db.InitializeDB()

	router := gin.Default()

	bookmarkRouter := router.Group("/bookmarks")
	{
		bookmarkRouter.POST("", handlers.AddBookmark)
		bookmarkRouter.GET("", handlers.GetBookmarks)
		// bookmarkRouter.PATCH("/:id", handlers.UpdateBookmark)
		// bookmarkRouter.DELETE("/", handlers.DeleteBookmark)
	}
	tagRouter := router.Group("/tags")
	{
		// tagRouter.POST("/", handlers.AddTag)
		// tagRouter.GET("/", handlers.GetAllTags)
		// tagRouter.PATCH("/:id", handlers.UpdateTag)
		// tagRouter.DELETE("/:id", handlers.DeleteTag)
		tagRouter.GET("/tree", handlers.GetLinkTree)
	}

	configRouter := router.Group("/config")
	{
		configRouter.GET("", handlers.GetConfig)
		configRouter.PATCH("/save-offline", handlers.ToggleSaveOffline)

		autotagRouter := configRouter.Group("/autotag-rule")
		{
			autotagRouter.POST("", handlers.UpdateAutotagRules)
			autotagRouter.PUT("", handlers.UpdateAutotagRules)
			autotagRouter.DELETE("", handlers.UpdateAutotagRules)
		}
	}
	router.Run()
}
