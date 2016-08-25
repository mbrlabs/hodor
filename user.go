package hodor

// User #
type User interface {
	GetID() string
	SetID(string)
	GetLogin() string
	SetLogin(string)
	GetEmail() string
	SetEmail(string)
	GetPassword() string
	SetPassword(string)

	GetUserRoles() []*UserRole
	GetIndividualUserRights() []*UserRight
}

// UserStore #
type UserStore interface {
	GetUserByLogin(string) User
	GetUserByID(string) User
	Authenticate(User, string) bool
}

// UserRole #
type UserRole struct {
	Name   string
	Rights []*UserRight
}

// UserRight #
type UserRight struct {
	Name string
}
