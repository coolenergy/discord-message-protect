package sessions

type ISessionManager interface {
	Authenticate(userId, userName string)
	IsAuthenticated(userId string) bool
}
