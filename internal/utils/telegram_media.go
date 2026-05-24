package utils

import (
	tg "github.com/amarnathcjd/gogram/telegram"
)

func GetFileDur(m *tg.NewMessage) int {
	doc := m.Document()
	if doc == nil {
		return 0
	}
	for _, attr := range doc.Attributes {
		switch a := attr.(type) {
		case *tg.DocumentAttributeAudio:
			return int(a.Duration)
		case *tg.DocumentAttributeVideo:
			return int(a.Duration)
		}
	}
	return 0
}
