package sessions

import (
	"fmt"
	"github.com/melardev/discord-message-protect/core"
	"github.com/melardev/discord-message-protect/logging"
	"path/filepath"
	"sync"
	"time"
)

type UserInfo struct {
	Id              string
	Username        string
	AuthenticatedAt time.Time
	RefreshedAt     time.Time
}

type InMemoryAuthenticator struct {
	Lock              sync.RWMutex
	Config            *core.SessionsConfig
	Users             map[string]UserInfo
	DisconnectedUsers int
	Log               logging.ILogger
}

func NewInMemoryAuthenticator(appConfig *core.Config) *InMemoryAuthenticator {
	authenticator := &InMemoryAuthenticator{
		Log: logging.NewCompositeLogger(
			&logging.ConsoleLogger{},
			logging.NewFileLogger(
				filepath.Join(appConfig.LogPath, "sessions.log"),
			),
		),
		Config: appConfig.SessionsConfig,
		Users:  map[string]UserInfo{},
	}

	authenticator.Config.SessionDuration = time.Duration(authenticator.Config.Ttl) * time.Minute
	authenticator.Config.DisconnectedUsersToRotateMap = 50

	go authenticator.monitor()
	return authenticator
}

func (ima *InMemoryAuthenticator) IsAuthenticated(userId string) bool {

	ima.Lock.RLock()
	if _, found := ima.Users[userId]; found {
		ima.Lock.RUnlock()
		return true
	} else {
		ima.Lock.RUnlock()
		return false
	}
}

func (ima *InMemoryAuthenticator) Authenticate(id, userName string) {
	ima.Lock.RLock()

	if session, found := ima.Users[id]; found {
		session.RefreshedAt = time.Now().UTC()
		ima.Lock.RUnlock()
	} else {
		ima.Lock.RUnlock()
		ima.Lock.Lock()

		// The reason why we have to double-check is let's say there are two threads
		// that called Authenticated with a user that is not yet know.
		// both threads compete for the lock, let's one thread got the lock and just when it
		// RUnlock()'s the execution is switched to the other thread
		// that other thread can't find the user either, so we end up with two threads
		// in this else statement, once a thread finishes creating a new session
		// the other thread will re-create it if we don't double-check.
		if session, found = ima.Users[id]; found {
			session.RefreshedAt = time.Now().UTC()
		} else {
			now := time.Now().UTC()
			ima.Users[id] = UserInfo{
				Id:              id,
				Username:        userName,
				AuthenticatedAt: now,
				RefreshedAt:     now,
			}
		}
		ima.Lock.Unlock()
	}
}

func (ima *InMemoryAuthenticator) monitor() {
	for {
		time.Sleep(time.Second * 5)
		now := time.Now().UTC()

		ima.Lock.Lock()
		for id, user := range ima.Users {
			if user.RefreshedAt.Sub(now) > ima.Config.SessionDuration {
				ima.Log.Debug(fmt.Sprintf("LoggingOff %s\n", user.Username))
				delete(ima.Users, id)
				ima.DisconnectedUsers++
			}
		}

		if ima.DisconnectedUsers >= ima.Config.DisconnectedUsersToRotateMap {
			newMap := map[string]UserInfo{}
			for id, user := range ima.Users {
				newMap[id] = user
			}
			ima.Users = newMap
		}

		ima.Lock.Unlock()
	}
}
