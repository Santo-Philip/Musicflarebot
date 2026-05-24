package dl

import (
	"time"

	"musicflarebot/internal/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
)

type directLink struct {
	query string
}

func newDirectLink(query string) *directLink {
	return &directLink{query: query}
}

// ssValid checks if the query looks like a valid URL.
func (d *directLink) isValid() bool {
	return strings.HasPrefix(d.query, "http://") || strings.HasPrefix(d.query, "https://")
}

func (d *directLink) getInfo() (types.PlatformTracks, error) {
	if !d.isValid() {
		return types.PlatformTracks{}, errors.New("invalid url")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		d.query,
	)

	output, err := cmd.Output()
	if err != nil {
		return types.PlatformTracks{}, fmt.Errorf("invalid or unplayable link: %w", err)
	}

	var info types.FFProbeFormat
	if err = json.Unmarshal(output, &info); err != nil {
		return types.PlatformTracks{}, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	duration := 0
	if info.Format.Duration != "" {
		if d, err := strconv.ParseFloat(info.Format.Duration, 64); err == nil {
			duration = int(d)
		}
	}

	title := info.Format.Tags.Title
	if title == "" {
		parts := strings.Split(d.query, "/")
		if len(parts) > 0 {
			title = parts[len(parts)-1]
			title = strings.SplitN(title, "?", 2)[0]
			title = strings.SplitN(title, "#", 2)[0]
			title, _ = url.QueryUnescape(title)
		}
		if title == "" {
			title = "Direct Link"
		}
	}

	const maxTitleLength = 30
	if len(title) > maxTitleLength {
		title = title[:maxTitleLength-3] + "..."
	}

	track := types.MusicTrack{
		Title:    title,
		Duration: duration,
		Url:      d.query,
		Id:       d.query,
		Platform: types.DirectLink,
	}

	return types.PlatformTracks{Results: []types.MusicTrack{track}}, nil
}

func (d *directLink) search() (types.PlatformTracks, error) {
	return d.getInfo()
}

func (d *directLink) getTrack() (types.TrackInfo, error) {
	info, err := d.getInfo()
	if err != nil {
		return types.TrackInfo{}, err
	}

	if len(info.Results) == 0 {
		return types.TrackInfo{}, errors.New("no track found")
	}

	track := info.Results[0]
	return types.TrackInfo{
		Id:       track.Id,
		URL:      track.Url,
		CdnURL:   track.Url,
		Platform: track.Platform,
	}, nil
}

func (d *directLink) downloadTrack(_ types.TrackInfo, _ bool) (string, error) {
	return d.query, nil
}
