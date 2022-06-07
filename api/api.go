package main

import (
	"bufio"
	"net/http"

	"github.com/abhijit-hota/rengoku/server/handlers"
	"github.com/abhijit-hota/rengoku/server/utils"
	"github.com/gin-gonic/gin"
)

func CreateServer() *gin.Engine {
	router := gin.Default()

	router.StaticFS("assets", http.FS(assets))
	router.GET("/", func(ctx *gin.Context) {
		html := utils.MustGet(Assets.Open("dist/index.html"))
		defer html.Close()
		rd := bufio.NewReader(html)
		rd.WriteTo(ctx.Writer)
	})
	router.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/")
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
		tagRouter.POST("/bulk", handlers.CreateBulkTags)
		tagRouter.POST("", handlers.CreateTag)
		tagRouter.GET("", handlers.GetAllTags)
		tagRouter.PATCH("/:id", handlers.UpdateTagName)
		tagRouter.DELETE("/:id", handlers.DeleteTag)
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
