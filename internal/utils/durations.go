package utils

import (
	"context"
	"encoding/json"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"musicflarebot/internal/types"
)

func SecToMin(seconds int) string {
	if seconds <= 0 {
		return "LIVE"
	}
	d := time.Duration(seconds) * time.Second
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return strconv.Itoa(h) + ":" + strconv.Itoa(m) + ":" + strconv.Itoa(s)
	}
	return strconv.Itoa(m) + ":" + strconv.Itoa(s)
}

func GetMediaDuration(filePath string) int {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "json",
		filePath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}

	var f types.FFProbeFormat
	if err := json.Unmarshal(out, &f); err != nil {
		return 0
	}

	dur, err := strconv.ParseFloat(strings.TrimSpace(f.Format.Duration), 64)
	if err != nil {
		return 0
	}

	slog.Debug("media duration", "file", filePath, "duration", dur)
	return int(dur)
}
