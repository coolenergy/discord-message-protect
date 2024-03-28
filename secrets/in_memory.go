package secrets

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/melardev/discord-message-protect/core"
	"sync"
	"time"
)

type InMemorySecretManager struct {
	SecretMutex sync.Mutex
	Secrets     map[string]*Secret
}

func NewInMemorySecretManager(*core.Config) *InMemorySecretManager {
	return &InMemorySecretManager{
		Secrets: map[string]*Secret{},
	}
}

func (s *InMemorySecretManager) GetById(id string) *Secret {
	//TODO implement me
	s.SecretMutex.Lock()
	if secret, found := s.Secrets[id]; found {
		s.SecretMutex.Unlock()
		return secret
	} else {
		s.SecretMutex.Unlock()
		return nil
	}
}

func (s *InMemorySecretManager) Delete(id string) {
	//TODO implement me
}

func (s *InMemorySecretManager) CreateOrUpdate(dto *CreateSecretDto) (*Secret, error) {
	now := time.Now().UTC()
	tries := 0
	var secret *Secret
	s.SecretMutex.Lock()
	for {
		secretId := uuid.Must(uuid.NewRandom()).String()
		if _, found := s.Secrets[secretId]; !found {
			secret = &Secret{
				SecretId: secretId,
				User: &core.DiscordUser{
					Id:       dto.User.Discriminator,
					Username: dto.User.Username,
				},
				Message:   dto.Message,
				ChannelId: dto.ChannelId,
				CreatedAt: now,
				UpdatedAt: now,
			}
			s.Secrets[secretId] = secret
			break
		}

		tries++

		if tries >= 500 {
			message := fmt.Sprintf("Tries 500 times to get unique secret id but failed ... bug??\n")
			// s.Context.DefaultLogger.Error(message)
			return nil, fmt.Errorf(message)
		}
	}

	s.SecretMutex.Unlock()

	return secret, nil
}
