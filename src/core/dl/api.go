package dl

import (
	"time"

	"musicflarebot/internal/config"
	"musicflarebot/internal/types"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// apiData provides a unified interface for fetching track and playlist information from various music platforms via an API gateway.
type apiData struct {
	Query    string
	ApiUrl   string
	APIKey   string
	Patterns map[string]*regexp.Regexp
}

var apiPatterns = map[string]*regexp.Regexp{
	types.Apple:      regexp.MustCompile(`(?i)^https?:\/\/music\.apple\.com\/[a-zA-Z-]+\/(?:song\/(?:[^\/]+\/)?\d+|album\/[^\/]+\/\d+(?:\?i=\d+)?|playlist\/[^\/]+\/pl\.[\w.-]+|artist\/[^\/]+\/\d+)(?:\?.*)?$`),
	types.Spotify:    regexp.MustCompile(`(?i)^(https?://)?([a-z0-9-]+\.)*spotify\.com/(track|playlist|album|artist)/[a-zA-Z0-9]+(\?.*)?$`),
	types.JioSaavn:   regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?jiosaavn\.com\/(song|album|playlist|featured)\/[^\/]+\/([A-Za-z0-9_]+)`),
	types.Deezer:     regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?deezer\.com\/(?:[a-z]{2}\/)?(track|album|playlist)\/(\d+)`),
	types.SoundCloud: regexp.MustCompile(`(?i)^(https?://)?(www\.)?soundcloud\.com/[a-zA-Z0-9_-]+/(sets/)?[a-zA-Z0-9._-]+(\?.*)?$`),
	types.Gaana:      regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?gaana\.com\/(song|album|playlist|artist)\/([A-Za-z0-9\-]+)`),
	types.Tidal:      regexp.MustCompile(`(?i)https?:\/\/(?:www\.|listen\.)?tidal\.com\/(?:browse\/)?(track|album|playlist)\/([a-zA-Z0-9-]+)(?:[\/?].*)?`),
	types.MXPlayer:   regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?mxplayer\.in\/(?:show|movie)\/.*`),
	types.Twitch:     regexp.MustCompile(`(?i)https?:\/\/(?:www\.|m\.)?twitch\.tv\/(?:videos|[\w._-]+\/video)\/\d+`),
	types.TwitchClip: regexp.MustCompile(
		`(?i)https?:\/\/(?:www\.|m\.)?(?:` +
			`twitch\.tv\/clip\/[\w-]+|` +
			`clips\.twitch\.tv\/[\w-]+|` +
			`twitch\.tv\/[\w-]+\/clip\/[\w-]+` +
			`)`,
	),
	types.Kick:     regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?kick\.com\/[\w._-]+\/videos\/[a-fA-F0-9-]+`),
	types.KickClip: regexp.MustCompile(`(?i)https?:\/\/(?:www\.)?kick\.com\/[\w._-]+\/clips\/[\w-]+`),
}

// newApiData creates and initializes a new apiData instance with the provided query.
func newApiData(query string) *apiData {
	return &apiData{
		Query:    strings.TrimSpace(query),
		ApiUrl:   strings.TrimRight(config.Conf.ApiUrl, "/"),
		APIKey:   config.Conf.ApiKey,
		Patterns: apiPatterns,
	}
}

func (a *apiData) isValid() bool {
	if a.Query == "" || a.ApiUrl == "" || a.APIKey == "" {
		return false
	}

	for _, pattern := range a.Patterns {
		if pattern.MatchString(a.Query) {
			return true
		}
	}
	return false
}

// getInfo retrieves metadata for a track or playlist from the API.
func (a *apiData) getInfo() (types.PlatformTracks, error) {
	if !a.isValid() {
		return types.PlatformTracks{}, errors.New("the provided URL is invalid or the platform is not supported")
	}

	fullURL := fmt.Sprintf("%s/api/get_url?%s", a.ApiUrl, url.Values{"url": {a.Query}}.Encode())
	resp, err := sendRequest(http.MethodGet, fullURL, nil, map[string]string{"X-API-Key": a.APIKey})
	if err != nil {
		return types.PlatformTracks{}, fmt.Errorf("the GetInfo request failed: %w", err)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return types.PlatformTracks{}, fmt.Errorf("unexpected status code while fetching info: %s", resp.Status)
	}

	var data types.PlatformTracks
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return types.PlatformTracks{}, fmt.Errorf("failed to decode the GetInfo response: %w", err)
	}
	return data, nil
}

// search queries the API for a track.
func (a *apiData) search() (types.PlatformTracks, error) {
	if a.isValid() {
		return a.getInfo()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	fullURL := fmt.Sprintf("%s/api/search?%s", a.ApiUrl, url.Values{
		"query": {a.Query},
		"limit": {"5"},
	}.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return types.PlatformTracks{}, fmt.Errorf("failed to create the search request: %w", err)
	}
	req.Header.Set("X-API-Key", a.APIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.PlatformTracks{}, fmt.Errorf("the search request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return types.PlatformTracks{}, fmt.Errorf("unexpected status code during search: %s", resp.Status)
	}

	var data types.PlatformTracks
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return types.PlatformTracks{}, fmt.Errorf("failed to decode the search response: %w", err)
	}
	return data, nil
}

// getTrack retrieves detailed information for a single track from the API.
func (a *apiData) getTrack() (types.TrackInfo, error) {
	fullURL := fmt.Sprintf("%s/api/track?%s", a.ApiUrl, url.Values{"url": {a.Query}}.Encode())
	resp, err := sendRequest(http.MethodGet, fullURL, nil, map[string]string{"X-API-Key": a.APIKey})
	if err != nil {
		return types.TrackInfo{}, fmt.Errorf("the GetTrack request failed: %w", err)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return types.TrackInfo{}, fmt.Errorf("unexpected status code while fetching the track: %s", resp.Status)
	}

	var data types.TrackInfo
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return types.TrackInfo{}, fmt.Errorf("failed to decode the GetTrack response: %w", err)
	}

	return data, nil
}

// downloadTrack downloads a track using the API. If the track is a YouTube video and video format is requested,
func (a *apiData) downloadTrack(info types.TrackInfo, video bool) (string, error) {
	// if the track is from YouTube and video:true
	yt := newYouTubeData(a.Query)
	if info.Platform == types.YouTube && video {
		return yt.downloadTrack(info, video)
	}

	downloader, err := newDownload(info)
	if err != nil {
		return "", fmt.Errorf("failed to initialize the download: %w", err)
	}

	filePath, err := downloader.Process()
	if err != nil {
		if info.Platform == types.YouTube {
			return yt.downloadTrack(info, video)
		}
		return "", fmt.Errorf("the download process failed: %w", err)
	}

	if strings.Contains(a.ApiUrl, filePath) {
		return downloadFile(filePath, "", false)
	}

	return filePath, nil
}
