package handlers

import (
	"fmt"
	"io"
	"musicflarebot/internal/config"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"musicflarebot/internal/types"
	"musicflarebot/internal/ui"
	"musicflarebot/internal/database"

	tg "github.com/amarnathcjd/gogram/telegram"
)

var (
	startMediaOnce sync.Once
	startMediaPath string
)

func getCustomStartMedia() string {
	customPath, err := database.GetSetting(types.SettingStartMedia)
	if err != nil || customPath == "" {
		return ""
	}
	if _, err := os.Stat(customPath); err == nil {
		return customPath
	}
	return ""
}

func getCustomStartMessage() string {
	msg, err := database.GetSetting(types.SettingStartMessage)
	if err != nil {
		return ""
	}
	return msg
}

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

func getCachedStartMedia() string {
	if startMediaPath != "" {
		return startMediaPath
	}

	startMediaOnce.Do(func() {
		url := config.Conf.StartImg
		if url == "" {
			return
		}

		ext := ".mp4"
		lower := strings.ToLower(url)
		switch {
		case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
			ext = ".jpg"
		case strings.HasSuffix(lower, ".png"):
			ext = ".png"
		case strings.HasSuffix(lower, ".gif"):
			ext = ".gif"
		case strings.HasSuffix(lower, ".webm"):
			ext = ".webm"
		case strings.HasSuffix(lower, ".mov"):
			ext = ".mov"
		}

		dest := filepath.Join(config.Conf.DownloadsDir, "start_media"+ext)

		if _, err := os.Stat(dest); err == nil {
			startMediaPath = dest
			return
		}

		resp, err := http.Get(url)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		f, err := os.Create(dest)
		if err != nil {
			return
		}
		defer f.Close()

		if _, err := io.Copy(f, resp.Body); err != nil {
			os.Remove(dest)
			return
		}

		startMediaPath = dest
	})

	return startMediaPath
}

func startHandler(m *tg.NewMessage) error {
	chatID := m.ChatID()
	if IsPrivate(m) {
		go func(chatID int64) {
			_ = database.AddUser(chatID)
		}(chatID)

		customMsg := getCustomStartMessage()
		response := customMsg
		if response == "" {
			response = fmt.Sprintf(
				"Hey %s,\nThis is %s !\n\n<b>Supported Platforms:</b> YouTube, Spotify, Apple Music, SoundCloud, MXPlayer, Deezer, Twitch, Kick....\n\n<b><i>Click on the help button for more info.</i></b>",
				firstName(m),
				client.Me().FirstName,
			)
		}

		media := getCustomStartMedia()
		if media == "" {
			media = getCachedStartMedia()
		}
		if media == "" {
			media = config.Conf.StartImg
		}

		var err error
		if media == "" {
			_, err = m.Reply(response, &tg.SendOptions{
				ParseMode:   "HTML",
				LinkPreview: false,
				ReplyMarkup: ui.AddMeMarkup(client.Me().Username),
			})
		} else {
			_, err = m.ReplyMedia(media, &tg.MediaOptions{
				ParseMode:   "HTML",
				Caption:     response,
				ReplyMarkup: ui.AddMeMarkup(client.Me().Username),
			})
		}

		return err
	} else if IsSuperGroup(m) {
		// Handle supergroup
		go func(chatID int64) {
			_ = database.AddChat(chatID)
		}(chatID)

		uptime := getFormattedDuration(time.Since(startTime))
		response := fmt.Sprintf(
			"<b>🎵 %s is ready</b> (Supergroup Mode)\n"+
				"<b>Uptime:</b> <code>%s</code>\n\n"+
				"<i>A music player bot with some awesome and useful features.</i>",
			client.Me().FirstName,
			uptime,
		)

		_, err := m.Reply(response, &tg.SendOptions{
			ParseMode:   "HTML",
			LinkPreview: false,
			ReplyMarkup: ui.SupportBtn(),
		})

		return err
	} else if IsRegularGroup(m) {
		// Handle regular group
		go func(chatID int64) {
			_ = database.AddChat(chatID)
		}(chatID)

		uptime := getFormattedDuration(time.Since(startTime))
		response := fmt.Sprintf(
			"<b>🎵 %s is ready</b> (Group Mode)\n"+
				"<b>Uptime:</b> <code>%s</code>\n\n"+
				"<i>A music player bot with some awesome and useful features.</i>",
			client.Me().FirstName,
			uptime,
		)

		_, err := m.Reply(response, &tg.SendOptions{
			ParseMode:   "HTML",
			LinkPreview: false,
			ReplyMarkup: ui.SupportBtn(),
		})

		return err
	}

	// Fallback for any other chat type
	go func(chatID int64) {
		_ = database.AddChat(chatID)
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
		ReplyMarkup: ui.SupportBtn(),
	})

	return err
}
