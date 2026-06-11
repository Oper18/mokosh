package form

// Register holds the fields submitted by a new user during self-registration.
// Only safe, user-supplied fields are exposed here; admin flag and auth
// provider are set server-side and never accepted from this form.
// Note: Role can be specified during registration but only to limited values (guest/photographer).
type Register struct {
	UserName    string `json:"UserName"`
	UserEmail   string `json:"UserEmail"`
	DisplayName string `json:"DisplayName"`
	Password    string `json:"Password"`
	Role        string `json:"Role"`
}
