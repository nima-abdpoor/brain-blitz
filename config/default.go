package config

var defaultConfig = map[string]interface{}{
	"auth.refresh_subject":          RefreshTokenSubject,
	"auth.access_subject":           AccessTokenSubject,
	"matchMaking.waitingListPrefix": WaitingListPrefix,
}
