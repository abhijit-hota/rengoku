package utils

import (
	"github.com/BurntSushi/toml"
)

type AutoTagRule struct {
	Pattern           string   `toml:"pattern"`
	Tags              []string `toml:"tags"`
	ShouldSaveOffline bool     `toml:"should_save_offline"`
}
type Config struct {
	DatabasePath      string `toml:"database_path"`
	ShouldSaveOffline bool   `toml:"should_save_offline"`

	AutoTagRules []AutoTagRule `toml:"autotag_rule"`
}

const ConfigPath = "./bingo.toml"

var config *Config

func LoadConfig() {
	_, err := toml.DecodeFile(ConfigPath, &config)
	Must(err)
}

func GetConfig() Config {
	if config == nil {
		LoadConfig()
	}
	return *config
}
