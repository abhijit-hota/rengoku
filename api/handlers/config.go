package handlers

import (
	"api/utils"
	"fmt"
	"net/http"

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
			urlAction.Tags = []int{}
		}
		fmt.Printf("%+v\n", urlAction)
		cfg.URLActions = append(cfg.URLActions, urlAction)
		break
	case "PUT":
		// TODO: Shift to merging logic
		for index, rule := range cfg.URLActions {
			if rule.Pattern == urlAction.Pattern {
				cfg.URLActions[index] = urlAction
				break
			}
		}
		ctx.JSON(http.StatusNotModified, gin.H{"message": "No change"})
		return
	case "DELETE":
		for index, rule := range cfg.URLActions {
			if rule.Pattern == urlAction.Pattern {
				cfg.URLActions = utils.RemoveIndex(cfg.URLActions, index)
			}
		}
		ctx.JSON(http.StatusNotModified, gin.H{"message": "No change"})
		return
	}

	utils.UpdateConfigFile(cfg)

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
