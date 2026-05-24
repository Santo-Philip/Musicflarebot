package types

type CachedTrack struct {
	URL       string `json:"url"`
	Name      string `json:"name"`
	Loop      int    `json:"loop"`
	User      string `json:"user"`
	UserID    int64  `json:"user_id"`
	FilePath  string `json:"file_path"`
	Thumbnail string `json:"thumbnail"`
	TrackID   string `json:"track_id"`
	Duration  int    `json:"duration"`
	Channel   string `json:"channel"`
	Views     string `json:"views"`
	IsVideo   bool   `json:"is_video"`
	Platform  string `json:"platform"`
}

type TrackInfo struct {
	Id       string `json:"id"`
	URL      string `json:"url"`
	CdnURL   string `json:"cdnurl"`
	Key      string `json:"key"`
	Platform string `json:"platform"`
}

type MusicTrack struct {
	Title     string `json:"title"`
	Id        string `json:"id"`
	Url       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Duration  int    `json:"duration"`
	Channel   string `json:"channel"`
	Views     string `json:"views"`
	Platform  string `json:"platform"`
}

type PlatformTracks struct {
	Results []MusicTrack `json:"results"`
}

type FFProbeFormat struct {
	Format struct {
		Duration string `json:"duration"`
		Tags     struct {
			Title string `json:"title"`
		} `json:"tags,omitempty"`
	} `json:"format"`
}
