package database

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"musicflarebot/internal/cache"
	"musicflarebot/internal/types"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

type Database struct {
	pool *pgxpool.Pool

	chatCache      *cache.Cache[string]
	userCache      *cache.Cache[string]
	assistantCache *cache.Cache[string]
	authCache      *cache.Cache[string]
	langCache      *cache.Cache[string]
	loggerCache    *cache.Cache[string]
	blChatsCache   *cache.Cache[string]
	blUsersCache   *cache.Cache[string]
}

var Instance *Database

func Connect(databaseUrl string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseUrl)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	Instance = &Database{
		pool:            pool,
		chatCache:       cache.New[string](20 * time.Minute),
		userCache:       cache.New[string](20 * time.Minute),
		assistantCache:  cache.New[string](20 * time.Minute),
		authCache:       cache.New[string](20 * time.Minute),
		langCache:       cache.New[string](20 * time.Minute),
		loggerCache:     cache.New[string](20 * time.Minute),
		blChatsCache:    cache.New[string](20 * time.Minute),
		blUsersCache:    cache.New[string](20 * time.Minute),
	}

	if err := Instance.migrate(); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}

	slog.Info("[DB] PostgreSQL connection established successfully.")
	return nil
}

func Close() {
	if Instance != nil {
		Instance.pool.Close()
	}
}

func (db *Database) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS chats (
			id BIGINT PRIMARY KEY,
			play_type INT DEFAULT 0,
			admin_play BOOLEAN DEFAULT FALSE,
			admin_mode TEXT DEFAULT 'everyone',
			cmd_delete BOOLEAN DEFAULT FALSE
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id BIGINT PRIMARY KEY
		)`,
		`CREATE TABLE IF NOT EXISTS assistant_assignments (
			chat_id BIGINT PRIMARY KEY,
			num INT DEFAULT -1
		)`,
		`CREATE TABLE IF NOT EXISTS auth_users (
			chat_id BIGINT,
			user_id BIGINT,
			PRIMARY KEY (chat_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS blacklist (
			type TEXT,
			id BIGINT,
			PRIMARY KEY (type, id)
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value JSONB DEFAULT '{}'::jsonb
		)`,
		`CREATE TABLE IF NOT EXISTS playlists (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			user_id BIGINT NOT NULL,
			songs JSONB DEFAULT '[]'::jsonb
		)`,
		`CREATE TABLE IF NOT EXISTS chat_languages (
			chat_id BIGINT PRIMARY KEY,
			lang TEXT DEFAULT 'en'
		)`,
		`CREATE TABLE IF NOT EXISTS bot_settings (
			s_key TEXT PRIMARY KEY,
			s_value TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS user_stats (
			user_id BIGINT NOT NULL,
			chat_id BIGINT NOT NULL,
			date DATE NOT NULL DEFAULT CURRENT_DATE,
			play_count INTEGER DEFAULT 0,
			UNIQUE(user_id, chat_id, date)
		)`,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, q := range queries {
		if _, err := tx.Exec(ctx, q); err != nil {
			return fmt.Errorf("migration failed on query: %s: %w", q, err)
		}
	}

	return tx.Commit(ctx)
}

func (db *Database) ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

var ErrNotFound = errors.New("not found")

func toKey(id int64) string {
	return fmt.Sprintf("%d", id)
}

func contains(list []int64, id int64) bool {
	for _, v := range list {
		if v == id {
			return true
		}
	}
	return false
}

// ─── Users ────────────────────────────────────────────────

func (db *Database) AddUser(userID int64) error {
	key := toKey(userID)
	if _, ok := db.userCache.Get(key); ok {
		return nil
	}

	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING`, userID,
	)
	if err != nil {
		return err
	}

	db.userCache.Set(key, "1")
	return nil
}

func (db *Database) IsUserExist(userID int64) (bool, error) {
	key := toKey(userID)
	if _, ok := db.userCache.Get(key); ok {
		return true, nil
	}

	ctx, cancel := db.ctx()
	defer cancel()

	var id int64
	err := db.pool.QueryRow(ctx, `SELECT id FROM users WHERE id = $1`, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	db.userCache.Set(key, "1")
	return true, nil
}

func (db *Database) GetAllUsers() ([]int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rows, err := db.pool.Query(ctx, `SELECT id FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		users = append(users, id)
	}
	return users, rows.Err()
}

// ─── Chat Settings ────────────────────────────────────────

func (db *Database) AddChat(chatID int64) error {
	key := toKey(chatID)
	if _, ok := db.chatCache.Get(key); ok {
		return nil
	}

	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO chats (id) VALUES ($1) ON CONFLICT (id) DO NOTHING`, chatID,
	)
	if err == nil {
		slog.Info("[DB] A new chat has been added", "id", chatID)
	}
	return err
}

func (db *Database) GetPlayType(chatID int64) int {
	ctx, cancel := db.ctx()
	defer cancel()

	var playType int
	err := db.pool.QueryRow(ctx,
		`SELECT COALESCE(play_type, 0) FROM chats WHERE id = $1`, chatID,
	).Scan(&playType)
	if err != nil {
		return 0
	}
	return playType
}

func (db *Database) SetPlayType(chatID int64, playType int) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO chats (id, play_type) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET play_type = $2`,
		chatID, playType,
	)
	if err == nil {
		db.chatCache.Delete(toKey(chatID))
	}
	return err
}

func (db *Database) GetPlayMode(chatID int64) bool {
	ctx, cancel := db.ctx()
	defer cancel()

	var adminPlay bool
	err := db.pool.QueryRow(ctx,
		`SELECT COALESCE(admin_play, FALSE) FROM chats WHERE id = $1`, chatID,
	).Scan(&adminPlay)
	if err != nil {
		return false
	}
	return adminPlay
}

func (db *Database) SetPlayMode(chatID int64, adminPlay bool) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO chats (id, admin_play) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET admin_play = $2`,
		chatID, adminPlay,
	)
	if err == nil {
		db.chatCache.Delete(toKey(chatID))
	}
	return err
}

