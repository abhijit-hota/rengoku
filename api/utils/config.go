package utils

import (
	"bufio"
	"os"

	"github.com/BurntSushi/toml"
)

type AutoTagRule struct {
	Pattern string `toml:"url_pattern" json:"pattern" binding:"required"`
	Tags    []int  `toml:"tags" json:"tags" binding:"required"`
}
type Config struct {
	ShouldSaveOffline bool          `toml:"should_save_offline" json:"shouldSaveOffline"`
	AutoTagRules      []AutoTagRule `toml:"autotag_rule" json:"autotagRules"`
}

var config *Config
var configPath string

func LoadConfig() {
	configPath = os.Getenv("CONFIG_PATH")
	_, err := toml.DecodeFile(configPath, &config)
	Must(err)
}

func UpdateConfigFile(updatedConfig Config) {
	file, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_SYNC, os.ModePerm)
	Must(err)
	defer file.Close()

	writer := bufio.NewWriter(file)
	err = toml.NewEncoder(writer).Encode(updatedConfig)
	Must(err)

	// Load config again after updating file
	LoadConfig()
}

func GetConfig() Config {
	if config == nil {
		LoadConfig()
	}
	return *config
}
