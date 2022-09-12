package utils

import (
	"encoding/json"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type URLAction struct {
	Pattern           string  `json:"pattern" binding:"required"`
	MatchDetection    string  `json:"matchDetection" binding:"required"`
	ShouldSaveOffline bool    `json:"shouldSaveOffline"`
	Tags              []int64 `json:"tags"`
	Folders           []int64 `json:"folders"`
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
		domain := levels[len(levels)-2] + "." + levels[len(levels)-1]
		return domain == u.Pattern
	case "starts_with":
		fallthrough
	default:
		return strings.HasPrefix(urlStr, u.Pattern)
	}
}

type Config struct {
	ShouldSaveOffline bool        `json:"shouldSaveOffline"`
	URLActions        []URLAction `json:"urlActions,omitempty"`
}

var config *Config
var configPath string

func LoadConfig() {
	configPath = os.Getenv("CONFIG_PATH")
	file := MustGet(os.OpenFile(configPath, os.O_RDONLY, os.ModePerm))
	defer file.Close()

	Must(json.NewDecoder(file).Decode(&config))
}

func UpdateConfigFile(updatedConfig Config) {
	file, err := os.OpenFile(configPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_SYNC, os.ModePerm)
	Must(err)
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(updatedConfig)
	Must(err)

	// Load config again after updating file
	*config = updatedConfig
}

func GetConfig() Config {
	if config == nil {
		LoadConfig()
	}

	return *config
}
