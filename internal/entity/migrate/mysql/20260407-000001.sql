CREATE TABLE IF NOT EXISTS `teams` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `team_uid` varbinary(42) NOT NULL DEFAULT '',
  `team_name` varchar(200) NOT NULL DEFAULT '',
  `user_uid` varbinary(42) NOT NULL DEFAULT '',
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uix_teams_team_uid` (`team_uid`),
  KEY `idx_teams_team_name` (`team_name`),
  KEY `idx_teams_user_uid` (`user_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `teams_users` (
  `uid` varbinary(42) NOT NULL DEFAULT '',
  `team_uid` varbinary(42) NOT NULL DEFAULT '',
  `user_uid` varbinary(42) NOT NULL DEFAULT '',
  PRIMARY KEY (`uid`, `team_uid`),
  KEY `idx_teams_users_team_uid` (`team_uid`),
  KEY `idx_teams_users_user_uid` (`user_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `teams_albums` (
  `uid` varbinary(42) NOT NULL DEFAULT '',
  `team_uid` varbinary(42) NOT NULL DEFAULT '',
  `album_uid` varbinary(42) NOT NULL DEFAULT '',
  PRIMARY KEY (`uid`, `team_uid`),
  KEY `idx_teams_albums_team_uid` (`team_uid`),
  KEY `idx_teams_albums_album_uid` (`album_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `teams_photos` (
  `uid` varbinary(42) NOT NULL DEFAULT '',
  `team_uid` varbinary(42) NOT NULL DEFAULT '',
  `photo_uid` varbinary(42) NOT NULL DEFAULT '',
  PRIMARY KEY (`uid`, `team_uid`),
  KEY `idx_teams_photos_team_uid` (`team_uid`),
  KEY `idx_teams_photos_photo_uid` (`photo_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
