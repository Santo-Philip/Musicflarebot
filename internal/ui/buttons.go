package ui

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func row(btns ...tg.KeyboardButton) *tg.KeyboardButtonRow {
	return &tg.KeyboardButtonRow{Buttons: btns}
}

func markup(rows ...*tg.KeyboardButtonRow) *tg.ReplyInlineMarkup {
	return &tg.ReplyInlineMarkup{Rows: rows}
}

func dataBtn(text, data string) tg.KeyboardButton {
	return tg.Button.Data(text, data)
}

func urlBtn(text, url string) tg.KeyboardButton {
	return tg.Button.URL(text, url)
}

func StartButtons(devs []int64) *tg.ReplyInlineMarkup {
	var rows []*tg.KeyboardButtonRow
	for _, dev := range devs {
		rows = append(rows, row(urlBtn("👨‍💻 Developer", fmt.Sprintf("tg://user?id=%d", dev))))
	}
	rows = append(rows,
		row(urlBtn("➕ Add to Group", "https://t.me/TuneNetwork_bot?startgroup=true")),
		row(dataBtn("💬 Support Group", "support_group"), dataBtn("📢 Channel", "support_channel")),
		row(dataBtn("❓ Help", "help"), dataBtn("ℹ️ About", "about")),
	)
	return markup(rows...)
}

func HelpButtons() *tg.ReplyInlineMarkup {
	return markup(
		row(
			dataBtn("🎵 Music", "help_music"),
			dataBtn("🎧 Playback", "help_playback"),
			dataBtn("🛠 Admin", "help_admin"),
		),
		row(dataBtn("ℹ️ About", "help_about"), dataBtn("🏠 Home", "back_to_start")),
	)
}

func PlayButtons(chatId int64) *tg.ReplyInlineMarkup {
	return markup(
		row(
			dataBtn("⏸ Pause", fmt.Sprintf("pause_%d", chatId)),
			dataBtn("⏭ Skip", fmt.Sprintf("skip_%d", chatId)),
			dataBtn("⏹ Stop", fmt.Sprintf("stop_%d", chatId)),
		),
		row(
			dataBtn("📃 Queue", fmt.Sprintf("queue_%d", chatId)),
			dataBtn("🔁 Loop", fmt.Sprintf("loop_%d", chatId)),
		),
		row(
			dataBtn("♻️ Skip to DJ", fmt.Sprintf("dj_%d", chatId)),
			dataBtn("↔️ Move to VC", fmt.Sprintf("move_vc_%d", chatId)),
		),
	)
}

func QueueButtons(page, totalPages int, chatId int64) *tg.ReplyInlineMarkup {
	navRow := []tg.KeyboardButton{}
	if page > 0 {
		navRow = append(navRow, dataBtn("◀️", fmt.Sprintf("queue_page_%d_%d", page-1, chatId)))
	}
	navRow = append(navRow, dataBtn(fmt.Sprintf("%d/%d", page+1, totalPages), "noop"))
	if page < totalPages-1 {
		navRow = append(navRow, dataBtn("▶️", fmt.Sprintf("queue_page_%d_%d", page+1, chatId)))
	}
	return markup(
		row(navRow...),
		row(dataBtn("🔄 Refresh", fmt.Sprintf("queue_%d", chatId)), dataBtn("🗑 Close", "close")),
	)
}

func LoopButton(chatId int64, loop int) *tg.ReplyInlineMarkup {
	modes := []struct {
		label string
		data  string
		active bool
	}{
		{"🔁 Off", fmt.Sprintf("loop_set_0_%d", chatId), loop == 0},
		{"🔂 One", fmt.Sprintf("loop_set_1_%d", chatId), loop == 1},
		{"🔁 All", fmt.Sprintf("loop_set_2_%d", chatId), loop == 2},
	}
	r := []tg.KeyboardButton{}
	for _, m := range modes {
		text := m.label
		if m.active {
			text = "✅ " + text
		}
		r = append(r, dataBtn(text, m.data))
	}
	return markup(row(r...), row(dataBtn("🗑 Close", "close")))
}

