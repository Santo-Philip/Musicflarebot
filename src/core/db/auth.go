package db

import (
	"musicflarebot/src/core/cache"
)

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
	key := toKey(chatID)
	if cached, ok := db.authCache.Get(key); ok {
		return cached
	}

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

	db.authCache.Set(key, users)
	return users
}

func (db *Database) IsAuthUser(chatID, userID int64) bool {
	admins, err := cache.GetChatAdminIDs(chatID)
	if err != nil || admins == nil {
		admins = []int64{}
	}
	if contains(admins, userID) {
		return true
	}
	users := db.GetAuthUsers(chatID)
	return contains(users, userID)
}

func (db *Database) IsAdmin(chatID, userID int64) bool {
	admins, err := cache.GetChatAdminIDs(chatID)
	if err != nil || admins == nil {
		admins = []int64{}
	}
	return contains(admins, userID)
}