func (db *Database) GetAdminMode(chatID int64) string {
	ctx, cancel := db.ctx()
	defer cancel()

	var mode string
	err := db.pool.QueryRow(ctx,
		`SELECT COALESCE(admin_mode, 'everyone') FROM chats WHERE id = $1`, chatID,
	).Scan(&mode)
	if err != nil {
		return "everyone"
	}
	return mode
}

func (db *Database) SetAdminMode(chatID int64, adminMode string) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO chats (id, admin_mode) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET admin_mode = $2`,
		chatID, adminMode,
	)
	if err == nil {
		db.chatCache.Delete(toKey(chatID))
	}
	return err
}

func (db *Database) GetCmdDelete(chatID int64) bool {
	ctx, cancel := db.ctx()
	defer cancel()

	var cmdDelete bool
	err := db.pool.QueryRow(ctx,
		`SELECT COALESCE(cmd_delete, FALSE) FROM chats WHERE id = $1`, chatID,
	).Scan(&cmdDelete)
	if err != nil {
		return false
	}
	return cmdDelete
}

func (db *Database) SetCmdDelete(chatID int64, cmdDelete bool) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO chats (id, cmd_delete) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET cmd_delete = $2`,
		chatID, cmdDelete,
	)
	if err == nil {
		db.chatCache.Delete(toKey(chatID))
	}
	return err
}

func (db *Database) GetAllChats() ([]int64, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	rows, err := db.pool.Query(ctx, `SELECT id FROM chats`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		chats = append(chats, id)
	}
	return chats, rows.Err()
}

// ─── Bot Settings ─────────────────────────────────────────

func (db *Database) GetSetting(key string) (string, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	var value string
	err := db.pool.QueryRow(ctx,
		`SELECT s_value FROM bot_settings WHERE s_key = $1`, key,
	).Scan(&value)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

func (db *Database) SetSetting(key, value string) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO bot_settings (s_key, s_value) VALUES ($1, $2) ON CONFLICT (s_key) DO UPDATE SET s_value = $2`,
		key, value,
	)
	return err
}

func (db *Database) DeleteSetting(key string) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx, `DELETE FROM bot_settings WHERE s_key = $1`, key)
	return err
}

func (db *Database) GetAllSettings() (map[string]string, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	rows, err := db.pool.Query(ctx, `SELECT s_key, s_value FROM bot_settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		settings[key] = value
	}
	return settings, nil
}

