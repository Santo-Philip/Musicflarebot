package db

import (
	"strconv"
	"strings"
)

const (
	SettingLoggerId       = "logger_id"
	SettingSupportGroup   = "support_group"
	SettingSupportChannel = "support_channel"
	SettingSongDuration   = "song_duration_limit"
	SettingDefaultService = "default_service"
	SettingDevs           = "devs"
)

func (db *Database) GetSetting(key string) (string, error) {
	ctx, cancel := db.ctx()
	defer cancel()

	var value string
	err := db.pool.QueryRow(ctx,
		`SELECT s_value FROM bot_settings WHERE s_key = $1`, key,
	).Scan(&value)

	if err != nil {
		if isNoRows(err) {
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

func (db *Database) DeleteAllSessionKeys() error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx, `DELETE FROM bot_settings WHERE s_key LIKE 'session_%'`)
	return err
}

func (db *Database) GetSessionStrings() ([]string, error) {
	all, err := db.GetAllSettings()
	if err != nil {
		return nil, err
	}

	var sessions []string
	for key, val := range all {
		if len(key) >= 8 && key[:8] == "session_" && val != "" {
			sessions = append(sessions, val)
		}
	}
	return sessions, nil
}

func (db *Database) GetOwnerLoggerId() int64 {
	val, err := db.GetSetting(SettingLoggerId)
	if err != nil || val == "" {
		return 0
	}
	id, _ := strconv.ParseInt(val, 10, 64)
	return id
}

func (db *Database) SetOwnerLoggerId(id int64) error {
	return db.SetSetting(SettingLoggerId, strconv.FormatInt(id, 10))
}

func (db *Database) GetOwnerSupportGroup() string {
	val, err := db.GetSetting(SettingSupportGroup)
	if err != nil {
		return ""
	}
	return val
}

func (db *Database) SetOwnerSupportGroup(link string) error {
	return db.SetSetting(SettingSupportGroup, link)
}

func (db *Database) GetOwnerSupportChannel() string {
	val, err := db.GetSetting(SettingSupportChannel)
	if err != nil {
		return ""
	}
	return val
}

func (db *Database) SetOwnerSupportChannel(link string) error {
	return db.SetSetting(SettingSupportChannel, link)
}

func (db *Database) GetOwnerSongDuration() int64 {
	val, err := db.GetSetting(SettingSongDuration)
	if err != nil || val == "" {
		return 0
	}
	d, _ := strconv.ParseInt(val, 10, 64)
	return d
}

func (db *Database) SetOwnerSongDuration(seconds int64) error {
	return db.SetSetting(SettingSongDuration, strconv.FormatInt(seconds, 10))
}

func (db *Database) GetOwnerDefaultService() string {
	val, err := db.GetSetting(SettingDefaultService)
	if err != nil {
		return ""
	}
	return val
}

func (db *Database) SetOwnerDefaultService(service string) error {
	return db.SetSetting(SettingDefaultService, strings.ToLower(service))
}

func (db *Database) GetOwnerDevs() []int64 {
	val, err := db.GetSetting(SettingDevs)
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
	return db.SetSetting(SettingDevs, strings.Join(parts, " "))
}
