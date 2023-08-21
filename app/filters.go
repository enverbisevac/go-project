package app

type UserFilter struct {
	ID    string
	Email *string
}

func (f UserFilter) String() string {
	s := ""
	if f.ID != "" {
		s = "id = " + f.ID
	}

	if f.Email != nil && *f.Email != "" {
		s = "email = " + *f.Email
	}

	return s
}

type IDOrNameFilter struct {
	ID   string
	Name string
}

func (f IDOrNameFilter) String() string {
	if f.ID != "" {
		return "id = " + f.ID
	}
	if f.Name != "" {
		return "name = " + f.Name
	}
	return ""
}

type PermissionFilter struct {
	UserID string
	RoleID string
}
