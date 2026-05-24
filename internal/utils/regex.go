package utils

import "regexp"

var TelegramMessageRegex = regexp.MustCompile(`https://t\.me/(?:c/)?(\d+|[\w_]+)/(\d+)`)
