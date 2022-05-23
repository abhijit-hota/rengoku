package main

import (
	"net/http"

	"github.com/abhijit-hota/rengoku/server/handlers"
	"github.com/gin-gonic/gin"
)

func CreateServer() *gin.Engine {
	router := gin.Default()
	router.Static("css", "views/css")
	router.Static("js", "views/js")
	router.LoadHTMLGlob("views/html/*.html")

	router.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", nil)
	})

	apiRouter := router.Group("/api")
	bookmarkRouter := apiRouter.Group("/bookmarks")
	{
		bookmarkRouter.POST("", handlers.AddBookmark)
		bookmarkRouter.GET("", handlers.GetBookmarks)

		bookmarkRouter.PUT("/:id/:property", handlers.AddBookmarkProperty)
		bookmarkRouter.DELETE("/:id/:property/:propertyId", handlers.DeleteBookmarkProperty)
		bookmarkRouter.PUT("/tags", handlers.BulkAddBookmarkTags)
		bookmarkRouter.DELETE("/:id", handlers.DeleteBookmark)
		bookmarkRouter.DELETE("", handlers.BulkDeleteBookmarks)
	}
	tagRouter := apiRouter.Group("/tags")
	{
		tagRouter.POST("", handlers.CreateTag)
		tagRouter.GET("", handlers.GetAllTags)
		tagRouter.PATCH("/:id", handlers.UpdateTagName)
		tagRouter.DELETE("/:id", handlers.DeleteTag)
		tagRouter.GET("/tree", handlers.GetLinkTree)
	}

	folderRouter := apiRouter.Group("/folders")
	{
		folderRouter.POST("", handlers.CreateFolder)
		folderRouter.GET("", handlers.GetRootFolders)
		folderRouter.PATCH("/:id", handlers.UpdateFolderName)
		folderRouter.DELETE("/:id", handlers.DeleteFolder)
		folderRouter.GET("/tree", handlers.GetLinkTree)
	}

	configRouter := apiRouter.Group("/config")
	{
		configRouter.GET("", handlers.GetConfig)
		configRouter.PATCH("/save-offline", handlers.ToggleSaveOffline)

		urlActionRouter := configRouter.Group("/url-action")
		{
			urlActionRouter.POST("", handlers.UpdateURLActions)
			urlActionRouter.PUT("", handlers.UpdateURLActions)
			urlActionRouter.DELETE("", handlers.UpdateURLActions)
		}
	}
	return router
}
