package utils

import (
	"bufio"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

type URLAction struct {
	Pattern           string `toml:"pattern" json:"pattern" binding:"required"`
	MatchDetection    string `toml:"match_detection" json:"matchDetection"`
	ShouldSaveOffline bool   `toml:"should_save_offline" json:"shouldSaveOffline,omitempty"`
	Tags              []int  `toml:"tags" json:"tags"`
}

func (u URLAction) Match(urlStr string) bool {
	switch u.MatchDetection {
	case "regex":
		matched, err := regexp.MatchString(u.Pattern, urlStr)
		Must(err)
		return matched
	case "origin":
		parsed, err := url.Parse(urlStr)
		Must(err)
		return parsed.Hostname() == u.Pattern
	case "domain":
		parsed, err := url.Parse(urlStr)
		Must(err)
		hostname := parsed.Hostname()
		levels := strings.Split(hostname, ".")
		domain := levels[len(levels)-1] + "." + levels[len(levels)-2]
		return domain == u.Pattern
	case "starts_with":
		fallthrough
	default:
		return strings.HasPrefix(urlStr, u.Pattern)
	}
}

type Config struct {
	ShouldSaveOffline bool        `toml:"should_save_offline" json:"shouldSaveOffline"`
	URLActions        []URLAction `toml:"url_action" json:"urlActions"`
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
