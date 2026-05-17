package config

import (
	"log/slog"
	"os"
	"strconv"
)

var Conf *BotConfig

func LoadConfig() error {
	envFiles := []string{".env.local", ".env"}
	if err := loadEnvFiles(envFiles...); err != nil {
		slog.Info("Warning loading env files", "error", err)
	}

	Conf = &BotConfig{
		ApiId:          getEnvInt32("API_ID", 0),
		ApiHash:        os.Getenv("API_HASH"),
		Token:          os.Getenv("TOKEN"),
		SessionStrings: getSessionStrings("STRING", 10),
		SessionType:    getEnvStr("SESSION_TYPE", "gogram"),
		DatabaseUrl:    os.Getenv("DATABASE_URL"),
		ApiUrl:         getEnvStr("API_URL", "https://beta.fallenapi.fun"),
		ApiKey:         os.Getenv("API_KEY"),
		OwnerId:        getEnvInt64("OWNER_ID"),
		Proxy:          os.Getenv("PROXY"),
		MaxFileSize:    getEnvInt64("MAX_FILE_SIZE"),
		DownloadsDir:   "downloads",
		StartImg:       getEnvStr("START_IMG", "https://v1.pinimg.com/videos/iht/expMp4/62/16/5a/62165a0f66af6a0db73b8a7763e8cf30_720w.mp4"),
		Port:           getEnvStr("PORT", "6060"),
		AutoLeave:      getEnvBool("AUTO_LEAVE", false),
	}

	if Conf.OwnerId != 0 && !containsInt(Conf.DEVS, Conf.OwnerId) {
		Conf.DEVS = append(Conf.DEVS, Conf.OwnerId)
	}

	if err := Conf.validate(); err != nil {
		return err
	}

	if err := os.MkdirAll(Conf.DownloadsDir, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll("src/cookies", 0750); err != nil {
		return err
	}

	return nil
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	if val, err := strconv.ParseBool(value); err == nil {
		return val
	}
	return defaultValue
}
