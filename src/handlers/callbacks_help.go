package handlers

import (
	"fmt"
	"musicflarebot/internal/config"
	"musicflarebot/internal/database"
	"musicflarebot/internal/types"
	"musicflarebot/internal/ui"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func supportGroupHandler(q *tg.CallbackQuery) error {
	username := config.Conf.SupportGroup
	if username == "" {
		_, _ = q.Answer("No support group configured.", &tg.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = q.Answer("Opening support group...")
	_, err := q.Edit("", &tg.SendOptions{
		ReplyMarkup: ui.LinkButton("💬 Support Group", fmt.Sprintf("https://t.me/%s", strings.TrimPrefix(username, "@"))),
	})
	return err
}

func supportChannelHandler(q *tg.CallbackQuery) error {
	username := config.Conf.SupportChannel
	if username == "" {
		_, _ = q.Answer("No support channel configured.", &tg.CallbackOptions{Alert: true})
		return nil
	}
	_, _ = q.Answer("Opening support channel...")
	_, err := q.Edit("", &tg.SendOptions{
		ReplyMarkup: ui.LinkButton("📢 Channel", fmt.Sprintf("https://t.me/%s", strings.TrimPrefix(username, "@"))),
	})
	return err
}

func aboutHandler(q *tg.CallbackQuery) error {
	text := "<b>ℹ️ About</b>\n\n" +
		"This is a Telegram music player bot built with Go.\n\n" +
		"<b>Supported Platforms:</b> YouTube, Spotify, Apple Music, SoundCloud, JioSaavn, Deezer, Twitch, Kick\n\n" +
		"<b>Developer:</b> <a href='https://t.me/nexfang'>@NexFang</a>"

	_, _ = q.Answer("")
	_, err := q.Edit(text, &tg.SendOptions{
		ParseMode:   "HTML",
		LinkPreview: false,
		ReplyMarkup: ui.BackHelpMenuKeyboard(),
	})
	return err
}

func helpMenuHandler(q *tg.CallbackQuery) error {
	response := getHelpText("all", "")
	_, _ = q.Answer("")
	_, err := q.Edit(response, &tg.SendOptions{
		ParseMode:   "HTML",
		LinkPreview: false,
		ReplyMarkup: ui.HelpMenuKeyboard(),
	})
	return err
}

func backToStartHandler(q *tg.CallbackQuery) error {
	customMsg, _ := database.GetSetting(types.SettingStartMessage)
	response := customMsg
	if response == "" {
		response = fmt.Sprintf(
			"Hey there,\nThis is %s !\n\n<b>Supported Platforms:</b> YouTube, Spotify, Apple Music, SoundCloud, MXPlayer, Deezer, Twitch, Kick....\n\n<b><i>Click on the help button for more info.</i></b>",
			client.Me().FirstName,
		)
	}
	_, _ = q.Answer("")
	_, err := q.Edit(response, &tg.SendOptions{
		ParseMode:   "HTML",
		LinkPreview: false,
		ReplyMarkup: ui.AddMeMarkup(client.Me().Username),
	})
	return err
}
