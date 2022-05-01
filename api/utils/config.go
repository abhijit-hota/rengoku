package utils

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	DatabasePath      string
	ShouldSaveOffline bool
	AutofillURLData   bool
}

var config Config

func LoadConfig() {
	_, err := toml.DecodeFile("./bingo.toml", &config)
	Must(err)
}

func GetConfig() Config {
	if config == (Config{}) {
		LoadConfig()
	}
	return config
}
