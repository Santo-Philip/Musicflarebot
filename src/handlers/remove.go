package handlers

import (
	"fmt"
	"strconv"

	"musicflarebot/src/core/cache"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func removeHandler(m *tg.NewMessage) error {
	if !adminMode(m) {
		return nil
	}

	chatID := m.ChatID()

	if !cache.ChatCache.IsActive(chatID) {
		_, _ = m.Reply("The bot is not streaming in the video chat.")
		return nil
	}

	queue := cache.ChatCache.GetQueue(chatID)
	if len(queue) == 0 {
		_, _ = m.Reply("The queue is currently empty.")
		return nil
	}

	args := Args(m)
	if args == "" {
		_, _ = m.Reply("<b>Usage:</b> <code>/remove [track number]</code>\n\nUse <code>1</code> to remove the first track, <code>2</code> for the second, and so on.", replyOpts)
		return nil
	}

	trackNum, err := strconv.Atoi(args)
	if err != nil {
		_, _ = m.Reply("Please provide a valid track number.")
		return nil
	}

	if trackNum <= 0 || trackNum > len(queue) {
		_, _ = m.Reply(fmt.Sprintf("Invalid track number. Please choose a number between 1 and %d.", len(queue)))
		return nil
	}

	cache.ChatCache.RemoveTrack(chatID, trackNum)
	_, err = m.Reply(fmt.Sprintf("Track #%d has been removed by %s.", trackNum, firstName(m)), replyOpts)
	return err
}
