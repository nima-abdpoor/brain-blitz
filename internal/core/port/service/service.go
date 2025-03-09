package service

type Service struct {
	BackofficeUserService  BackofficeUserService
	AuthService            AuthGenerator
	AuthorizationService   AuthorizationService
	MatchManagementService MatchManagementService
	Presence               PresenceService
}
