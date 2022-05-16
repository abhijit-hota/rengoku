package main

import (
	"api/db"
	"api/handlers"
	"api/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	godotenv.Load()
	utils.LoadConfig()
	db.InitializeDB()

	router := gin.Default()
	router.Use(CORSMiddleware())
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

		bookmarkRouter.PUT("/:id/tag", handlers.AddBookmarkTag)
		bookmarkRouter.PUT("/tags", handlers.BulkAddBookmarkTags)
		bookmarkRouter.DELETE("/:id/tag/:tagId", handlers.DeleteBookmarkTag)
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
	router.Run()
}