func (db *Database) GetSessionStrings() ([]string, error) {
	all, err := db.GetAllSettings()
	if err != nil {
		return nil, err
	}

	var sessions []string
	for key, val := range all {
		if strings.HasPrefix(key, "session_") && val != "" {
			sessions = append(sessions, val)
		}
	}
	return sessions, nil
}

func (db *Database) GetOwnerLoggerId() int64 {
	val, err := db.GetSetting("logger_id")
	if err != nil || val == "" {
		return 0
	}
	id, _ := strconv.ParseInt(val, 10, 64)
	return id
}

func (db *Database) SetOwnerLoggerId(id int64) error {
	return db.SetSetting("logger_id", strconv.FormatInt(id, 10))
}

func (db *Database) GetOwnerDevs() []int64 {
	val, err := db.GetSetting("devs")
	if err != nil || val == "" {
		return nil
	}
	var devs []int64
	for _, part := range strings.Fields(strings.ReplaceAll(val, ",", " ")) {
		id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err == nil {
			devs = append(devs, id)
		}
	}
	return devs
}

func (db *Database) SetOwnerDevs(devs []int64) error {
	var parts []string
	for _, id := range devs {
		parts = append(parts, strconv.FormatInt(id, 10))
	}
	return db.SetSetting("devs", strings.Join(parts, " "))
}

func (db *Database) GetOwnerSupportGroup() string {
	v, _ := db.GetSetting("support_group")
	return v
}

func (db *Database) GetOwnerSupportChannel() string {
	v, _ := db.GetSetting("support_channel")
	return v
}

func (db *Database) GetOwnerSongDuration() int64 {
	v, err := db.GetSetting("song_duration_limit")
	if err != nil || v == "" {
		return 0
	}
	d, _ := strconv.ParseInt(v, 10, 64)
	return d
}

func (db *Database) GetOwnerDefaultService() string {
	v, _ := db.GetSetting("default_service")
	return v
}

func (db *Database) DeleteAllSessionKeys() error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx, `DELETE FROM bot_settings WHERE s_key LIKE 'session_%'`)
	return err
}

// ─── Auth ─────────────────────────────────────────────────

func (db *Database) AddAuthUser(chatID, userID int64) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO auth_users (chat_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		chatID, userID,
	)
	if err == nil {
		db.authCache.Delete(toKey(chatID))
	}
	return err
}

func (db *Database) RemoveAuthUser(chatID, userID int64) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`DELETE FROM auth_users WHERE chat_id = $1 AND user_id = $2`,
		chatID, userID,
	)
	if err == nil {
		db.authCache.Delete(toKey(chatID))
	}
	return err
}

func (db *Database) GetAuthUsers(chatID int64) []int64 {
	ctx, cancel := db.ctx()
	defer cancel()

	rows, err := db.pool.Query(ctx, `SELECT user_id FROM auth_users WHERE chat_id = $1`, chatID)
	if err != nil {
		return []int64{}
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var uid int64
		if err := rows.Scan(&uid); err != nil {
			return []int64{}
		}
		users = append(users, uid)
	}

	return users
}

func (db *Database) IsAuthUser(chatID, userID int64) bool {
	users := db.GetAuthUsers(chatID)
	return contains(users, userID)
}

func (db *Database) IsAdmin(chatID, userID int64) bool {
	admins, err := cache.GetChatAdminIDs(chatID)
	if err != nil || admins == nil {
		return false
	}
	return contains(admins, userID)
}

// ─── Blacklist ────────────────────────────────────────────

func (db *Database) AddBlacklistedChat(chatID int64) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO blacklist (type, id) VALUES ('chat', $1) ON CONFLICT DO NOTHING`, chatID,
	)
	return err
}

func (db *Database) RemoveBlacklistedChat(chatID int64) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`DELETE FROM blacklist WHERE type = 'chat' AND id = $1`, chatID,
	)
	return err
}

func (db *Database) GetBlacklistedChats() []int64 {
	ctx, cancel := db.ctx()
	defer cancel()

	rows, err := db.pool.Query(ctx, `SELECT id FROM blacklist WHERE type = 'chat'`)
	if err != nil {
		return []int64{}
	}
	defer rows.Close()

	var chats []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return []int64{}
		}
		chats = append(chats, id)
	}

	return chats
}

func (db *Database) IsBlacklistedChat(chatID int64) bool {
	chats := db.GetBlacklistedChats()
	return contains(chats, chatID)
}

func (db *Database) AddBlacklistedUser(userID int64) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO blacklist (type, id) VALUES ('user', $1) ON CONFLICT DO NOTHING`, userID,
	)
	return err
}

