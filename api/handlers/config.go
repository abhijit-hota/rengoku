package handlers

import (
	"net/http"

	"github.com/abhijit-hota/rengoku/server/utils"

	"github.com/gin-gonic/gin"
)

func GetConfig(ctx *gin.Context) {
	cfg := utils.GetConfig()
	ctx.JSON(http.StatusOK, cfg)
}

func ToggleSaveOffline(ctx *gin.Context) {
	cfg := utils.GetConfig()
	cfg.ShouldSaveOffline = !cfg.ShouldSaveOffline
	utils.UpdateConfigFile(cfg)

	ctx.JSON(http.StatusOK, cfg)
}

func UpdateURLActions(ctx *gin.Context) {
	var urlAction utils.URLAction

	if err := ctx.ShouldBindJSON(&urlAction); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := utils.GetConfig()

	switch ctx.Request.Method {
	case "POST":
		if urlAction.MatchDetection == "" {
			urlAction.MatchDetection = "starts_with"
		}
		if len(urlAction.Tags) == 0 {
			urlAction.Tags = []int64{}
		}
		cfg.URLActions = append(cfg.URLActions, urlAction)
		utils.UpdateConfigFile(cfg)
		ctx.JSON(http.StatusOK, gin.H{"success": true})
		return

	case "PUT":
		// TODO: Shift to merging logic
		for index, rule := range cfg.URLActions {
			if rule.Pattern == urlAction.Pattern {
				cfg.URLActions[index] = urlAction
				utils.UpdateConfigFile(cfg)
				ctx.JSON(http.StatusOK, gin.H{"success": true})
				return
			}
		}

	case "DELETE":
		for index, rule := range cfg.URLActions {
			if rule.Pattern == urlAction.Pattern {
				cfg.URLActions = utils.RemoveIndex(cfg.URLActions, index)
				utils.UpdateConfigFile(cfg)
				ctx.JSON(http.StatusOK, gin.H{"success": true})
				return
			}
		}
	}
	ctx.JSON(http.StatusNotModified, gin.H{"message": "No change"})
	return
}
