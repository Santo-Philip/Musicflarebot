package handlers

import (
	"fmt"

	"musicflarebot/src/core"
	"musicflarebot/src/core/cache"
	"musicflarebot/src/vc"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func pauseHandler(m *tg.NewMessage) error {
	if !adminMode(m) {
		return nil
	}

	chatID := m.ChatID()

	if !cache.ChatCache.IsActive(chatID) {
		_, _ = m.Reply("There is no active playback in the video chat.")
		return nil
	}

	if _, err := vc.Calls.Pause(chatID); err != nil {
		_, _ = m.Reply(fmt.Sprintf("Failed to pause the playback: %s", err.Error()))
		return nil
	}

	_, err := m.Reply(fmt.Sprintf("Playback has been paused by %s.", firstName(m)), &tg.SendOptions{ReplyMarkup: core.ControlButtons("pause")})
	return err
}

func resumeHandler(m *tg.NewMessage) error {
	if !adminMode(m) {
		return nil
	}

	chatID := m.ChatID()

	if !cache.ChatCache.IsActive(chatID) {
		_, _ = m.Reply("There is no active playback in the video chat.")
		return nil
	}

	if _, err := vc.Calls.Resume(chatID); err != nil {
		_, _ = m.Reply(fmt.Sprintf("Failed to resume the playback: %s", err.Error()))
		return nil
	}

	_, err := m.Reply(fmt.Sprintf("Playback has been resumed by %s.", firstName(m)), &tg.SendOptions{ReplyMarkup: core.ControlButtons("resume")})
	return err
}
