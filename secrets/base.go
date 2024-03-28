package secrets

import (
	"github.com/bwmarrin/discordgo"
	"github.com/melardev/discord-message-protect/core"
	"time"
)

type RevealRequest struct {
	User      *core.DiscordUser
	Secret    *Secret
	ChannelId string

	// It is imperative to use the session passed along with the interaction
	// we can not use the global bot session, for some reason the global bot session works
	// while i was experimenting in a unit test
	// however in the app it does not, probably because of some delays? anyway, save the session
	// and use it to edit interactions
	Session     *discordgo.Session
	Interaction *discordgo.Interaction

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Secret struct {
	Id        uint   `gorm:"primarykey"`
	SecretId  string `gorm:"uniqueIndex;size:36"`
	Message   string
	ChannelId string
	// used for the database column
	UserName    string
	User        *core.DiscordUser      `gorm:"-"`
	MessageObj  *discordgo.Message     `gorm:"-"`
	Interaction *discordgo.Interaction `gorm:"-"`
	MessageId   string
	CreatedAt   time.Time

	UpdatedAt time.Time
	ImageUrl  string
}

func (s *Secret) TableName() string {
	return "secrets"
}

type CreateSecretDto struct {
	Id        string
	Content   string
	Message   string
	ChannelId string
	User      *discordgo.User
	ImageUrl  string
}

type ISecretManager interface {
	GetById(id string) (*Secret, error)
	Create(dto *CreateSecretDto) (*Secret, error)
	Delete(id string) error
	Update(secret *Secret) error
	UpdateMessageId(secret *Secret) error
}
