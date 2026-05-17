package db

func (db *Database) AddBlacklistedChat(chatID int64) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO blacklist (type, id) VALUES ('chat', $1) ON CONFLICT DO NOTHING`, chatID,
	)
	if err == nil {
		db.blChatsCache.Delete("bl_chats")
	}
	return err
}

func (db *Database) RemoveBlacklistedChat(chatID int64) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`DELETE FROM blacklist WHERE type = 'chat' AND id = $1`, chatID,
	)
	if err == nil {
		db.blChatsCache.Delete("bl_chats")
	}
	return err
}

func (db *Database) GetBlacklistedChats() []int64 {
	if cached, ok := db.blChatsCache.Get("bl_chats"); ok {
		return cached
	}

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

	db.blChatsCache.Set("bl_chats", chats)
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
	if err == nil {
		db.blUsersCache.Delete("bl_users")
	}
	return err
}

func (db *Database) RemoveBlacklistedUser(userID int64) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`DELETE FROM blacklist WHERE type = 'user' AND id = $1`, userID,
	)
	if err == nil {
		db.blUsersCache.Delete("bl_users")
	}
	return err
}

func (db *Database) GetBlacklistedUsers() []int64 {
	if cached, ok := db.blUsersCache.Get("bl_users"); ok {
		return cached
	}

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

	db.blUsersCache.Set("bl_users", users)
	return users
}

func (db *Database) IsBlacklistedUser(userID int64) bool {
	users := db.GetBlacklistedUsers()
	return contains(users, userID)
}
