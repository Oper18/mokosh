package entity

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Team identifier prefix.
const (
	TeamUID    = byte('t')
	TeamPrefix = "team"
)

// Teams is a convenience alias for slices of Team.
type Teams []Team

// Team represents a group of users that can be granted shared access to albums.
type Team struct {
	ID        uint      `gorm:"primary_key" json:"ID" yaml:"-"`
	TeamUID   string    `gorm:"type:VARBINARY(42);unique_index;" json:"UID" yaml:"UID"`
	TeamName  string    `gorm:"size:200;index;" json:"Name" yaml:"Name"`
	UserUID   string    `gorm:"type:VARBINARY(42);index;" json:"UserUID" yaml:"UserUID"`
	CreatedAt time.Time `json:"CreatedAt" yaml:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt" yaml:"UpdatedAt"`
}

// TableName returns the database table name.
func (Team) TableName() string {
	return "teams"
}

// BeforeCreate generates a random UID if not set before inserting a new row.
func (m *Team) BeforeCreate(scope *gorm.Scope) error {
	if rnd.InvalidUID(m.TeamUID, TeamUID) {
		m.TeamUID = rnd.GenerateUID(TeamUID)
		return scope.SetColumn("TeamUID", m.TeamUID)
	}
	return nil
}

// AfterDelete cascades deletion to related TeamUser, TeamAlbum, and AlbumUser records.
func (m *Team) AfterDelete(tx *gorm.DB) error {
	tx.Where("team_uid = ?", m.TeamUID).Delete(&TeamUser{})
	tx.Where("team_uid = ?", m.TeamUID).Delete(&TeamAlbum{})
	tx.Where("team_uid = ?", m.TeamUID).Delete(&TeamPhoto{})
	tx.Where("team_uid = ?", m.TeamUID).Delete(&AlbumUser{})
	return nil
}

// NewTeam creates a new team owned by the given user UID.
func NewTeam(name, userUID string) *Team {
	return &Team{
		TeamUID:  rnd.GenerateUID(TeamUID),
		TeamName: name,
		UserUID:  userUID,
	}
}

// Create inserts a new record into the database.
func (m *Team) Create() error {
	return Db().Create(m).Error
}

// Save updates the record or inserts it when no row exists yet.
func (m *Team) Save() error {
	return Db().Save(m).Error
}

// Delete permanently removes the team and all related records from the database.
func (m *Team) Delete() error {
	return Db().Delete(m).Error
}

// FirstOrCreateTeam returns the existing record or inserts it when missing.
func FirstOrCreateTeam(m *Team) *Team {
	found := Team{}

	if err := Db().Where("team_uid = ?", m.TeamUID).First(&found).Error; err == nil {
		return &found
	} else if err = m.Create(); err != nil {
		event.AuditErr([]string{"team %s", "failed to create", status.Error(err)}, m.TeamUID)
		return nil
	}

	return m
}

// FindTeam returns the team with the given UID or nil if not found.
func FindTeam(uid string) *Team {
	if rnd.InvalidUID(uid, TeamUID) {
		return nil
	}

	m := Team{}

	if err := Db().Where("team_uid = ?", uid).First(&m).Error; err != nil {
		return nil
	}

	return &m
}
