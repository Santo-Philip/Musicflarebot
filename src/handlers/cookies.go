package handlers

import (
	"fmt"
	"path/filepath"
	"time"

	"musicflarebot/config"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func setCookiesHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	if m.ReplyToMsgID() == 0 {
		_, _ = m.Reply("Reply to a cookies.txt file with /setcookies.")
		return nil
	}

	reply, err := getReplyMessage(m)
	if err != nil {
		_, _ = m.Reply("Could not fetch replied message.")
		return nil
	}

	if reply.Document() == nil {
		_, _ = m.Reply("Replied message is not a file. Please reply to a cookies.txt file.")
		return nil
	}

	fileName := fmt.Sprintf("cookies_%d.txt", time.Now().UnixMilli())
	dest := filepath.Join("src/cookies", fileName)

	_, err = reply.Download(&tg.DownloadOptions{
		FileName: dest,
	})
	if err != nil {
		_, _ = m.Reply(fmt.Sprintf("Download failed: %s", err.Error()))
		return nil
	}

	absPath, _ := filepath.Abs(dest)
	config.Conf.CookiesPath = append(config.Conf.CookiesPath, absPath)

	_, _ = m.Reply(fmt.Sprintf("Cookies file saved to:\n<code>%s</code>\n\nTotal cookie files: %d", absPath, len(config.Conf.CookiesPath)), &tg.SendOptions{
		ParseMode: "HTML",
	})

	return nil
}