func (db *Database) RemoveBlacklistedUser(userID int64) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`DELETE FROM blacklist WHERE type = 'user' AND id = $1`, userID,
	)
	return err
}

func (db *Database) GetBlacklistedUsers() []int64 {
	ctx, cancel := db.ctx()
	defer cancel()

	rows, err := db.pool.Query(ctx, `SELECT id FROM blacklist WHERE type = 'user'`)
	if err != nil {
		return []int64{}
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return []int64{}
		}
		users = append(users, id)
	}

	return users
}

func (db *Database) IsBlacklistedUser(userID int64) bool {
	users := db.GetBlacklistedUsers()
	return contains(users, userID)
}

// ─── Logger ───────────────────────────────────────────────

func (db *Database) GetLoggerStatus() bool {
	ctx, cancel := db.ctx()
	defer cancel()

	var status bool
	err := db.pool.QueryRow(ctx,
		`SELECT COALESCE((SELECT value->>'status' FROM settings WHERE key = 'logger')::boolean, false)`,
	).Scan(&status)
	if err != nil {
		return false
	}
	return status
}

func (db *Database) SetLoggerStatus(status bool) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO settings (key, value) VALUES ('logger', jsonb_build_object('status', $1)) ON CONFLICT (key) DO UPDATE SET value = jsonb_build_object('status', $1)`,
		status,
	)
	return err
}

// ─── Language ─────────────────────────────────────────────

func (db *Database) GetLanguage(chatID int64) (string, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	var lang string
	err := db.pool.QueryRow(ctx,
		`SELECT lang FROM chat_languages WHERE chat_id = $1`, chatID,
	).Scan(&lang)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "en", nil
		}
		return "", err
	}

	db.langCache.Set(toKey(chatID), lang)
	return lang, nil
}

func (db *Database) SetLanguage(chatID int64, langCode string) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO chat_languages (chat_id, lang) VALUES ($1, $2) ON CONFLICT (chat_id) DO UPDATE SET lang = $2`,
		chatID, langCode,
	)
	if err == nil {
		db.langCache.Set(toKey(chatID), langCode)
	}
	return err
}

// ─── Playlists ────────────────────────────────────────────

func generateUniquePlaylistID() string {
	b := make([]byte, 5)
	rand.Read(b)
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
		if errors.Is(err, pgx.ErrNoRows) {
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

func (db *Database) AddSongToPlaylist(id string, song Song) error {
	ctx, cancel := db.ctx()
	defer cancel()

	songJSON, _ := json.Marshal(song)
	_, err := db.pool.Exec(ctx,
		`UPDATE playlists SET songs = songs || $1::jsonb WHERE id = $2 AND NOT EXISTS (
			SELECT 1 FROM jsonb_array_elements(songs) elem WHERE elem->>'track_id' = $3
		)`,
		[]byte(fmt.Sprintf("[%s]", string(songJSON))), id, song.TrackID,
	)
	return err
}

func (db *Database) RemoveSongFromPlaylist(id string, trackID string) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`UPDATE playlists SET songs = (
			SELECT jsonb_agg(elem) FROM jsonb_array_elements(songs) elem WHERE elem->>'track_id' != $2
		) WHERE id = $1`,
		id, trackID,
	)
	return err
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

// ─── Assistant ───────────────────────────────────────────

func (db *Database) GetAssistant(chatID int64) (int, error) {
	key := toKey(chatID)
	if cached, ok := db.assistantCache.Get(key); ok {
		n, _ := strconv.Atoi(cached)
		return n, nil
	}

	ctx, cancel := db.ctx()
	defer cancel()

	var num int
	err := db.pool.QueryRow(ctx,
		`SELECT num FROM assistant_assignments WHERE chat_id = $1`, chatID,
	).Scan(&num)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, nil
		}
		return -1, err
	}

	db.assistantCache.Set(key, strconv.Itoa(num))
	return num, nil
}

