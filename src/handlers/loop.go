package handlers

import (
	"fmt"
	"strconv"

	"musicflarebot/src/core/cache"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func loopHandler(m *tg.NewMessage) error {
	if !adminMode(m) {
		return nil
	}

	chatID := m.ChatID()

	if !cache.ChatCache.IsActive(chatID) {
		_, err := m.Reply("There is no active playback in the video chat.")
		return err
	}

	args := Args(m)
	if args == "" {
		_, err := m.Reply("<b>Loop Control</b>\n\n<b>Usage:</b> <code>/loop [count]</code>\n0 to disable looping\n1-10 to set the number of repeats", &tg.SendOptions{ParseMode: "HTML"})
		return err
	}

	argsInt, err := strconv.Atoi(args)
	if err != nil {
		_, _ = m.Reply("Invalid loop value. Please provide a number between 0 and 10.")
		return nil
	}

	if argsInt < 0 || argsInt > 10 {
		_, err = m.Reply("Loop count must be between 0 and 10.")
		return err
	}

	cache.ChatCache.SetLoopCount(chatID, argsInt)

	var action string
	if argsInt == 0 {
		action = "Looping has been disabled"
	} else {
		action = fmt.Sprintf("Looping has been set to %d time(s)", argsInt)
	}

	_, err = m.Reply(fmt.Sprintf("%s.\nChanged by: %s", action, firstName(m)))
	return err
}
