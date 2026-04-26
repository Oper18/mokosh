package query

import (
	"time"

	"github.com/photoprism/photoprism/internal/entity"
)

// ReactionCount is an aggregated count of a single emoji reaction on a photo.
type ReactionCount struct {
	Emoji string `json:"emoji"`
	Count int    `json:"count"`
}

// ReactionDetail is a single reaction row with user display name.
type ReactionDetail struct {
	ID              uint       `json:"id"`
	Emoji           string     `json:"emoji,omitempty"`
	Comment         string     `json:"comment,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	UserDisplayName string     `json:"user"`
}

// PhotoReactionCounts returns aggregated emoji counts for a photo, ordered by popularity.
func PhotoReactionCounts(photoUID string) ([]ReactionCount, error) {
	var results []ReactionCount
	err := UnscopedDb().
		Model(&entity.Reaction{}).
		Select("emoji, count(*) as count").
		Where("photo_uid = ? AND emoji != ''", photoUID).
		Group("emoji").
		Order("count DESC").
		Scan(&results).Error
	return results, err
}

// PhotoReactionDetails returns a paginated list of reactions for a photo, newest first.
func PhotoReactionDetails(photoUID string, offset, limit int) ([]ReactionDetail, error) {
	var results []ReactionDetail
	err := UnscopedDb().
		Table("reactions").
		Select("reactions.id, reactions.emoji, reactions.comment, reactions.created_at, COALESCE(NULLIF(auth_users.display_name, ''), auth_users.user_name) as user_display_name").
		Joins("LEFT JOIN auth_users ON auth_users.user_uid = reactions.user_uid").
		Where("reactions.photo_uid = ?", photoUID).
		Order("reactions.created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&results).Error
	return results, err
}

// PhotoReactionTotal returns the total number of reactions for a photo.
func PhotoReactionTotal(photoUID string) (int, error) {
	var count int
	err := UnscopedDb().
		Model(&entity.Reaction{}).
		Where("photo_uid = ?", photoUID).
		Count(&count).Error
	return count, err
}
