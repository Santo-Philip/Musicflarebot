package handlers

import (
	"fmt"
	"musicflarebot/internal/utils"
	"strconv"

	"musicflarebot/internal/cache"
	"musicflarebot/src/vc"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func seekHandler(m *tg.NewMessage) error {
	if !adminMode(m) {
		return nil
	}
	chatID := m.ChatID()

	if !cache.ChatCache.IsActive(chatID) {
		_, err := m.Reply("The bot is not streaming in the video chat.")
		return err
	}

	playingSong := cache.ChatCache.GetPlayingTrack(chatID)
	if playingSong == nil {
		_, err := m.Reply("The bot is not streaming in the video chat.")
		return err
	}

	args := Args(m)
	if args == "" {
		_, _ = m.Reply("<b>Usage:</b> /seek duration\n<b>Example:</b> <code>/seek 15</code>", replyOpts)
		return nil
	}

	seekTime, err := strconv.Atoi(args)
	if err != nil {
		_, _ = m.Reply("Invalid seek time provided. Please use a valid number of seconds.")
		return nil
	}

	if seekTime < 10 {
		_, _ = m.Reply("Minimum seek time is 10 seconds.")
		return nil
	}

	currDur, err := vc.Calls.PlayedTime(chatID)
	if err != nil {
		_, _ = m.Reply("Failed to fetch the duration of the ongoing stream.")
		return nil
	}

	toSeek := int(currDur) + seekTime
	if toSeek >= playingSong.Duration {
		_, _ = m.Reply(fmt.Sprintf("You cannot seek beyond the track duration. Maximum allowed is %s.", utils.SecToMin(playingSong.Duration)))
		return nil
	}

	if err = vc.Calls.SeekStream(
		chatID,
		playingSong.FilePath,
		toSeek,
		playingSong.Duration,
		playingSong.IsVideo,
	); err != nil {
		_, _ = m.Reply(fmt.Sprintf("An error occurred while seeking the track: %s", err.Error()), replyOpts)
		return nil
	}

	_, _ = m.Reply(fmt.Sprintf("<b>Stream skipped %s and started from %s seconds by</b> %s", utils.SecToMin(seekTime), utils.SecToMin(toSeek), firstName(m)), replyOpts)
	return nil
}
