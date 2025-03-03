package service

type Service struct {
	BackofficeUserService  BackofficeUserService
	AuthService            AuthGenerator
	AuthorizationService   AuthorizationService
	MatchMakingService     MatchMakingService
	MatchManagementService MatchManagementService
	Presence               PresenceService
	Notification           Notification
}
