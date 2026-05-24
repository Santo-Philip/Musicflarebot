package vc

import (
	"musicflarebot/internal/config"
	"musicflarebot/internal/types"
	"musicflarebot/internal/utils"
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"
)

// sendLogger sends a formatted log message to the designated logger chat.
func sendLogger(client *tg.Client, chatID int64, song *types.CachedTrack) {
	if chatID == 0 || song == nil || chatID == config.Conf.LoggerId {
		return
	}

	text := fmt.Sprintf(
		"<b>A song is playing</b> in <code>%d</code>\n\n‣ <b>Title:</b> <a href='%s'>%s</a>\n‣ <b>Duration:</b> %s\n‣ <b>Requested by:</b> %s\n‣ <b>Platform:</b> %s\n‣ <b>Is Video:</b> %t",
		chatID,
		song.URL,
		song.Name,
		utils.SecToMin(song.Duration),
		song.User,
		song.Platform,
		song.IsVideo,
	)

	_, err := client.SendMessage(config.Conf.LoggerId, text, &tg.SendOptions{ParseMode: "HTML", LinkPreview: false})
	if err != nil {
		logger.Warn("Failed to send the message", "error", err)
	}
}
