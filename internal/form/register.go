package form

// Register holds the fields submitted by a new user during self-registration.
// Only safe, user-supplied fields are exposed here; role, admin flag and auth
// provider are set server-side and never accepted from this form.
type Register struct {
	UserName    string `json:"UserName"`
	UserEmail   string `json:"UserEmail"`
	DisplayName string `json:"DisplayName"`
	Password    string `json:"Password"`
}