func (db *Database) SetAssistant(chatID int64, num int) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO assistant_assignments (chat_id, num) VALUES ($1, $2) ON CONFLICT (chat_id) DO UPDATE SET num = $2`,
		chatID, num,
	)
	if err == nil {
		db.assistantCache.Set(toKey(chatID), strconv.Itoa(num))
	}
	return err
}

func (db *Database) RemoveAssistant(chatID int64) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`DELETE FROM assistant_assignments WHERE chat_id = $1`, chatID,
	)
	if err == nil {
		db.assistantCache.Delete(toKey(chatID))
	}
	return err
}

func (db *Database) AssignAssistant(chatID int64, proposedAssistant int) (int, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return -1, err
	}
	defer tx.Rollback(ctx)

	var existingNum int
	err = tx.QueryRow(ctx,
		`SELECT num FROM assistant_assignments WHERE chat_id = $1 FOR UPDATE`, chatID,
	).Scan(&existingNum)

	if err == nil && existingNum != -1 {
		tx.Commit(ctx)
		return existingNum, nil
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO assistant_assignments (chat_id, num) VALUES ($1, $2) ON CONFLICT (chat_id) DO UPDATE SET num = $2`,
		chatID, proposedAssistant,
	)
	if err != nil {
		return -1, err
	}

	if err := tx.Commit(ctx); err != nil {
		return -1, err
	}

	db.assistantCache.Set(toKey(chatID), strconv.Itoa(proposedAssistant))
	return proposedAssistant, nil
}

func (db *Database) ClearAllAssistants() (int64, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	result, err := db.pool.Exec(ctx, `DELETE FROM assistant_assignments`)
	if err != nil {
		return 0, err
	}

	db.assistantCache.Clear()
	return result.RowsAffected(), nil
}

// ─── Rankings / Stats ────────────────────────────────────

func (db *Database) IncrementPlayCount(userID, chatID int64) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO user_stats (user_id, chat_id, date, play_count)
		 VALUES ($1, $2, CURRENT_DATE, 1)
		 ON CONFLICT (user_id, chat_id, date)
		 DO UPDATE SET play_count = user_stats.play_count + 1`,
		userID, chatID,
	)
	return err
}

func (db *Database) GetUserTotalPlays(userID int64) (int, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	var total int
	err := db.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(play_count), 0) FROM user_stats WHERE user_id = $1`, userID,
	).Scan(&total)
	return total, err
}

