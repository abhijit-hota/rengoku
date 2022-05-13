package handlers

import (
	"api/utils"
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

func UpdateAutotagRules(ctx *gin.Context) {
	var autotag utils.AutoTagRule

	if err := ctx.ShouldBindJSON(&autotag); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := utils.GetConfig()

	switch ctx.Request.Method {
	case "POST":
		cfg.AutoTagRules = append(cfg.AutoTagRules, autotag)
		break
	case "PUT":
		for index, rule := range cfg.AutoTagRules {
			if rule.Pattern == autotag.Pattern {
				cfg.AutoTagRules[index] = autotag
				break
			}
		}
		ctx.JSON(http.StatusNotModified, gin.H{"message": "No change"})
		return
	case "DELETE":
		for index, rule := range cfg.AutoTagRules {
			if rule.Pattern == autotag.Pattern {
				cfg.AutoTagRules = utils.RemoveIndex(cfg.AutoTagRules, index)
			}
		}
		ctx.JSON(http.StatusNotModified, gin.H{"message": "No change"})
		return
	}

	utils.UpdateConfigFile(cfg)

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
