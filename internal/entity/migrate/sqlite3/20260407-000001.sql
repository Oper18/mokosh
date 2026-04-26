CREATE TABLE IF NOT EXISTS `teams` (
  `id` integer NOT NULL PRIMARY KEY AUTOINCREMENT,
  `team_uid` varchar(42) NOT NULL DEFAULT '',
  `team_name` varchar(200) NOT NULL DEFAULT '',
  `user_uid` varchar(42) NOT NULL DEFAULT '',
  `created_at` datetime,
  `updated_at` datetime
);
CREATE UNIQUE INDEX IF NOT EXISTS `uix_teams_team_uid` ON `teams` (`team_uid`);
CREATE INDEX IF NOT EXISTS `idx_teams_team_name` ON `teams` (`team_name`);
CREATE INDEX IF NOT EXISTS `idx_teams_user_uid` ON `teams` (`user_uid`);

CREATE TABLE IF NOT EXISTS `teams_users` (
  `uid` varchar(42) NOT NULL DEFAULT '',
  `team_uid` varchar(42) NOT NULL DEFAULT '',
  `user_uid` varchar(42) NOT NULL DEFAULT '',
  PRIMARY KEY (`uid`, `team_uid`)
);
CREATE INDEX IF NOT EXISTS `idx_teams_users_team_uid` ON `teams_users` (`team_uid`);
CREATE INDEX IF NOT EXISTS `idx_teams_users_user_uid` ON `teams_users` (`user_uid`);

CREATE TABLE IF NOT EXISTS `teams_albums` (
  `uid` varchar(42) NOT NULL DEFAULT '',
  `team_uid` varchar(42) NOT NULL DEFAULT '',
  `album_uid` varchar(42) NOT NULL DEFAULT '',
  PRIMARY KEY (`uid`, `team_uid`)
);
CREATE INDEX IF NOT EXISTS `idx_teams_albums_team_uid` ON `teams_albums` (`team_uid`);
CREATE INDEX IF NOT EXISTS `idx_teams_albums_album_uid` ON `teams_albums` (`album_uid`);

CREATE TABLE IF NOT EXISTS `teams_photos` (
  `uid` varchar(42) NOT NULL DEFAULT '',
  `team_uid` varchar(42) NOT NULL DEFAULT '',
  `photo_uid` varchar(42) NOT NULL DEFAULT '',
  PRIMARY KEY (`uid`, `team_uid`)
);
CREATE INDEX IF NOT EXISTS `idx_teams_photos_team_uid` ON `teams_photos` (`team_uid`);
CREATE INDEX IF NOT EXISTS `idx_teams_photos_photo_uid` ON `teams_photos` (`photo_uid`);
