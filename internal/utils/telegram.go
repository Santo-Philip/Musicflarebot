package utils

import (
	"fmt"
	"strconv"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func GetMessage(c *tg.Client, link string) (*tg.NewMessage, error) {
	parsed, err := ParseMessageLink(link)
	if err != nil {
		return nil, err
	}
	return c.GetMessageByID(parsed.ChatID, parsed.MsgID)
}

type MessageLink struct {
	ChatID int64
	MsgID  int32
}

func ParseMessageLink(link string) (*MessageLink, error) {
	link = strings.TrimSpace(link)
	if strings.HasPrefix(link, "https://t.me/c/") {
		parts := strings.Split(link, "/")
		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid Telegram link")
		}
		chatIDStr := strings.TrimPrefix(parts[3], "c/")
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid chat ID in link: %w", err)
		}
		msgID, err := strconv.Atoi(parts[4])
		if err != nil {
			return nil, fmt.Errorf("invalid message ID in link: %w", err)
		}
		return &MessageLink{ChatID: -1000000000000 - chatID, MsgID: int32(msgID)}, nil
	}
	return nil, fmt.Errorf("unsupported Telegram link format")
}
