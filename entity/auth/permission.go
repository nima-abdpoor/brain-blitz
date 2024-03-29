package entity

type Permission struct {
	ID    uint
	Title string
}

type PermissionTitle string

const (
	UserListPermission   = PermissionTitle("user-list")
	UserDeletePermission = PermissionTitle("user-delete")
)
