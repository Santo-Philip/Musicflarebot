package handlers

import (
	"fmt"
	"log/slog"
	"musicflarebot/internal/config"
	"musicflarebot/internal/cache"
	"musicflarebot/internal/database"
	"musicflarebot/internal/types"
	"strconv"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func setLoggerIdHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	args := Args(m)
	if args == "" {
		_, _ = m.Reply("Usage: /set_logger_id [chat_id]")
		return nil
	}

	id, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		_, _ = m.Reply("Invalid chat ID. Please provide a numeric ID.")
		return nil
	}

	if err = database.SetSetting("logger_id", args); err != nil {
		slog.Error("Failed to set logger_id", "error", err)
		_, _ = m.Reply("Failed to save logger ID to database.")
		return nil
	}

	config.Conf.LoggerId = id
	_ = database.SetSetting("logger_id", fmt.Sprintf("%d", id))

	_, _ = m.Reply(fmt.Sprintf("Logger ID has been set to <code>%d</code>.", id))
	return nil
}

func setSupportGroupHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	args := Args(m)
	if args == "" {
		_, _ = m.Reply("Usage: /set_support_group [username]")
		return nil
	}

	username := strings.TrimPrefix(args, "@")
	if err := database.SetSetting("support_group", username); err != nil {
		slog.Error("Failed to set support_group", "error", err)
		_, _ = m.Reply("Failed to save support group to database.")
		return nil
	}

	config.Conf.SupportGroup = username
	_, _ = m.Reply(fmt.Sprintf("Support group has been set to <b>@%s</b>.", username))
	return nil
}

func setSupportChannelHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	args := Args(m)
	if args == "" {
		_, _ = m.Reply("Usage: /set_support_channel [username]")
		return nil
	}

	username := strings.TrimPrefix(args, "@")
	if err := database.SetSetting("support_channel", username); err != nil {
		slog.Error("Failed to set support_channel", "error", err)
		_, _ = m.Reply("Failed to save support channel to database.")
		return nil
	}

	config.Conf.SupportChannel = username
	_, _ = m.Reply(fmt.Sprintf("Support channel has been set to <b>@%s</b>.", username))
	return nil
}

func setSongDurationHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	args := Args(m)
	if args == "" {
		_, _ = m.Reply("Usage: /set_song_duration [seconds]")
		return nil
	}

	duration, err := strconv.Atoi(args)
	if err != nil || duration <= 0 {
		_, _ = m.Reply("Invalid duration. Please provide a positive number in seconds.")
		return nil
	}

	if err = database.SetSetting("song_duration_limit", args); err != nil {
		slog.Error("Failed to set song_duration_limit", "error", err)
		_, _ = m.Reply("Failed to save song duration limit to database.")
		return nil
	}

	config.Conf.SongDurationLimit = int64(duration)
	_, _ = m.Reply(fmt.Sprintf("Song duration limit has been set to <b>%d seconds</b> (%d minutes).", duration, duration/60))
	return nil
}

func setDefaultServiceHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	args := Args(m)
	if args == "" {
		_, _ = m.Reply("Usage: /set_default_service [service_name]")
		return nil
	}

	service := strings.ToLower(strings.TrimSpace(args))
	validServices := []string{"youtube", "spotify", "jiosaavn", "applemusic", "soundcloud"}
	valid := false
	for _, s := range validServices {
		if s == service {
			valid = true
			break
		}
	}

	if !valid {
		_, _ = m.Reply(fmt.Sprintf("Invalid service. Valid options: %s", strings.Join(validServices, ", ")))
		return nil
	}

	if err := database.SetSetting("default_service", service); err != nil {
		slog.Error("Failed to set default_service", "error", err)
		_, _ = m.Reply("Failed to save default service to database.")
		return nil
	}

	config.Conf.DefaultService = service
	_, _ = m.Reply(fmt.Sprintf("Default service has been set to <b>%s</b>.", service))
	return nil
}

func addDevHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	args := Args(m)
	if args == "" {
		_, _ = m.Reply("Usage: /add_dev [user_id]")
		return nil
	}

	userID, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		_, _ = m.Reply("Invalid user ID. Please provide a numeric ID.")
		return nil
	}

	for _, dev := range config.Conf.DEVS {
		if dev == userID {
			_, _ = m.Reply("This user is already a developer.")
			return nil
		}
	}

	config.Conf.DEVS = append(config.Conf.DEVS, userID)
	_ = database.SetSetting("devs", joinInt64s(config.Conf.DEVS, ","))

	_, _ = m.Reply(fmt.Sprintf("User <code>%d</code> has been added as a developer.", userID))
	return nil
}

func removeDevHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	args := Args(m)
	if args == "" {
		_, _ = m.Reply("Usage: /remove_dev [user_id]")
		return nil
	}

	userID, err := strconv.ParseInt(args, 10, 64)
	if err != nil {
		_, _ = m.Reply("Invalid user ID. Please provide a numeric ID.")
		return nil
	}

	newDevs := make([]int64, 0, len(config.Conf.DEVS))
	removed := false
	for _, dev := range config.Conf.DEVS {
		if dev == userID {
			removed = true
			continue
		}
		newDevs = append(newDevs, dev)
	}

	if !removed {
		_, _ = m.Reply("This user is not a developer.")
		return nil
	}

	config.Conf.DEVS = newDevs
	_ = database.SetSetting("devs", joinInt64s(config.Conf.DEVS, ","))

	_, _ = m.Reply(fmt.Sprintf("User <code>%d</code> has been removed from developers.", userID))
	return nil
}

func listSettingsHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("<b>Bot Settings</b>\n\n")
	sb.WriteString(fmt.Sprintf("<b>Logger ID:</b> <code>%d</code>\n", config.Conf.LoggerId))
	sb.WriteString(fmt.Sprintf("<b>Support Group:</b> <b>@%s</b>\n", config.Conf.SupportGroup))
	sb.WriteString(fmt.Sprintf("<b>Support Channel:</b> <b>@%s</b>\n", config.Conf.SupportChannel))
	sb.WriteString(fmt.Sprintf("<b>Song Duration Limit:</b> <code>%d</code> seconds (%d min)\n", config.Conf.SongDurationLimit, config.Conf.SongDurationLimit/60))
	sb.WriteString(fmt.Sprintf("<b>Default Service:</b> <code>%s</code>\n", config.Conf.DefaultService))
	sb.WriteString(fmt.Sprintf("<b>DEVS:</b> <code>%v</code>\n", config.Conf.DEVS))

	startMsg, _ := database.GetSetting(types.SettingStartMessage)
	startMedia, _ := database.GetSetting(types.SettingStartMedia)
	if startMsg != "" {
		sb.WriteString(fmt.Sprintf("<b>Start Msg:</b> set (%d chars)\n", len(startMsg)))
	} else {
		sb.WriteString("<b>Start Msg:</b> default\n")
	}
	if startMedia != "" {
		sb.WriteString(fmt.Sprintf("<b>Start Media:</b> <code>%s</code>\n", startMedia))
	} else {
		sb.WriteString("<b>Start Media:</b> default\n")
	}

	_, err := m.Reply(sb.String(), &tg.SendOptions{ParseMode: "HTML"})
	return err
}

func joinInt64s(items []int64, sep string) string {
	parts := make([]string, len(items))
	for i, v := range items {
		parts[i] = strconv.FormatInt(v, 10)
	}
	return strings.Join(parts, sep)
}

var _ = cache.ChatCache // silence unused import
