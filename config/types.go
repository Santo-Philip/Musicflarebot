package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type BotConfig struct {
	ApiId             int32
	ApiHash           string
	Token             string
	SessionStrings    []string
	SessionType       string
	DatabaseUrl       string
	ApiUrl            string
	ApiKey            string
	OwnerId           int64
	LoggerId          int64
	Proxy             string
	DefaultService    string
	MaxFileSize       int64
	SongDurationLimit int64
	DownloadsDir      string
	SupportGroup      string
	SupportChannel    string
	DEVS              []int64
	CookiesPath       []string
	StartImg          string
	Port              string
	AutoLeave         bool
}

func getSessionStrings(prefix string, max int) []string {
	var sessions []string
	for i := 1; i <= max; i++ {
		key := fmt.Sprintf("%s%d", prefix, i)
		if session := os.Getenv(key); session != "" {
			sessions = append(sessions, session)
		}
	}
	if session := os.Getenv(prefix); session != "" {
		sessions = append(sessions, session)
	}
	return sessions
}

func getEnvStr(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvInt32(key string, defaultValue int32) int32 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	if val, err := strconv.ParseInt(value, 10, 32); err == nil {
		return int32(val)
	}
	return defaultValue
}

func getEnvInt64(key string) int64 {
	value := os.Getenv(key)
	if value == "" {
		return 0
	}
	if val, err := strconv.ParseInt(value, 10, 64); err == nil {
		return val
	}
	return 0
}

func containsInt(slice []int64, val int64) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func (c *BotConfig) validate() error {
	required := []struct {
		name  string
		value string
		check func() bool
	}{
		{"API_ID", fmt.Sprintf("%d", c.ApiId), func() bool { return c.ApiId > 0 }},
		{"API_HASH", c.ApiHash, func() bool { return c.ApiHash != "" }},
		{"TOKEN", c.Token, func() bool { return c.Token != "" }},
		{"DATABASE_URL", c.DatabaseUrl, func() bool { return c.DatabaseUrl != "" }},
		{"OWNER_ID", fmt.Sprintf("%d", c.OwnerId), func() bool { return c.OwnerId > 0 }},
	}

	var missing []string
	for _, req := range required {
		if !req.check() {
			missing = append(missing, req.name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}

	if c.MaxFileSize <= 0 {
		c.MaxFileSize = 500 * 1024 * 1024
	}

	if c.SongDurationLimit <= 0 {
		c.SongDurationLimit = 3600
	}

	if !isValidService(c.DefaultService) {
		c.DefaultService = "youtube"
		slog.Info("Invalid DEFAULT_SERVICE, defaulting to 'youtube'", "Service", c.DefaultService)
	}

	return nil
}

func isValidService(service string) bool {
	validServices := map[string]bool{
		"youtube": true,
		"spotify": true,
	}
	return validServices[strings.ToLower(service)]
}
