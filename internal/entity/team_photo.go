package entity

import (
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TeamPhoto maps a team to a photo, granting the team access to it.
type TeamPhoto struct {
	UID      string `gorm:"type:VARBINARY(42);primary_key;auto_increment:false" json:"UID" yaml:"UID"`
	TeamUID  string `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;index" json:"TeamUID" yaml:"TeamUID"`
	PhotoUID string `gorm:"type:VARBINARY(42);index" json:"PhotoUID,omitempty" yaml:"PhotoUID,omitempty"`
}

// TableName returns the database table name.
func (TeamPhoto) TableName() string {
	return "teams_photos"
}

// NewTeamPhoto creates a new team-photo access record.
func NewTeamPhoto(teamUID, photoUID string) *TeamPhoto {
	return &TeamPhoto{
		UID:      rnd.GenerateUID(TeamUID),
		TeamUID:  teamUID,
		PhotoUID: photoUID,
	}
}

// Create inserts a new record into the database.
func (m *TeamPhoto) Create() error {
	return Db().Create(m).Error
}

// Save updates the record or inserts it when no row exists yet.
func (m *TeamPhoto) Save() error {
	return Db().Save(m).Error
}

// Delete removes the access record from the database.
func (m *TeamPhoto) Delete() error {
	return Db().Delete(m).Error
}

// FirstOrCreateTeamPhoto returns the existing record or inserts it when missing.
func FirstOrCreateTeamPhoto(m *TeamPhoto) *TeamPhoto {
	found := TeamPhoto{}

	if err := Db().Where("uid = ?", m.UID).First(&found).Error; err == nil {
		return &found
	} else if err = m.Create(); err != nil {
		event.AuditErr([]string{"team %s", "failed to add photo", status.Error(err)}, m.TeamUID)
		return nil
	}

	return m
}
