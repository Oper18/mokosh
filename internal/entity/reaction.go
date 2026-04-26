package entity

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// Reaction represents a user's emoji reaction and/or comment on a photo.
// Multiple reactions per user are allowed; each row is independent.
type Reaction struct {
	ID        uint      `gorm:"primary_key" json:"ID"`
	PhotoUID  string    `gorm:"type:VARBINARY(42);index" json:"PhotoUID,omitempty"`
	UserUID   string    `gorm:"type:VARBINARY(42);index" json:"UserUID,omitempty"`
	Emoji     string    `gorm:"type:VARBINARY(64)" json:"Emoji,omitempty"`
	Comment   *string   `gorm:"type:TEXT" json:"Comment,omitempty"`
	CreatedAt time.Time `json:"CreatedAt"`
}

// TableName returns the entity table name.
func (Reaction) TableName() string {
	return "reactions"
}

// NewReaction creates a new Reaction for the given photo and user.
func NewReaction(photoUID, userUID string) *Reaction {
	return &Reaction{
		PhotoUID: photoUID,
		UserUID:  userUID,
	}
}

// WithEmoji sets the emoji on the reaction.
func (m *Reaction) WithEmoji(emoji string) *Reaction {
	m.Emoji = emoji
	return m
}

// WithComment sets the comment on the reaction.
func (m *Reaction) WithComment(comment string) *Reaction {
	if comment != "" {
		m.Comment = &comment
	}
	return m
}

// Empty checks whether both emoji and comment are unset.
func (m *Reaction) Empty() bool {
	return m.Emoji == "" && (m.Comment == nil || *m.Comment == "")
}

// InvalidUID checks if the photo or user uid is missing.
func (m *Reaction) InvalidUID() bool {
	return m.PhotoUID == "" || m.UserUID == ""
}

// Invalid checks if the reaction cannot be saved.
func (m *Reaction) Invalid() bool {
	return m.InvalidUID() || m.Empty()
}

// FindReactionByID returns the reaction with the given ID or nil.
func FindReactionByID(id uint) *Reaction {
	if id == 0 {
		return nil
	}
	m := &Reaction{}
	if Db().First(m, "id = ?", id).Error != nil {
		return nil
	}
	return m
}

// FindUserReactions returns all reactions left by a user for a photo.
func FindUserReactions(photoUID, userUID string) []Reaction {
	if rnd.InvalidUID(photoUID, 0) || rnd.InvalidUID(userUID, 0) {
		return nil
	}
	var results []Reaction
	if Db().Where("photo_uid = ? AND user_uid = ?", photoUID, userUID).
		Order("created_at DESC").
		Find(&results).Error != nil {
		return nil
	}
	return results
}

// Create inserts a new Reaction row.
func (m *Reaction) Create() error {
	if m.Invalid() {
		return fmt.Errorf("reaction is invalid")
	}
	m.CreatedAt = *TimeStamp()
	return Db().Create(m).Error
}

// Delete removes this Reaction from the database.
func (m *Reaction) Delete() error {
	if m.ID == 0 {
		return fmt.Errorf("reaction has no id")
	}
	return Db().Delete(m, "id = ?", m.ID).Error
}
