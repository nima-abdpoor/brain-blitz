package entity

type Permission struct {
	ID    uint
	Title string
}

type PermissionTitle string

const (
	UserListPermission   = PermissionTitle("USER_LIST")
	UserDeletePermission = PermissionTitle("USER_DELETE")
	UserCreatePermission = PermissionTitle("USER_CREATE")
)
