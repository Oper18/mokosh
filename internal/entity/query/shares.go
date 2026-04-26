package query

import (
	"github.com/photoprism/photoprism/internal/entity"
)

// UserSharesByOwner returns all UserShare records where the shared resource
// (album or photo) was created by the given user UID.
func UserSharesByOwner(ownerUID string) (shares entity.UserShares, err error) {
	if ownerUID == "" {
		return shares, nil
	}

	err = Db().
		Where(`share_uid IN (
			SELECT album_uid FROM albums WHERE created_by = ?
			UNION
			SELECT photo_uid FROM photos WHERE created_by = ?
		)`, ownerUID, ownerUID).
		Find(&shares).Error

	return shares, err
}

// UserShareOwned returns a single UserShare and verifies that the resource it
// references is owned by ownerUID. Returns nil when not found or not owned.
func UserShareOwned(shareUID, userUID, ownerUID string) *entity.UserShare {
	if shareUID == "" || userUID == "" || ownerUID == "" {
		return nil
	}

	// Confirm ownership of the resource.
	var count int64

	Db().Raw(`
		SELECT COUNT(*) FROM (
			SELECT album_uid AS uid FROM albums WHERE album_uid = ? AND created_by = ?
			UNION
			SELECT photo_uid AS uid FROM photos WHERE photo_uid = ? AND created_by = ?
		) owned`, shareUID, ownerUID, shareUID, ownerUID).Count(&count)

	if count == 0 {
		return nil
	}

	return entity.FindUserShare(entity.UserShare{UserUID: userUID, ShareUID: shareUID})
}
