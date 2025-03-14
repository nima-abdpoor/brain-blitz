package service

type Service struct {
	BackofficeUserService BackofficeUserService
	AuthorizationService  AuthorizationService
	Presence              PresenceService
}
