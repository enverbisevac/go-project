package app

type UserAggregate struct {
	User
	Permissions []PermissionCheck `json:"permissions"`
	Roles       []string          `json:"roles"`
}

type RoleAggregate struct {
	Role
	Permissions []PermissionCheck `json:"permissions"`
}
