package form

// React contains the emoji and optional comment for a photo reaction.
type React struct {
	Emoji   string `json:"emoji"`
	Comment string `json:"comment"`
}
