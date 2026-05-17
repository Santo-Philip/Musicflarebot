package utils

import "regexp"

var (
	TelegramMessageRegex = regexp.MustCompile(`^https://t\.me/(?:([a-zA-Z0-9_]{4,})|c/(\d+))/(\d+)(?:\?.*)?$`)
)
