package form

// Team represents a team create/update form.
type Team struct {
	TeamName string `json:"Name"`
}

// TeamUser represents the form for adding a user to a team.
type TeamUser struct {
	UserUID string `json:"UserUID"`
}

// TeamAlbum represents the form for attaching a team to an album.
type TeamAlbum struct {
	AlbumUID string `json:"AlbumUID"`
}

// TeamPhoto represents the form for attaching a team to a photo.
type TeamPhoto struct {
	PhotoUID string `json:"PhotoUID"`
}
