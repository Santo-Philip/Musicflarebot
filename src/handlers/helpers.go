package handlers

import (
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func entityText(text string, entity tg.MessageEntity) string {
	switch e := entity.(type) {
	case *tg.MessageEntityURL:
		return text[e.Offset : e.Offset+e.Length]
	case *tg.MessageEntityTextURL:
		return e.URL
	}
	return ""
}

func getUrl(m *tg.NewMessage, isReply bool) string {
	text := m.Text()
	entities := m.Message.Entities

	if isReply {
		reply, err := getReplyMessage(m)
		if err == nil && reply != nil {
			text = reply.Text()
			entities = reply.Message.Entities
		}
	}

	for _, entity := range entities {
		if url := entityText(text, entity); url != "" {
			return url
		}
	}

	return ""
}

func isValidMedia(m *tg.NewMessage) bool {
	if m == nil {
		return false
	}

	if m.Audio() != nil || m.Voice() != nil || m.Video() != nil {
		return true
	}

	if doc := m.Document(); doc != nil {
		mime := strings.ToLower(doc.MimeType)
		if strings.HasPrefix(mime, "audio/") || strings.HasPrefix(mime, "video/") {
			return true
		}
	}

	return false
}

func documentName(doc *tg.DocumentObj) string {
	for _, attr := range doc.Attributes {
		switch a := attr.(type) {
		case *tg.DocumentAttributeAudio:
			if a.Title != "" {
				return a.Title
			}
		case *tg.DocumentAttributeFilename:
			return a.FileName
		case *tg.DocumentAttributeVideo:
		}
	}
	return ""
}

func getFileInfo(m *tg.NewMessage) (string, int64) {
	if m == nil {
		return "", 0
	}

	if audio := m.Audio(); audio != nil {
		name := documentName(audio)
		if name == "" {
			name = "audio"
		}
		return name, audio.Size
	}
	if voice := m.Voice(); voice != nil {
		return "voice_note.ogg", voice.Size
	}
	if video := m.Video(); video != nil {
		name := documentName(video)
		if name == "" {
			name = "video"
		}
		return name, video.Size
	}
	if doc := m.Document(); doc != nil {
		name := documentName(doc)
		if name == "" {
			name = "file"
		}
		return name, doc.Size
	}

	return "", 0
}

func coalesce(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
