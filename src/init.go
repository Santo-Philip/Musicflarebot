package src

import (
	"fmt"
	"log/slog"
	"time"

	"musicflarebot/internal/config"
	"musicflarebot/internal/database"
	"musicflarebot/src/vc"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func loadOwnerSettings() {
	if loggerId := database.GetOwnerLoggerId(); loggerId > 0 {
		config.Conf.LoggerId = loggerId
		slog.Info("[DB] Loaded LoggerId from database", "id", loggerId)
	}
	if group := database.GetOwnerSupportGroup(); group != "" {
		config.Conf.SupportGroup = group
		slog.Info("[DB] Loaded SupportGroup from database")
	}
	if channel := database.GetOwnerSupportChannel(); channel != "" {
		config.Conf.SupportChannel = channel
		slog.Info("[DB] Loaded SupportChannel from database")
	}
	if dur := database.GetOwnerSongDuration(); dur > 0 {
		config.Conf.SongDurationLimit = dur
		slog.Info("[DB] Loaded SongDurationLimit from database", "seconds", dur)
	}
	if svc := database.GetOwnerDefaultService(); svc != "" {
		config.Conf.DefaultService = svc
		slog.Info("[DB] Loaded DefaultService from database", "service", svc)
	}
	if devs := database.GetOwnerDevs(); len(devs) > 0 {
		config.Conf.DEVS = devs
		slog.Info("[DB] Loaded DEVS from database", "count", len(devs))
	}
}

func loadDBSessions() {
	dbSessions, err := database.GetSessionStrings()
	if err != nil {
		slog.Error("[DB] Failed to load session strings", "error", err)
		return
	}

	existing := make(map[string]bool, len(config.Conf.SessionStrings))
	for _, s := range config.Conf.SessionStrings {
		existing[s] = true
	}

	for _, s := range dbSessions {
		if !existing[s] {
			config.Conf.SessionStrings = append(config.Conf.SessionStrings, s)
			existing[s] = true
		}
	}

	if len(dbSessions) > 0 {
		slog.Info("[DB] Loaded session strings from database", "count", len(dbSessions))
	}
}

func Init(client *tg.Client) error {
	if database.Instance == nil {
		return fmt.Errorf("database not initialized")
	}

	loadOwnerSettings()
	loadDBSessions()

	validSessions := make([]string, 0, len(config.Conf.SessionStrings))
	for _, session := range config.Conf.SessionStrings {
		_, err := vc.Calls.StartClient(config.Conf.ApiId, config.Conf.ApiHash, session)
		if err != nil {
			slog.Error("[Init] Failed to start client, removing session", "error", err)
			continue
		}
		validSessions = append(validSessions, session)
	}
	if len(validSessions) != len(config.Conf.SessionStrings) {
		config.Conf.SessionStrings = validSessions
		_ = database.DeleteAllSessionKeys()
		for _, s := range validSessions {
			_ = database.SetSetting(fmt.Sprintf("session_db_%d", time.Now().UnixNano()+int64(len(validSessions))), s)
		}
	}

	vc.Calls.RegisterHandlers(client)
	return nil
}
