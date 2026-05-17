package handlers

import (
	"fmt"
	"musicflarebot/src/core/db"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
)

var (
	broadcastCancelFlag atomic.Bool
	broadcastInProgress atomic.Bool
)

func getFloodWait(err error) int {
	if err == nil {
		return 0
	}
	return 0
}

func cancelBroadcastHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}
	if !broadcastInProgress.Load() {
		_, _ = m.Reply("No broadcast in progress.")
		return nil
	}

	broadcastCancelFlag.Store(true)
	_, _ = m.Reply("Broadcast stopped.")
	return nil
}

func broadcastHandler(m *tg.NewMessage) error {
	if !isDev(m) {
		return nil
	}

	if broadcastInProgress.Load() {
		_, _ = m.Reply("A broadcast is already in progress.")
		return nil
	}

	reply, err := getReplyMessage(m)
	if err != nil {
		usage := `Please reply to a message to broadcast.

Usage:
-chat  : groups only
-user  : users only
-both  : groups + users (default)
-copy  : send as copy

Examples:
/broadcast
/broadcast -chat
/broadcast -user -copy
`

		_, _ = m.Reply(usage)
		return nil
	}

	args := strings.Fields(Args(m))

	copyMode := false
	mode := "both"

	for _, a := range args {
		switch a {
		case "-copy":
			copyMode = true
		case "-chat":
			mode = "chat"
		case "-user":
			mode = "user"
		case "-both":
			mode = "both"
		}
	}

	chats, _ := db.Instance.GetAllChats()
	users, _ := db.Instance.GetAllUsers()

	groupsMap := make(map[int64]bool)
	for _, id := range chats {
		groupsMap[id] = true
	}

	var targets []int64

	switch mode {
	case "chat":
		targets = append(targets, chats...)
	case "user":
		targets = append(targets, users...)
	case "both":
		targets = append(targets, chats...)
		targets = append(targets, users...)
	}

	if len(targets) == 0 {
		_, _ = m.Reply("No targets found.")
		return nil
	}

	broadcastCancelFlag.Store(false)
	broadcastInProgress.Store(true)

	sentMsg, _ := m.Reply("Broadcast started.")

	go func() {
		defer broadcastInProgress.Store(false)

		var failedBuilder strings.Builder
		count, ucount := 0, 0

		for _, chatID := range targets {
			if broadcastCancelFlag.Load() {
				_, _ = sentMsg.Edit(
					fmt.Sprintf("Broadcast stopped.\nGroups: %d\nUsers: %d", count, ucount),
				)
				return
			}

			var errSend error
			if copyMode {
				_, errSend = reply.ForwardTo(chatID, &tg.ForwardOptions{HideAuthor: true})
			} else {
				_, errSend = reply.ForwardTo(chatID)
			}

			if errSend == nil {
				if groupsMap[chatID] {
					count++
				} else {
					ucount++
				}
				time.Sleep(200 * time.Millisecond)
			} else {
				wait := getFloodWait(errSend)
				if wait > 0 {
					time.Sleep(time.Duration(wait+30) * time.Second)
					continue
				}
				failedBuilder.WriteString(fmt.Sprintf("%d - %v\n", chatID, errSend))
			}
		}

		text := fmt.Sprintf("Broadcast ended.\nGroups: %d\nUsers: %d", count, ucount)
		failedStr := failedBuilder.String()

		if failedStr != "" {
			errFile := filepath.Join(
				os.TempDir(),
				fmt.Sprintf("errors_%d.txt", time.Now().UnixNano()),
			)

			if err := os.WriteFile(errFile, []byte(failedStr), 0644); err == nil {
				defer os.Remove(errFile)

				_, errSendDoc := m.ReplyMedia(errFile, &tg.MediaOptions{
					ForceDocument: true,
					Caption:       text,
				})

				if errSendDoc != nil {
					_, _ = sentMsg.Edit(text)
				}
			} else {
				_, _ = sentMsg.Edit(text)
			}
		} else {
			_, _ = sentMsg.Edit(text)
		}
	}()

	return nil
}
