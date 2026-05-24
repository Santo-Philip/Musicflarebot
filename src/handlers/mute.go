package handlers

import (
	"fmt"

	"musicflarebot/internal/ui"
	"musicflarebot/internal/cache"
	"musicflarebot/src/vc"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func muteHandler(m *tg.NewMessage) error {
	if !adminMode(m) {
		return nil
	}

	if args := Args(m); args != "" {
		return nil
	}

	chatID := m.ChatID()
	if !cache.ChatCache.IsActive(chatID) {
		_, err := m.Reply("There is no active playback in the video chat.")
		return err
	}

	if _, err := vc.Calls.Mute(chatID); err != nil {
		_, err = m.Reply(fmt.Sprintf("Failed to mute the playback: %s", err.Error()))
		return err
	}

	_, err := m.Reply(fmt.Sprintf("Playback has been muted by %s.", firstName(m)), &tg.SendOptions{ReplyMarkup: ui.ControlButtons("mute")})
	return err
}

func unmuteHandler(m *tg.NewMessage) error {
	if !adminMode(m) {
		return nil
	}

	if args := Args(m); args != "" {
		return nil
	}

	chatID := m.ChatID()
	if !cache.ChatCache.IsActive(chatID) {
		_, err := m.Reply("There is no active playback in the video chat.")
		return err
	}

	if _, err := vc.Calls.Unmute(chatID); err != nil {
		_, err = m.Reply(fmt.Sprintf("Failed to unmute the playback: %s", err.Error()))
		return err
	}

	_, err := m.Reply(fmt.Sprintf("Playback has been unmuted by %s.", firstName(m)), &tg.SendOptions{ReplyMarkup: ui.ControlButtons("unmute")})
	return err
}
