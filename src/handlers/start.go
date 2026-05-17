package handlers

import (
	"fmt"
	"musicflarebot/config"
	"runtime"
	"time"

	"musicflarebot/src/core"
	"musicflarebot/src/core/db"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func pingHandler(m *tg.NewMessage) error {
	start := time.Now()

	msg, err := m.Reply("Pinging… please wait…")
	if err != nil {
		return err
	}

	latency := time.Since(start).Milliseconds()
	uptime := getFormattedDuration(time.Since(startTime))

	response := fmt.Sprintf(
		"<b>📊 System Performance Metrics</b>\n\n"+
			"<b>Bot Latency:</b> <code>%d ms</code>\n"+
			"<b>Uptime:</b> <code>%s</code>\n"+
			"<b>Go Routines:</b> <code>%d</code>\n",
		latency, uptime, runtime.NumGoroutine(),
	)

	_, err = msg.Edit(response, &tg.SendOptions{ParseMode: "HTML"})
	return err
}

func startHandler(m *tg.NewMessage) error {
	chatID := m.ChatID()
	if IsPrivate(m) {
		go func(chatID int64) {
			_ = db.Instance.AddUser(chatID)
		}(chatID)

		response := fmt.Sprintf(
			"Hey %s,\nThis is %s !\n\n<b>Supported Platforms:</b> YouTube, Spotify, Apple Music, SoundCloud, MXPlayer, Deezer, Twitch, Kick....\n\n<b><i>Click on the help button for more info.</i></b>",
			firstName(m),
			client.Me().FirstName,
		)

		_, err := m.ReplyMedia(config.Conf.StartImg, &tg.MediaOptions{
			ParseMode:   "HTML",
			Caption:     response,
			ReplyMarkup: core.AddMeMarkup(client.Me().Username),
		})

		return err
	}

	go func(chatID int64) {
		_ = db.Instance.AddChat(chatID)
	}(chatID)

	uptime := getFormattedDuration(time.Since(startTime))
	response := fmt.Sprintf(
		"<b>🎵 %s is ready</b>\n"+
			"<b>Uptime:</b> <code>%s</code>\n\n"+
			"<i>A music player bot with some awesome and useful features.</i>",
		client.Me().FirstName,
		uptime,
	)

	_, err := m.Reply(response, &tg.SendOptions{
		ParseMode:   "HTML",
		LinkPreview: false,
		ReplyMarkup: core.SupportBtn(),
	})

	return err
}
