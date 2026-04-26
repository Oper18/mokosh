package entity

import (
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TeamUser maps a user to a team.
type TeamUser struct {
	UID     string `gorm:"type:VARBINARY(42);primary_key;auto_increment:false" json:"UID" yaml:"UID"`
	TeamUID string `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;index" json:"TeamUID" yaml:"TeamUID"`
	UserUID string `gorm:"type:VARBINARY(42);index" json:"UserUID,omitempty" yaml:"UserUID,omitempty"`
}

// TableName returns the database table name.
func (TeamUser) TableName() string {
	return "teams_users"
}

// NewTeamUser creates a new team-user membership record.
func NewTeamUser(teamUID, userUID string) *TeamUser {
	return &TeamUser{
		UID:     rnd.GenerateUID(TeamUID),
		TeamUID: teamUID,
		UserUID: userUID,
	}
}

// Create inserts a new record into the database.
func (m *TeamUser) Create() error {
	return Db().Create(m).Error
}

// Save updates the record or inserts it when no row exists yet.
func (m *TeamUser) Save() error {
	return Db().Save(m).Error
}

// Delete removes the membership record from the database.
func (m *TeamUser) Delete() error {
	return Db().Delete(m).Error
}

// FirstOrCreateTeamUser returns the existing record or inserts it when missing.
func FirstOrCreateTeamUser(m *TeamUser) *TeamUser {
	found := TeamUser{}

	if err := Db().Where("uid = ?", m.UID).First(&found).Error; err == nil {
		return &found
	} else if err = m.Create(); err != nil {
		event.AuditErr([]string{"team %s", "failed to add member", status.Error(err)}, m.TeamUID)
		return nil
	}

	return m
}
