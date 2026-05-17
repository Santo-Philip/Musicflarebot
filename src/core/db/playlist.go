package db

import (
	"musicflarebot/src/utils"
	"crypto/rand"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Song struct {
	URL      string `json:"url"`
	Name     string `json:"name"`
	TrackID  string `json:"track_id"`
	Duration int    `json:"duration"`
	Platform string `json:"platform"`
}

type Playlist struct {
	ID     string
	Name   string
	UserID int64
	Songs  []Song
}

func generateUniquePlaylistID() string {
	b := make([]byte, 5)
	_, _ = rand.Read(b)
	return fmt.Sprintf("tgpl_%x", b)
}

func (db *Database) CreatePlaylist(name string, userID int64) (string, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	id := generateUniquePlaylistID()
	_, err := db.pool.Exec(ctx,
		`INSERT INTO playlists (id, name, user_id, songs) VALUES ($1, $2, $3, '[]'::jsonb)`,
		id, name, userID,
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (db *Database) GetPlaylist(id string) (*Playlist, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	var p Playlist
	var songsJSON []byte
	err := db.pool.QueryRow(ctx,
		`SELECT id, name, user_id, songs FROM playlists WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.UserID, &songsJSON)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("playlist not found")
		}
		return nil, err
	}

	if err := json.Unmarshal(songsJSON, &p.Songs); err != nil {
		p.Songs = []Song{}
	}

	return &p, nil
}

func (db *Database) DeletePlaylist(id string, userID int64) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`DELETE FROM playlists WHERE id = $1 AND user_id = $2`, id, userID,
	)
	return err
}

func (db *Database) songExists(id string, trackID string) bool {
	ctx, cancel := db.ctx()
	defer cancel()

	var songsJSON []byte
	err := db.pool.QueryRow(ctx,
		`SELECT songs FROM playlists WHERE id = $1`, id,
	).Scan(&songsJSON)
	if err != nil {
		return false
	}

	var songs []Song
	if err := json.Unmarshal(songsJSON, &songs); err != nil {
		return false
	}

	for _, song := range songs {
		if song.TrackID == trackID {
			return true
		}
	}
	return false
}

func (db *Database) AddSongToPlaylist(id string, song Song) error {
	if db.songExists(id, song.TrackID) {
		return nil
	}

	ctx, cancel := db.ctx()
	defer cancel()

	songJSON, _ := json.Marshal(song)
	_, err := db.pool.Exec(ctx,
		`UPDATE playlists SET songs = songs || $1::jsonb WHERE id = $2`,
		[]byte(fmt.Sprintf("[%s]", string(songJSON))), id,
	)
	return err
}

func (db *Database) RemoveSongFromPlaylist(id string, trackID string) error {
	if !db.songExists(id, trackID) {
		return fmt.Errorf("track with ID %s not found in playlist", trackID)
	}

	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`UPDATE playlists SET songs = (
			SELECT jsonb_agg(elem) FROM jsonb_array_elements(songs) elem
			WHERE elem->>'track_id' != $2
		) WHERE id = $1`,
		id, trackID,
	)
	if err != nil {
		return fmt.Errorf("error removing song: %w", err)
	}
	return nil
}

func (db *Database) GetUserPlaylists(userID int64) ([]Playlist, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	rows, err := db.pool.Query(ctx,
		`SELECT id, name, user_id, songs FROM playlists WHERE user_id = $1`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []Playlist
	for rows.Next() {
		var p Playlist
		var songsJSON []byte
		if err := rows.Scan(&p.ID, &p.Name, &p.UserID, &songsJSON); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(songsJSON, &p.Songs); err != nil {
			p.Songs = []Song{}
		}
		playlists = append(playlists, p)
	}
	return playlists, rows.Err()
}

func ConvertSongsToTracks(songs []Song) []utils.MusicTrack {
	tracks := make([]utils.MusicTrack, 0, len(songs))
	for _, song := range songs {
		tracks = append(tracks, utils.MusicTrack{
			Url:      song.URL,
			Title:    song.Name,
			Id:       song.TrackID,
			Duration: song.Duration,
			Platform: song.Platform,
		})
	}
	return tracks
}
