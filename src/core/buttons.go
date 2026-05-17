package core

import (
	"musicflarebot/config"
	"musicflarebot/src/utils"
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"
)

var CloseBtn = tg.Button.Data("Close", "vcplay_close")
var HomeBtn = tg.Button.Data("Home", "help_back")
var HelpBtn = tg.Button.Data("Help", "help_all")
var UserBtn = tg.Button.Data("Users", "help_user")
var AdminBtn = tg.Button.Data("Admins", "help_admin")
var OwnerBtn = tg.Button.Data("Owner", "help_owner")
var DevsBtn = tg.Button.Data("Devs", "help_devs")
var PlaylistBtn = tg.Button.Data("Playlist", "help_playlist")

var SourceCodeBtn = tg.Button.URL("Source Code", "https://github.com/FlareBase/MusicFlareBot")

func supportRow() *tg.KeyboardBuilder {
	kb := tg.NewKeyboard()
	if config.Conf.SupportChannel != "" {
		kb.AddRow(tg.Button.URL("Updates", config.Conf.SupportChannel))
	}
	if config.Conf.SupportGroup != "" {
		kb.AddRow(tg.Button.URL("Group", config.Conf.SupportGroup))
	}
	return kb
}

func SupportKeyboard() *tg.ReplyInlineMarkup {
	return supportRow().AddRow(CloseBtn).Build()
}

func SupportBtn() *tg.ReplyInlineMarkup {
	return supportRow().Build()
}

func SettingsKeyboard(playMode, adminMode string, cmdDelete bool, language string) *tg.ReplyInlineMarkup {
	playText := "Everyone"
	if playMode == utils.Admins {
		playText = "Admins"
	}

	deleteText := "False"
	if cmdDelete {
		deleteText = "True"
	}

	adminText := "Everyone"
	if adminMode == utils.Admins {
		adminText = "Admins"
	}

	langText := "English"
	if language != "en" && language != "" {
		langText = language
	}

	return tg.NewKeyboard().
		AddRow(tg.Button.Data("Play Mode ➜", "settings_main"), tg.Button.Data(playText, "settings_play")).
		AddRow(tg.Button.Data("Command Delete ➜", "settings_main"), tg.Button.Data(deleteText, "settings_delete")).
		AddRow(tg.Button.Data("Admin Mode ➜", "settings_main"), tg.Button.Data(adminText, "settings_admin")).
		AddRow(tg.Button.Data("Language ➜", "settings_main"), tg.Button.Data(langText, "settings_lang")).
		AddRow(CloseBtn).
		Build()
}

func HelpMenuKeyboard() *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(UserBtn, AdminBtn, OwnerBtn).
		AddRow(PlaylistBtn, DevsBtn, CloseBtn).
		AddRow(HomeBtn).
		Build()
}

func BackHelpMenuKeyboard() *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(HelpBtn, HomeBtn).
		AddRow(CloseBtn, SourceCodeBtn).
		Build()
}

func ControlButtons(mode string) *tg.ReplyInlineMarkup {
	skipBtn := tg.Button.Data("‣‣I", "play_skip")
	stopBtn := tg.Button.Data("▢", "play_stop")
	pauseBtn := tg.Button.Data("II", "play_pause")
	resumeBtn := tg.Button.Data("▷", "play_resume")
	muteBtn := tg.Button.Data("🔇", "play_mute")
	unmuteBtn := tg.Button.Data("🔊", "play_unmute")
	addToPlaylistBtn := tg.Button.Data("➕", "play_add_to_list")

	switch mode {
	case "play":
		return tg.NewKeyboard().
			AddRow(skipBtn, stopBtn, pauseBtn).
			AddRow(addToPlaylistBtn, CloseBtn).
			Build()
	case "pause":
		return tg.NewKeyboard().
			AddRow(skipBtn, stopBtn, resumeBtn).
			AddRow(CloseBtn).
			Build()
	case "resume":
		return tg.NewKeyboard().
			AddRow(skipBtn, stopBtn, pauseBtn).
			AddRow(CloseBtn).
			Build()
	case "mute":
		return tg.NewKeyboard().
			AddRow(skipBtn, stopBtn, unmuteBtn).
			AddRow(CloseBtn).
			Build()
	case "unmute":
		return tg.NewKeyboard().
			AddRow(skipBtn, stopBtn, muteBtn).
			AddRow(CloseBtn).
			Build()
	default:
		return tg.NewKeyboard().
			AddRow(CloseBtn).
			Build()
	}
}

func AddMeMarkup(username string) *tg.ReplyInlineMarkup {
	addMeBtn := tg.Button.URL(
		"Aᴅᴅ ᴍᴇ ᴛᴏ ʏᴏᴜʀ ɢʀᴏᴜᴘ",
		fmt.Sprintf("https://t.me/%s?startgroup=true", username),
	)

	kb := tg.NewKeyboard().AddRow(addMeBtn).AddRow(HelpBtn)
	row := make([]tg.KeyboardButton, 0, 2)
	if config.Conf.SupportChannel != "" {
		row = append(row, tg.Button.URL("Updates", config.Conf.SupportChannel))
	}
	if config.Conf.SupportGroup != "" {
		row = append(row, tg.Button.URL("Group", config.Conf.SupportGroup))
	}
	if len(row) > 0 {
		kb.AddRow(row...)
	}
	return kb.AddRow(SourceCodeBtn).Build()
}