func ConfirmButton(action string, id interface{}) *tg.ReplyInlineMarkup {
	return markup(
		row(
			dataBtn("✅ Yes", fmt.Sprintf("confirm_%s_%v", action, id)),
			dataBtn("❌ No", "close"),
		),
	)
}

func PlaylistButtons(playlists []string, userId int64) *tg.ReplyInlineMarkup {
	rows := [][]tg.KeyboardButton{}
	for _, name := range playlists {
		rows = append(rows, []tg.KeyboardButton{dataBtn(name, fmt.Sprintf("playlist_%s_%d", name, userId))})
	}
	rows = append(rows, []tg.KeyboardButton{dataBtn("🗑 Close", "close")})
	var r []*tg.KeyboardButtonRow
	for _, btns := range rows {
		r = append(r, row(btns...))
	}
	return markup(r...)
}

func ServicesKeyboard() *tg.ReplyInlineMarkup {
	return markup(
		row(dataBtn("🎵 YouTube", "service_youtube"), dataBtn("🎧 Spotify", "service_spotify")),
		row(dataBtn("🎶 JioSaavn", "service_jiosaavn"), dataBtn("🍎 Apple Music", "service_applemusic")),
		row(dataBtn("☁️ SoundCloud", "service_soundcloud"), dataBtn("🗑 Close", "close")),
	)
}

func SettingsKeyboard() *tg.ReplyInlineMarkup {
	return markup(
		row(dataBtn("🎵 Default Service", "set_service"), dataBtn("⏱ Max Duration", "set_duration")),
		row(dataBtn("📢 Channel", "set_channel"), dataBtn("💬 Support Group", "set_support")),
		row(dataBtn("🆘 Start Message", "set_start_msg"), dataBtn("🎬 Start Media", "set_start_media")),
		row(dataBtn("🗑 Close", "close")),
	)
}

func CloseButton() *tg.ReplyInlineMarkup {
	return markup(row(dataBtn("🗑 Close", "close")))
}

func LinkButton(text, url string) *tg.ReplyInlineMarkup {
	return markup(row(urlBtn(text, url)))
}

var CloseBtn = dataBtn("Close", "vcplay_close")
var HomeBtn = dataBtn("Home", "help_back")

func ControlButtons(mode string) *tg.ReplyInlineMarkup {
	skipBtn := dataBtn("‣‣I", "play_skip")
	stopBtn := dataBtn("▢", "play_stop")
	pauseBtn := dataBtn("II", "play_pause")
	resumeBtn := dataBtn("▷", "play_resume")
	muteBtn := dataBtn("🔇", "play_mute")
	unmuteBtn := dataBtn("🔊", "play_unmute")

	switch mode {
	case "play":
		return markup(row(skipBtn, stopBtn, pauseBtn), row(CloseBtn))
	case "pause":
		return markup(row(skipBtn, stopBtn, resumeBtn), row(CloseBtn))
	case "resume":
		return markup(row(skipBtn, stopBtn, pauseBtn), row(CloseBtn))
	case "mute":
		return markup(row(skipBtn, stopBtn, unmuteBtn), row(CloseBtn))
	case "unmute":
		return markup(row(skipBtn, stopBtn, muteBtn), row(CloseBtn))
	default:
		return markup(row(CloseBtn))
	}
}

func AddMeMarkup(username string) *tg.ReplyInlineMarkup {
	return StartButtons(nil)
}

func SupportBtn() *tg.ReplyInlineMarkup {
	return markup()
}

func SupportKeyboard() *tg.ReplyInlineMarkup {
	return markup(row(CloseBtn))
}

func HelpMenuKeyboard() *tg.ReplyInlineMarkup {
	return markup(row(HomeBtn), row(CloseBtn))
}

func BackHelpMenuKeyboard() *tg.ReplyInlineMarkup {
	return markup(row(HomeBtn), row(CloseBtn))
}
