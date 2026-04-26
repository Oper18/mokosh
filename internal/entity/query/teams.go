package query

import "github.com/photoprism/photoprism/internal/entity"

// AlbumUIDsByTeamMember returns all album UIDs accessible to userUID via their
// team memberships.
func AlbumUIDsByTeamMember(userUID string) ([]string, error) {
	if userUID == "" {
		return nil, nil
	}

	var rows []entity.TeamAlbum

	err := Db().
		Where("team_uid IN (SELECT team_uid FROM teams_users WHERE user_uid = ?)", userUID).
		Find(&rows).Error

	uids := make([]string, len(rows))
	for i, r := range rows {
		uids[i] = r.AlbumUID
	}

	return uids, err
}

// PhotoUIDsByTeamMember returns all photo UIDs accessible to userUID via their
// team memberships.
func PhotoUIDsByTeamMember(userUID string) ([]string, error) {
	if userUID == "" {
		return nil, nil
	}

	var rows []entity.TeamPhoto

	err := Db().
		Where("team_uid IN (SELECT team_uid FROM teams_users WHERE user_uid = ?)", userUID).
		Find(&rows).Error

	uids := make([]string, len(rows))
	for i, r := range rows {
		uids[i] = r.PhotoUID
	}

	return uids, err
}
