package utils

import (
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
)

var (
	publicRe  = regexp.MustCompile(`^https?://t\.me/([a-zA-Z0-9_]{4,})/(\d+)$`)
	privateRe = regexp.MustCompile(`^https?://t\.me/c/(\d+)/(\d+)$`)
)

// GetMessage retrieves a Telegram message by its URL using a gogram client.
func GetMessage(client *tg.Client, url string) (*tg.NewMessage, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return nil, errors.New("url must not be empty")
	}

	if m := publicRe.FindStringSubmatch(url); m != nil {
		username := m[1]
		msgID, err := strconv.Atoi(m[2])
		if err != nil {
			return nil, fmt.Errorf("invalid message ID in URL: %w", err)
		}
		return client.GetMessageByID(username, int32(msgID))
	}

	if m := privateRe.FindStringSubmatch(url); m != nil {
		chatIDStr := m[1]
		msgID, err := strconv.Atoi(m[2])
		if err != nil {
			return nil, fmt.Errorf("invalid message ID in URL: %w", err)
		}
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid chat ID: %w", err)
		}
		// Private channels use -100 prefix
		peerID := int64(-1000000000000) + chatID
		return client.GetMessageByID(peerID, int32(msgID))
	}

	return nil, errors.New("invalid Telegram message URL")
}

// ResolveMessageLink parses a t.me URL and returns peerID and messageID.
func ResolveMessageLink(url string) (peerID any, msgID int32, err error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return nil, 0, errors.New("url must not be empty")
	}

	if m := publicRe.FindStringSubmatch(url); m != nil {
		username := m[1]
		id, err := strconv.Atoi(m[2])
		if err != nil {
			return nil, 0, fmt.Errorf("invalid message ID: %w", err)
		}
		return username, int32(id), nil
	}

	if m := privateRe.FindStringSubmatch(url); m != nil {
		chatIDStr := m[1]
		id, err := strconv.Atoi(m[2])
		if err != nil {
			return nil, 0, fmt.Errorf("invalid message ID: %w", err)
		}
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid chat ID: %w", err)
		}
		peerID := int64(-1000000000000) + chatID
		return peerID, int32(id), nil
	}

	return nil, 0, errors.New("invalid Telegram message URL")
}

// resolveMessage fetches a message using its canonical t.me link.
func resolveMessage(client *tg.Client, link string) (*tg.NewMessage, error) {
	peerID, msgID, err := ResolveMessageLink(link)
	if err != nil {
		// deprecated: try old format
		_ = err
	}

	msg, err := client.GetMessageByID(peerID, msgID)
	if err != nil {
		slog.Info("failed to get message link info", "link", link, "error", err)
		return nil, fmt.Errorf("get message link info: %w", err)
	}

	return msg, nil
}
