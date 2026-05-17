package src

import (
	"fmt"
	"log/slog"
	"time"

	"musicflarebot/config"
	"musicflarebot/src/core/db"
	"musicflarebot/src/vc"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func loadOwnerSettings() {
	if loggerId := db.Instance.GetOwnerLoggerId(); loggerId > 0 {
		config.Conf.LoggerId = loggerId
		slog.Info("[DB] Loaded LoggerId from database", "id", loggerId)
	}
	if group := db.Instance.GetOwnerSupportGroup(); group != "" {
		config.Conf.SupportGroup = group
		slog.Info("[DB] Loaded SupportGroup from database")
	}
	if channel := db.Instance.GetOwnerSupportChannel(); channel != "" {
		config.Conf.SupportChannel = channel
		slog.Info("[DB] Loaded SupportChannel from database")
	}
	if dur := db.Instance.GetOwnerSongDuration(); dur > 0 {
		config.Conf.SongDurationLimit = dur
		slog.Info("[DB] Loaded SongDurationLimit from database", "seconds", dur)
	}
	if svc := db.Instance.GetOwnerDefaultService(); svc != "" {
		config.Conf.DefaultService = svc
		slog.Info("[DB] Loaded DefaultService from database", "service", svc)
	}
	if devs := db.Instance.GetOwnerDevs(); len(devs) > 0 {
		config.Conf.DEVS = devs
		slog.Info("[DB] Loaded DEVS from database", "count", len(devs))
	}
}

func loadDBSessions() {
	dbSessions, err := db.Instance.GetSessionStrings()
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
	if err := db.InitDatabase(); err != nil {
		return err
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
		_ = db.Instance.DeleteAllSessionKeys()
		for _, s := range validSessions {
			_ = db.Instance.SetSetting(fmt.Sprintf("session_db_%d", time.Now().UnixNano()+int64(len(validSessions))), s)
		}
	}

	vc.Calls.RegisterHandlers(client)
	return nil
}
