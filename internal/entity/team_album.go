package entity

import (
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TeamAlbum maps a team to an album, granting the team access to it.
type TeamAlbum struct {
	UID      string `gorm:"type:VARBINARY(42);primary_key;auto_increment:false" json:"UID" yaml:"UID"`
	TeamUID  string `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;index" json:"TeamUID" yaml:"TeamUID"`
	AlbumUID string `gorm:"type:VARBINARY(42);index" json:"AlbumUID,omitempty" yaml:"AlbumUID,omitempty"`
}

// TableName returns the database table name.
func (TeamAlbum) TableName() string {
	return "teams_albums"
}

// NewTeamAlbum creates a new team-album access record.
func NewTeamAlbum(teamUID, albumUID string) *TeamAlbum {
	return &TeamAlbum{
		UID:      rnd.GenerateUID(TeamUID),
		TeamUID:  teamUID,
		AlbumUID: albumUID,
	}
}

// Create inserts a new record into the database.
func (m *TeamAlbum) Create() error {
	return Db().Create(m).Error
}

// Save updates the record or inserts it when no row exists yet.
func (m *TeamAlbum) Save() error {
	return Db().Save(m).Error
}

// Delete removes the access record from the database.
func (m *TeamAlbum) Delete() error {
	return Db().Delete(m).Error
}

// FirstOrCreateTeamAlbum returns the existing record or inserts it when missing.
func FirstOrCreateTeamAlbum(m *TeamAlbum) *TeamAlbum {
	found := TeamAlbum{}

	if err := Db().Where("uid = ?", m.UID).First(&found).Error; err == nil {
		return &found
	} else if err = m.Create(); err != nil {
		event.AuditErr([]string{"team %s", "failed to add album", status.Error(err)}, m.TeamUID)
		return nil
	}

	return m
}
