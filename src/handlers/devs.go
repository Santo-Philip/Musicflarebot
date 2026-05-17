package handlers

import (
	"fmt"
	"musicflarebot/config"
	"strings"

	"musicflarebot/src/core/cache"
	"musicflarebot/src/core/db"
	"musicflarebot/src/vc"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func activeVcHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	activeChats := cache.ChatCache.GetActiveChats()
	if len(activeChats) == 0 {
		_, err := m.Reply("No active chats found.")
		return err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🎵 <b>Active Voice Chats</b> (%d):\n\n", len(activeChats)))

	for _, chatID := range activeChats {
		queueLength := cache.ChatCache.GetQueueLength(chatID)
		currentSong := cache.ChatCache.GetPlayingTrack(chatID)

		var songInfo string
		if currentSong != nil {
			songInfo = fmt.Sprintf(
				"🎶 <b>Now Playing:</b> <a href='%s'>%s</a> (%ds)",
				currentSong.URL,
				currentSong.Name,
				currentSong.Duration,
			)
		} else {
			songInfo = "🔇 No song playing."
		}

		sb.WriteString(fmt.Sprintf(
			"➤ <b>Chat ID:</b> <code>%d</code>\n📌 <b>Queue Size:</b> %d\n%s\n\n",
			chatID,
			queueLength,
			songInfo,
		))
	}

	text := sb.String()
	if len(text) > 4096 {
		text = fmt.Sprintf("🎵 <b>Active Voice Chats</b> (%d)", len(activeChats))
	}

	_, err := m.Reply(text, &tg.SendOptions{ParseMode: "HTML", LinkPreview: false})
	if err != nil {
		return err
	}

	return nil
}

func clearAssistantsHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	done, err := db.Instance.ClearAllAssistants()
	if err != nil {
		_, _ = m.Reply(fmt.Sprintf("failed to clear assistants: %s", err.Error()))
		return nil
	}

	_, err = m.Reply(fmt.Sprintf("Removed assistant from %d chats", done))
	return err
}

func leaveAllHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	reply, err := m.Reply("Assistant is leaving all chats...")
	if err != nil {
		return err
	}

	leftCount, err := vc.Calls.LeaveAll()
	if err != nil {
		_, _ = reply.Edit(fmt.Sprintf("Failed to leave all chats: %s", err.Error()))
		return err
	}

	_, err = reply.Edit(fmt.Sprintf("Assistant's Left %d chats", leftCount))
	return err
}

func loggerHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	if config.Conf.LoggerId == 0 {
		_, _ = m.Reply("Please set LOGGER_ID in .env first.")
		return nil
	}

	loggerStatus := db.Instance.GetLoggerStatus()
	args := strings.ToLower(Args(m))
	if len(args) == 0 {
		_, _ = m.Reply(fmt.Sprintf("Usage: /logger [enable|disable|on|off]\nCurrent status: %t", loggerStatus))
		return nil
	}

	switch args {
	case "enable", "on":
		_ = db.Instance.SetLoggerStatus(true)
		_, _ = m.Reply("Logger Enabled")
	case "disable", "off":
		_ = db.Instance.SetLoggerStatus(false)
		_, _ = m.Reply("Logger disabled")
	default:
		_, _ = m.Reply("Invalid argument. Use 'enable', 'disable', 'on', or 'off'.")
	}

	return nil
}
