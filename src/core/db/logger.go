package db

func (db *Database) GetLoggerStatus() bool {
	if cached, ok := db.loggerCache.Get("logger"); ok {
		return cached
	}

	ctx, cancel := db.ctx()
	defer cancel()

	var status bool
	err := db.pool.QueryRow(ctx,
		`SELECT COALESCE((SELECT value->>'status' FROM settings WHERE key = 'logger')::boolean, false)`,
	).Scan(&status)
	if err != nil {
		return false
	}

	db.loggerCache.Set("logger", status)
	return status
}

func (db *Database) SetLoggerStatus(status bool) error {
	ctx, cancel := db.ctx()
	defer cancel()

	_, err := db.pool.Exec(ctx,
		`INSERT INTO settings (key, value) VALUES ('logger', jsonb_build_object('status', $1)) ON CONFLICT (key) DO UPDATE SET value = jsonb_build_object('status', $1)`,
		status,
	)
	if err == nil {
		db.loggerCache.Set("logger", status)
	}
	return err
}