func (db *Database) GetUserGroupPlays(userID, chatID int64) (int, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	var total int
	err := db.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(play_count), 0) FROM user_stats WHERE user_id = $1 AND chat_id = $2`,
		userID, chatID,
	).Scan(&total)
	return total, err
}

func (db *Database) GetUserGlobalRank(userID int64) (int, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	var rank int
	err := db.pool.QueryRow(ctx,
		`SELECT rank FROM (
			SELECT user_id, RANK() OVER (ORDER BY SUM(play_count) DESC) as rank
			FROM user_stats GROUP BY user_id
		) r WHERE user_id = $1`, userID,
	).Scan(&rank)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return rank, err
}

func (db *Database) GetGroupTop(chatID int64, limit int) ([]struct{ UserID int64; Plays int }, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	rows, err := db.pool.Query(ctx,
		`SELECT user_id, SUM(play_count) as plays
		 FROM user_stats WHERE chat_id = $1
		 GROUP BY user_id ORDER BY plays DESC LIMIT $2`,
		chatID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []struct{ UserID int64; Plays int }
	for rows.Next() {
		var item struct{ UserID int64; Plays int }
		if err := rows.Scan(&item.UserID, &item.Plays); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (db *Database) GetGlobalTop(limit int) ([]struct{ UserID int64; Plays int }, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	rows, err := db.pool.Query(ctx,
		`SELECT user_id, SUM(play_count) as plays
		 FROM user_stats GROUP BY user_id
		 ORDER BY plays DESC LIMIT $1`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []struct{ UserID int64; Plays int }
	for rows.Next() {
		var item struct{ UserID int64; Plays int }
		if err := rows.Scan(&item.UserID, &item.Plays); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// ─── Convert ─────────────────────────────────────────────

func ConvertSongsToTracks(songs []Song) []types.MusicTrack {
	tracks := make([]types.MusicTrack, 0, len(songs))
	for _, song := range songs {
		tracks = append(tracks, types.MusicTrack{
			Url: song.URL, Title: song.Name, Id: song.TrackID,
			Duration: song.Duration, Platform: song.Platform,
		})
	}
	return tracks
}

// ─── Package-level wrappers (delegate to Instance) ──────

func GetSetting(key string) (string, error) {
	if Instance == nil {
		return "", fmt.Errorf("database not initialized")
	}
	return Instance.GetSetting(key)
}

func SetSetting(key, value string) error {
	if Instance == nil {
		return fmt.Errorf("database not initialized")
	}
	return Instance.SetSetting(key, value)
}

func DeleteSetting(key string) error {
	if Instance == nil {
		return fmt.Errorf("database not initialized")
	}
	return Instance.DeleteSetting(key)
}

func GetAllSettings() (map[string]string, error) {
	if Instance == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return Instance.GetAllSettings()
}

func AddUser(userID int64) error {
	if Instance == nil {
		return fmt.Errorf("database not initialized")
	}
	return Instance.AddUser(userID)
}

func AddChat(chatID int64) error {
	if Instance == nil {
		return fmt.Errorf("database not initialized")
	}
	return Instance.AddChat(chatID)
}

func GetLoggerStatus() bool {
	if Instance == nil {
		return false
	}
	return Instance.GetLoggerStatus()
}

func SetLoggerStatus(status bool) error {
	if Instance == nil {
		return fmt.Errorf("database not initialized")
	}
	return Instance.SetLoggerStatus(status)
}

func IsAuthUser(chatID, userID int64) bool {
	if Instance == nil {
		return false
	}
	return Instance.IsAuthUser(chatID, userID)
}

func GetPlaylist(id string) (*Playlist, error) {
	if Instance == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return Instance.GetPlaylist(id)
}

func DeletePlaylist(id string, userID int64) error {
	if Instance == nil {
		return fmt.Errorf("database not initialized")
	}
	return Instance.DeletePlaylist(id, userID)
}

func GetAllChats() ([]int64, error) {
	if Instance == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return Instance.GetAllChats()
}

func GetAllUsers() ([]int64, error) {
	if Instance == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return Instance.GetAllUsers()
}

func GetUserPlaylists(userID int64) ([]Playlist, error) {
	if Instance == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return Instance.GetUserPlaylists(userID)
}

func CreatePlaylist(name string, userID int64) (string, error) {
	if Instance == nil {
		return "", fmt.Errorf("database not initialized")
	}
	return Instance.CreatePlaylist(name, userID)
}

func AddSongToPlaylist(id string, song Song) error {
	if Instance == nil {
		return fmt.Errorf("database not initialized")
	}
	return Instance.AddSongToPlaylist(id, song)
}

func RemoveSongFromPlaylist(id string, trackID string) error {
	if Instance == nil {
		return fmt.Errorf("database not initialized")
	}
	return Instance.RemoveSongFromPlaylist(id, trackID)
}

func GetAssistant(chatID int64) (int, error) {
	if Instance == nil {
		return -1, fmt.Errorf("database not initialized")
	}
	return Instance.GetAssistant(chatID)
}

func SetAssistant(chatID int64, num int) error {
	if Instance == nil {
		return fmt.Errorf("database not initialized")
	}
	return Instance.SetAssistant(chatID, num)
}

func RemoveAssistant(chatID int64) error {
	if Instance == nil {
		return fmt.Errorf("database not initialized")
	}
	return Instance.RemoveAssistant(chatID)
}

func AssignAssistant(chatID int64, proposedAssistant int) (int, error) {
	if Instance == nil {
		return -1, fmt.Errorf("database not initialized")
	}
	return Instance.AssignAssistant(chatID, proposedAssistant)
}

func ClearAllAssistants() (int64, error) {
	if Instance == nil {
		return 0, fmt.Errorf("database not initialized")
	}
	return Instance.ClearAllAssistants()
}

func GetPlayMode(chatID int64) bool {
	if Instance == nil {
		return false
	}
	return Instance.GetPlayMode(chatID)
}

func SetPlayMode(chatID int64, adminPlay bool) error {
	if Instance == nil {
		return fmt.Errorf("database not initialized")
	}
	return Instance.SetPlayMode(chatID, adminPlay)
}

func GetAdminMode(chatID int64) string {
	if Instance == nil {
		return "everyone"
	}
	return Instance.GetAdminMode(chatID)
}

func SetAdminMode(chatID int64, adminMode string) error {
	if Instance == nil {
		return fmt.Errorf("database not initialized")
	}
	return Instance.SetAdminMode(chatID, adminMode)
}

func GetCmdDelete(chatID int64) bool {
	if Instance == nil {
		return false
	}
	return Instance.GetCmdDelete(chatID)
}

func SetCmdDelete(chatID int64, cmdDelete bool) error {
	if Instance == nil {
		return fmt.Errorf("database not initialized")
	}
	return Instance.SetCmdDelete(chatID, cmdDelete)
}

func GetLanguage(chatID int64) (string, error) {
	if Instance == nil {
		return "en", fmt.Errorf("database not initialized")
	}
	return Instance.GetLanguage(chatID)
}

func IncrementPlayCount(userID, chatID int64) error {
	if Instance == nil {
		return fmt.Errorf("database not initialized")
	}
	return Instance.IncrementPlayCount(userID, chatID)
}

func GetUserTotalPlays(userID int64) (int, error) {
	if Instance == nil {
		return 0, fmt.Errorf("database not initialized")
	}
	return Instance.GetUserTotalPlays(userID)
}

func GetUserGroupPlays(userID, chatID int64) (int, error) {
	if Instance == nil {
		return 0, fmt.Errorf("database not initialized")
	}
	return Instance.GetUserGroupPlays(userID, chatID)
}

func GetUserGlobalRank(userID int64) (int, error) {
	if Instance == nil {
		return 0, fmt.Errorf("database not initialized")
	}
	return Instance.GetUserGlobalRank(userID)
}

func GetGroupTop(chatID int64, limit int) ([]struct{ UserID int64; Plays int }, error) {
	if Instance == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return Instance.GetGroupTop(chatID, limit)
}

func GetGlobalTop(limit int) ([]struct{ UserID int64; Plays int }, error) {
	if Instance == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return Instance.GetGlobalTop(limit)
}

func GetOwnerDevs() []int64 {
	if Instance == nil {
		return nil
	}
	return Instance.GetOwnerDevs()
}

func GetOwnerLoggerId() int64 {
	if Instance == nil {
		return 0
	}
	return Instance.GetOwnerLoggerId()
}

func GetOwnerSupportGroup() string {
	if Instance == nil {
		return ""
	}
	return Instance.GetOwnerSupportGroup()
}

func GetOwnerSupportChannel() string {
	if Instance == nil {
		return ""
	}
	return Instance.GetOwnerSupportChannel()
}

func GetOwnerSongDuration() int64 {
	if Instance == nil {
		return 0
	}
	return Instance.GetOwnerSongDuration()
}

func GetOwnerDefaultService() string {
	if Instance == nil {
		return ""
	}
	return Instance.GetOwnerDefaultService()
}

func GetSessionStrings() ([]string, error) {
	if Instance == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return Instance.GetSessionStrings()
}

func DeleteAllSessionKeys() error {
	if Instance == nil {
		return fmt.Errorf("database not initialized")
	}
	return Instance.DeleteAllSessionKeys()
}

func GetAuthUsers(chatID int64) []int64 {
	if Instance == nil {
		return nil
	}
	return Instance.GetAuthUsers(chatID)
}

func AddAuthUser(chatID, userID int64) error {
	if Instance == nil {
		return fmt.Errorf("database not initialized")
	}
	return Instance.AddAuthUser(chatID, userID)
}

func RemoveAuthUser(chatID, userID int64) error {
	if Instance == nil {
		return fmt.Errorf("database not initialized")
	}
	return Instance.RemoveAuthUser(chatID, userID)
}

func IsAdmin(chatID, userID int64) bool {
	if Instance == nil {
		return false
	}
	return Instance.IsAdmin(chatID, userID)
}
