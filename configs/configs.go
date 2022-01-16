package config

import (
	"os"
	"strconv"
)

func getEnv(env string, fallback string) string {
	val := os.Getenv(env)
	if len(val) == 0 {
		return fallback
	}
	return val
}

func getBoolEnv(env string, fallback bool) bool {
	val := os.Getenv(env)
	if len(val) == 0 {
		return fallback
	}
	return val == "true"
}

func getIntEnv(env string, fallback int) int {
	val := os.Getenv(env)
	if len(val) == 0 {
		return fallback
	}
	valInt32, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return fallback
	}
	return int(valInt32)
}

var PORT string = getEnv("PORT", "3000")
var SITE_URL string = getEnv("SITE_URL", "https://icanhas.cheezburger.com")
var CARD_IMG_SELECTOR = getEnv("CARD_IMG_SELECTOR", ".mu-post.mu-thumbnail > img")
var MIN_CARDS_PER_PAGE = getIntEnv("MIN_CARDS_PER_PAGE", 10)
var TIMEOUT = getIntEnv("TIMEOUT", 600) // seconds
var DEBUG = getBoolEnv("DEBUG", false)
