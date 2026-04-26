package form

// Link represents a link sharing form.
type Link struct {
	Password    string `json:"Password"` //nolint:gosec // G117: Expected user-supplied credential field.
	LinkName    string `json:"Name"`
	ShareSlug   string `json:"Slug"`
	LinkToken   string `json:"Token"`
	LinkExpires int    `json:"Expires"`
	MaxViews    uint   `json:"MaxViews"`
	Perm        uint   `json:"Perm"`
	CanComment  bool   `json:"CanComment"`
	CanEdit     bool   `json:"CanEdit"`
}
