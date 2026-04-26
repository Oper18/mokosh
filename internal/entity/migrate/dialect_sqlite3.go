package migrate

// Generated code, do not edit.

var DialectSQLite3 = Migrations{
	{
		ID:         "20211121-094727",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"DROP INDEX IF EXISTS idx_places_place_label;"},
	},
	{
		ID:         "20211124-120008",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"DROP INDEX IF EXISTS uix_places_place_label;", "DROP INDEX IF EXISTS uix_places_label;"},
	},
	{
		ID:         "20220329-040000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"DROP INDEX IF EXISTS idx_albums_album_filter;"},
	},
	{
		ID:         "20220329-050000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"CREATE INDEX IF NOT EXISTS idx_albums_album_filter ON albums (album_filter);"},
	},
	{
		ID:         "20220329-061000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"CREATE INDEX IF NOT EXISTS idx_files_photo_id ON files (photo_id, file_primary);"},
	},
	{
		ID:         "20220329-071000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"UPDATE files SET photo_taken_at = (SELECT taken_at_local FROM photos WHERE photos.id = photo_id) WHERE photo_id IS NOT NULL;"},
	},
	{
		ID:         "20220329-081000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"CREATE UNIQUE INDEX IF NOT EXISTS idx_files_search_media ON files (media_id);"},
	},
	{
		ID:         "20220329-083000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"UPDATE files SET media_id = CASE WHEN photo_id IS NOT NULL AND file_missing = 0 AND deleted_at IS NULL THEN ((10000000000 - photo_id) || '-' || (1 + file_sidecar - file_primary) || '-' || file_uid) END WHERE 1;"},
	},
	{
		ID:         "20220329-091000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"CREATE UNIQUE INDEX IF NOT EXISTS idx_files_search_timeline ON files (time_index);"},
	},
	{
		ID:         "20220329-093000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"UPDATE files SET time_index = CASE WHEN media_id IS NOT NULL AND photo_taken_at IS NOT NULL THEN ((100000000000000 - strftime('%Y%m%d%H%M%S', photo_taken_at)) || '-' || media_id) ELSE NULL END WHERE photo_id IS NOT NULL;"},
	},
	{
		ID:         "20220421-200000",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"CREATE INDEX IF NOT EXISTS idx_files_missing_root ON files (file_missing, file_root);"},
	},
	{
		ID:         "20221015-100000",
		Dialect:    "sqlite3",
		Stage:      "pre",
		Statements: []string{"ALTER TABLE accounts RENAME TO services;"},
	},
	{
		ID:         "20221015-100100",
		Dialect:    "sqlite3",
		Stage:      "pre",
		Statements: []string{"ALTER TABLE files_sync RENAME COLUMN account_id TO service_id;", "ALTER TABLE files_share RENAME COLUMN account_id TO service_id;"},
	},
	{
		ID:         "20230309-000001",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"UPDATE auth_users SET auth_provider = 'local' WHERE id = 1;", "UPDATE auth_users SET auth_provider = 'none' WHERE id = -1;", "UPDATE auth_users SET auth_provider = 'token' WHERE id = -2;", "UPDATE auth_users SET auth_provider = 'default' WHERE auth_provider = '' OR auth_provider = 'password' OR auth_provider IS NULL;"},
	},
	{
		ID:         "20230313-000001",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"UPDATE auth_users SET user_role = 'contributor' WHERE user_role = 'uploader';", "UPDATE auth_sessions SET auth_provider = 'link' WHERE auth_provider = 'token';"},
	},
	{
		ID:         "20240112-000001",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"DELETE FROM auth_sessions;"},
	},
	{
		ID:         "20240709-000001",
		Dialect:    "sqlite3",
		Stage:      "pre",
		Statements: []string{"ALTER TABLE auth_sessions RENAME COLUMN auth_domain TO auth_issuer;"},
	},
	{
		ID:         "20241010-000001",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"UPDATE countries SET country_name = 'United States' WHERE country_name = 'USA' AND country_slug = 'usa';", "UPDATE albums SET album_location = 'United States' WHERE album_location = 'USA' AND album_type = 'state';"},
	},
	{
		ID:         "20241202-000001",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"UPDATE auth_users_details SET birth_year = -1 WHERE birth_year >= 0 AND birth_year < 1000 OR birth_year < -1 OR birth_year IS NULL;", "UPDATE auth_users_details SET birth_month = -1 WHERE birth_month = 0 OR birth_month < -1 OR birth_month > 12 OR birth_month IS NULL;", "UPDATE auth_users_details SET birth_day = -1 WHERE birth_day = 0 OR birth_day < -1 OR birth_day > 31 OR birth_day IS NULL;", "UPDATE auth_users_details SET user_country = 'zz' WHERE user_country = '' OR user_country IS NULL;"},
	},
	{
		ID:         "20250117-000001",
		Dialect:    "sqlite3",
		Stage:      "pre",
		Statements: []string{"ALTER TABLE photos RENAME COLUMN photo_description TO photo_caption;", "ALTER TABLE photos RENAME COLUMN description_src TO caption_src;"},
	},
	{
		ID:         "20250315-000001",
		Dialect:    "sqlite3",
		Stage:      "pre",
		Statements: []string{"ALTER TABLE auth_users_settings RENAME COLUMN default_page TO ui_start_page;"},
	},
	{
		ID:         "20250416-000001",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"UPDATE photos SET time_zone = 'Local' WHERE time_zone = '' OR time_zone IS NULL;"},
	},
	{
		ID:         "20251005-000001",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"UPDATE labels SET label_nsfw = 0 WHERE label_nsfw IS NULL;", "UPDATE photos_labels SET nsfw = 0 WHERE nsfw IS NULL;", "UPDATE photos_labels SET topicality = 0 WHERE topicality IS NULL;"},
	},
	{
		ID:         "20251007-000001",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"UPDATE photos SET indexed_at = checked_at WHERE indexed_at IS NULL;"},
	},
	{
		ID:         "20260407-000001",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"CREATE TABLE IF NOT EXISTS `teams` (\n  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,\n  `team_uid` varchar(42) NOT NULL DEFAULT '',\n  `team_name` varchar(200) NOT NULL DEFAULT '',\n  `user_uid` varchar(42) NOT NULL DEFAULT '',\n  `created_at` datetime,\n  `updated_at` datetime\n);", "CREATE UNIQUE INDEX IF NOT EXISTS `uix_teams_team_uid` ON `teams` (`team_uid`);", "CREATE INDEX IF NOT EXISTS `idx_teams_team_name` ON `teams` (`team_name`);", "CREATE INDEX IF NOT EXISTS `idx_teams_user_uid` ON `teams` (`user_uid`);", "CREATE TABLE IF NOT EXISTS `teams_users` (\n  `uid` varchar(42) NOT NULL DEFAULT '',\n  `team_uid` varchar(42) NOT NULL DEFAULT '',\n  `user_uid` varchar(42) NOT NULL DEFAULT '',\n  PRIMARY KEY (`uid`, `team_uid`)\n);", "CREATE INDEX IF NOT EXISTS `idx_teams_users_team_uid` ON `teams_users` (`team_uid`);", "CREATE INDEX IF NOT EXISTS `idx_teams_users_user_uid` ON `teams_users` (`user_uid`);", "CREATE TABLE IF NOT EXISTS `teams_albums` (\n  `uid` varchar(42) NOT NULL DEFAULT '',\n  `team_uid` varchar(42) NOT NULL DEFAULT '',\n  `album_uid` varchar(42) NOT NULL DEFAULT '',\n  PRIMARY KEY (`uid`, `team_uid`)\n);", "CREATE INDEX IF NOT EXISTS `idx_teams_albums_team_uid` ON `teams_albums` (`team_uid`);", "CREATE INDEX IF NOT EXISTS `idx_teams_albums_album_uid` ON `teams_albums` (`album_uid`);", "CREATE TABLE IF NOT EXISTS `teams_photos` (\n  `uid` varchar(42) NOT NULL DEFAULT '',\n  `team_uid` varchar(42) NOT NULL DEFAULT '',\n  `photo_uid` varchar(42) NOT NULL DEFAULT '',\n  PRIMARY KEY (`uid`, `team_uid`)\n);", "CREATE INDEX IF NOT EXISTS `idx_teams_photos_team_uid` ON `teams_photos` (`team_uid`);", "CREATE INDEX IF NOT EXISTS `idx_teams_photos_photo_uid` ON `teams_photos` (`photo_uid`);"},
	},
	{
		ID:      "20260419-000002",
		Dialect: "sqlite3",
		Stage:   "main",
		Statements: []string{
			"DROP TABLE IF EXISTS reactions;",
			"CREATE TABLE IF NOT EXISTS `reactions` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `photo_uid` varchar(42) DEFAULT '', `user_uid` varchar(42) DEFAULT '', `emoji` varchar(64) DEFAULT NULL, `comment` text DEFAULT NULL, `created_at` datetime);",
			"CREATE INDEX IF NOT EXISTS `idx_reactions_photo_uid` ON `reactions` (`photo_uid`);",
			"CREATE INDEX IF NOT EXISTS `idx_reactions_user_uid` ON `reactions` (`user_uid`);",
		},
	},
	{
		ID:         "20260420-000001",
		Dialect:    "sqlite3",
		Stage:      "main",
		Statements: []string{"ALTER TABLE links ADD COLUMN link_name varchar(160) NOT NULL DEFAULT '';"},
	},
	{
		ID:      "20260420-000002",
		Dialect: "sqlite3",
		Stage:   "main",
		Statements: []string{
			"DROP INDEX IF EXISTS idx_auth_users_user_name;",
			"CREATE UNIQUE INDEX IF NOT EXISTS uix_auth_users_user_name ON auth_users (user_name) WHERE user_name != '' AND user_name IS NOT NULL;",
		},
	},
}
