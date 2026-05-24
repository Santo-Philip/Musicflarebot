package handlers

import (
	"fmt"
	"musicflarebot/internal/cache"
	"musicflarebot/src/vc"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func eqHandler(m *tg.NewMessage) error {
	if m.ChatID() > 0 {
		_, err := m.Reply("Equalizer works only in group voice chats.")
		return err
	}

	chatID := m.ChatID()
	current := cache.ChatCache.GetEQPreset(chatID)
	if current == "" {
		current = "normal"
	}

	kb := buildEQKeyboard(chatID)
	text := fmt.Sprintf("<b>Equalizer</b>\n\nCurrent: <b>%s</b>\n\nSelect a preset to change the audio effect.", formatEQName(current))

	_, err := m.Reply(text, &tg.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: kb,
	})
	return err
}

func formatEQName(name string) string {
	eq := vc.EQByName(name)
	if eq != nil {
		return eq.Label
	}
	return name
}

func eqCallbackHandler(q *tg.CallbackQuery) error {
	data := q.DataString()
	if !strings.HasPrefix(data, "eq_") {
		return nil
	}

	preset := strings.TrimPrefix(data, "eq_")
	chatID := q.ChatID

	if preset == "close" {
		q.Delete()
		return nil
	}

	cache.ChatCache.SetEQPreset(chatID, preset)

	playing := cache.ChatCache.GetPlayingTrack(chatID)
	if playing != nil {
		eqFilter := vc.BuildAudioFilterFlag(preset)
		if err := vc.Calls.PlayMedia(chatID, playing.FilePath, playing.IsVideo, eqFilter); err != nil {
			_, _ = q.Answer(fmt.Sprintf("Failed to apply preset: %s", err.Error()), &tg.CallbackOptions{Alert: true})
			return nil
		}
	}

	_, _ = q.Edit(fmt.Sprintf("<b>Equalizer</b>\n\nCurrent: <b>%s</b>\n\nSelect a preset to change the audio effect.", formatEQName(preset)), &tg.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: buildEQKeyboard(chatID),
	})

	_, _ = q.Answer(fmt.Sprintf("EQ set to %s", formatEQName(preset)))
	return nil
}

func playEqCallbackHandler(q *tg.CallbackQuery) error {
	chatID := q.ChatID
	current := cache.ChatCache.GetEQPreset(chatID)
	if current == "" {
		current = "normal"
	}

	_, err := q.Edit(fmt.Sprintf("<b>Equalizer</b>\n\nCurrent: <b>%s</b>\n\nSelect a preset to change the audio effect.", formatEQName(current)), &tg.SendOptions{
		ParseMode:   "HTML",
		ReplyMarkup: buildEQKeyboard(chatID),
	})
	return err
}

func buildEQKeyboard(chatID int64) *tg.ReplyInlineMarkup {
	current := cache.ChatCache.GetEQPreset(chatID)
	if current == "" {
		current = "normal"
	}

	kb := tg.NewKeyboard()

	var row []tg.KeyboardButton
	for i, eq := range vc.EQs {
		btnText := eq.Label
		if eq.Name == current {
			btnText = "✅ " + btnText
		}
		row = append(row, tg.Button.Data(btnText, "eq_"+eq.Name))

		if len(row) == 2 || i == len(vc.EQs)-1 {
			kb.AddRow(row...)
			row = nil
		}
	}

	kb.AddRow(tg.Button.Data("Close", "eq_close"))
	return kb.Build()
}
