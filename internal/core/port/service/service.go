package service

type Service struct {
	UserService           UserService
	BackofficeUserService BackofficeUserService
	AuthService           AuthGenerator
}
