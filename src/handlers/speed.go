package handlers

import (
	"fmt"
	"strconv"

	"musicflarebot/src/core/cache"
	"musicflarebot/src/vc"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func speedHandler(m *tg.NewMessage) error {
	if !adminMode(m) {
		return nil
	}
	chatID := m.ChatID()

	if !cache.ChatCache.IsActive(chatID) {
		_, err := m.Reply("The bot is not streaming in the video chat.")
		return err
	}

	if playingSong := cache.ChatCache.GetPlayingTrack(chatID); playingSong == nil {
		_, err := m.Reply("The bot is not streaming in the video chat.")
		return err
	}

	args := Args(m)
	if args == "" {
		_, _ = m.Reply("<b>Change Playback Speed</b>\n\n<b>Usage:</b> <code>/speed [value]</code>\n\nThe speed can be set between <code>0.5</code> and <code>4.0</code>.", replyOpts)
		return nil
	}

	speed, err := strconv.ParseFloat(args, 64)
	if err != nil {
		_, _ = m.Reply("Invalid speed value. Please provide a number between 0.5 and 4.0.")
		return nil
	}

	if err = vc.Calls.ChangeSpeed(chatID, speed); err != nil {
		_, _ = m.Reply(fmt.Sprintf("An error occurred while changing the speed: %s", err.Error()), replyOpts)
		return nil
	}

	_, _ = m.Reply(fmt.Sprintf("Playback speed has been set to <code>%.2fx</code>.", speed), replyOpts)
	return nil
}
