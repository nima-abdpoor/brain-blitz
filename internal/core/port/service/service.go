package service

type Service struct {
	BackofficeUserService BackofficeUserService
	AuthService           AuthGenerator
	AuthorizationService  AuthorizationService
	Presence              PresenceService
}
