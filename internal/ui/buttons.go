package ui

import (
	"fmt"
	"strconv"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func StartButtons(devs []int64) *tg.InlineKeyboardMarkup {
	buttons := [][]tg.InlineKeyboardButton{}
	row := []tg.InlineKeyboardButton{}
	for _, dev := range devs {
		row = append(row, tg.InlineKeyboardButton{
			Text: "👨‍💻 Developer",
			URL:  fmt.Sprintf("tg://user?id=%d", dev),
		})
	}
	if len(row) > 0 {
		buttons = append(buttons, row)
	}
	buttons = append(buttons, []tg.InlineKeyboardButton{
		{Text: "➕ Add to Group", URL: "https://t.me/TuneNetwork_bot?startgroup=true"},
	})
	buttons = append(buttons, []tg.InlineKeyboardButton{
		{Text: "💬 Support Group", CallbackData: "support_group"},
		{Text: "📢 Channel", CallbackData: "support_channel"},
	})
	buttons = append(buttons, []tg.InlineKeyboardButton{
		{Text: "❓ Help", CallbackData: "help"},
		{Text: "ℹ️ About", CallbackData: "about"},
	})
	if len(buttons) > 0 {
		return &tg.InlineKeyboardMarkup{InlineKeyboard: buttons}
	}
	return nil
}

func HelpButtons() *tg.InlineKeyboardMarkup {
	return &tg.InlineKeyboardMarkup{
		InlineKeyboard: [][]tg.InlineKeyboardButton{
			{
				{Text: "🎵 Music", CallbackData: "help_music"},
				{Text: "🎧 Playback", CallbackData: "help_playback"},
				{Text: "🛠 Admin", CallbackData: "help_admin"},
			},
			{
				{Text: "ℹ️ About", CallbackData: "help_about"},
				{Text: "🏠 Home", CallbackData: "back_to_start"},
			},
		},
	}
}

func PlayButtons(chatId int64) *tg.InlineKeyboardMarkup {
	return &tg.InlineKeyboardMarkup{
		InlineKeyboard: [][]tg.InlineKeyboardButton{
			{
				{Text: "⏸ Pause", CallbackData: fmt.Sprintf("pause_%d", chatId)},
				{Text: "⏭ Skip", CallbackData: fmt.Sprintf("skip_%d", chatId)},
				{Text: "⏹ Stop", CallbackData: fmt.Sprintf("stop_%d", chatId)},
			},
			{
				{Text: "📃 Queue", CallbackData: fmt.Sprintf("queue_%d", chatId)},
				{Text: "🔁 Loop", CallbackData: fmt.Sprintf("loop_%d", chatId)},
			},
			{
				{Text: "♻️ Skip to DJ", CallbackData: fmt.Sprintf("dj_%d", chatId)},
				{Text: "↔️ Move to VC", CallbackData: fmt.Sprintf("move_vc_%d", chatId)},
			},
		},
	}
}

func QueueButtons(page, totalPages int, chatId int64) *tg.InlineKeyboardMarkup {
	buttons := [][]tg.InlineKeyboardButton{}
	navRow := []tg.InlineKeyboardButton{}
	if page > 0 {
		navRow = append(navRow, tg.InlineKeyboardButton{
			Text:         "◀️",
			CallbackData: fmt.Sprintf("queue_page_%d_%d", page-1, chatId),
		})
	}
	navRow = append(navRow, tg.InlineKeyboardButton{
		Text:         fmt.Sprintf("%d/%d", page+1, totalPages),
		CallbackData: "noop",
	})
	if page < totalPages-1 {
		navRow = append(navRow, tg.InlineKeyboardButton{
			Text:         "▶️",
			CallbackData: fmt.Sprintf("queue_page_%d_%d", page+1, chatId),
		})
	}
	buttons = append(buttons, navRow)
	buttons = append(buttons, []tg.InlineKeyboardButton{
		{Text: "🔄 Refresh", CallbackData: fmt.Sprintf("queue_%d", chatId)},
		{Text: "🗑 Close", CallbackData: "close"},
	})
	return &tg.InlineKeyboardMarkup{InlineKeyboard: buttons}
}

func LoopButton(chatId int64, loop int) *tg.InlineKeyboardMarkup {
	modes := []struct {
		label string
		data string
		active bool
	}{
		{"🔁 Off", fmt.Sprintf("loop_set_0_%d", chatId), loop == 0},
		{"🔂 One", fmt.Sprintf("loop_set_1_%d", chatId), loop == 1},
		{"🔁 All", fmt.Sprintf("loop_set_2_%d", chatId), loop == 2},
	}
	row := []tg.InlineKeyboardButton{}
	for _, m := range modes {
		text := m.label
		if m.active {
			text = "✅ " + text
		}
		row = append(row, tg.InlineKeyboardButton{Text: text, CallbackData: m.data})
	}
	return &tg.InlineKeyboardMarkup{
		InlineKeyboard: [][]tg.InlineKeyboardButton{
			row,
			{{Text: "🗑 Close", CallbackData: "close"}},
		},
	}
}

func ConfirmButton(action string, id interface{}) *tg.InlineKeyboardMarkup {
	var idStr string
	switch v := id.(type) {
	case int64:
		idStr = strconv.FormatInt(v, 10)
	case string:
		idStr = v
	case int:
		idStr = strconv.Itoa(v)
	}
	return &tg.InlineKeyboardMarkup{
		InlineKeyboard: [][]tg.InlineKeyboardButton{
			{
				{Text: "✅ Yes", CallbackData: fmt.Sprintf("confirm_%s_%s", action, idStr)},
				{Text: "❌ No", CallbackData: "close"},
			},
		},
	}
}

func PlaylistButtons(playlists []string, userId int64) *tg.InlineKeyboardMarkup {
	buttons := [][]tg.InlineKeyboardButton{}
	for _, name := range playlists {
		buttons = append(buttons, []tg.InlineKeyboardButton{
			{Text: name, CallbackData: fmt.Sprintf("playlist_%s_%d", name, userId)},
		})
	}
	buttons = append(buttons, []tg.InlineKeyboardButton{
		{Text: "🗑 Close", CallbackData: "close"},
	})
	return &tg.InlineKeyboardMarkup{InlineKeyboard: buttons}
}

func ServicesKeyboard() *tg.InlineKeyboardMarkup {
	return &tg.InlineKeyboardMarkup{
		InlineKeyboard: [][]tg.InlineKeyboardButton{
			{
				{Text: "🎵 YouTube", CallbackData: "service_youtube"},
				{Text: "🎧 Spotify", CallbackData: "service_spotify"},
			},
			{
				{Text: "🎶 JioSaavn", CallbackData: "service_jiosaavn"},
				{Text: "🍎 Apple Music", CallbackData: "service_applemusic"},
			},
			{
				{Text: "☁️ SoundCloud", CallbackData: "service_soundcloud"},
				{Text: "🗑 Close", CallbackData: "close"},
			},
		},
	}
}

func SettingsKeyboard() *tg.InlineKeyboardMarkup {
	return &tg.InlineKeyboardMarkup{
		InlineKeyboard: [][]tg.InlineKeyboardButton{
			{
				{Text: "🎵 Default Service", CallbackData: "set_service"},
				{Text: "⏱ Max Duration", CallbackData: "set_duration"},
			},
			{
				{Text: "📢 Channel", CallbackData: "set_channel"},
				{Text: "💬 Support Group", CallbackData: "set_support"},
			},
			{
				{Text: "🆘 Start Message", CallbackData: "set_start_msg"},
				{Text: "🎬 Start Media", CallbackData: "set_start_media"},
			},
			{
				{Text: "🗑 Close", CallbackData: "close"},
			},
		},
	}
}

func CloseButton() *tg.InlineKeyboardMarkup {
	return &tg.InlineKeyboardMarkup{
		InlineKeyboard: [][]tg.InlineKeyboardButton{
			{{Text: "🗑 Close", CallbackData: "close"}},
		},
	}
}

func LinkButton(text, url string) *tg.InlineKeyboardMarkup {
	return &tg.InlineKeyboardMarkup{
		InlineKeyboard: [][]tg.InlineKeyboardButton{
			{{Text: text, URL: url}},
		},
	}
}
